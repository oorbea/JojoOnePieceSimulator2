// Package s3store adapts storage.Backend onto any S3-API-compatible object
// store via aws-sdk-go-v2. This is the only package in the codebase allowed
// to import the AWS SDK or know a provider's endpoint shape - Cloudflare R2,
// Backblaze B2, and Supabase Storage all speak the same S3 API, so one
// implementation covers all three; only Config differs per provider.
package s3store

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Config holds everything needed to reach one S3-compatible bucket.
type Config struct {
	// Name identifies this backend for the ledger and config, e.g. "r2",
	// "b2", "supabase". Returned by Backend.Name.
	Name string
	// Endpoint is the provider's S3 API base URL, e.g.
	// "https://<account>.r2.cloudflarestorage.com" for R2.
	Endpoint string
	// Region is the SigV4 region. R2 has no notion of regions and expects
	// "auto"; B2/Supabase expect the bucket's/project's actual region.
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	// PresignTTL bounds how long a URL returned by PresignGet stays valid.
	PresignTTL time.Duration
}

// Backend is the s3store.Config-configured storage.Backend implementation.
type Backend struct {
	name    string
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
	ttl     time.Duration
}

var _ ports.IStorageBackend = (*Backend)(nil)

// New builds a Backend talking to the bucket described by cfg.
func New(ctx context.Context, cfg Config) (*Backend, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("loading aws config for %s: %w", cfg.Name, err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
		// R2 (and, empirically, B2/Supabase's S3-compatible fronts) reject
		// the SDK's default trailer-based CRC checksums; only compute/
		// validate one when a request explicitly requires it.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &Backend{
		name:    cfg.Name,
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
		ttl:     cfg.PresignTTL,
	}, nil
}

// R2Endpoint builds the endpoint for a Cloudflare R2 account. R2's region is
// always "auto". Backblaze B2 and Supabase Storage don't get an equivalent
// helper: their S3-compatible endpoint is shown verbatim in each provider's
// own dashboard, so config takes it as-is instead of reconstructing it from
// an id and risking a wrong guess at their URL shape.
func R2Endpoint(accountID string) string {
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
}

// Name implements storage.Backend.
func (b *Backend) Name() string { return b.name }

// Put implements storage.Backend.
func (b *Backend) Put(ctx context.Context, key string, content io.Reader, contentType string, size int64) error {
	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(b.bucket),
		Key:           aws.String(key),
		Body:          content,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return fmt.Errorf("uploading %q to %s: %w", key, b.name, err)
	}
	return nil
}

// PresignGet implements storage.Backend.
func (b *Backend) PresignGet(ctx context.Context, key string) (string, error) {
	req, err := b.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(b.ttl))
	if err != nil {
		return "", fmt.Errorf("presigning %q on %s: %w", key, b.name, err)
	}
	return req.URL, nil
}

// Del implements storage.Backend.
func (b *Backend) Del(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting %q from %s: %w", key, b.name, err)
	}
	return nil
}

// Walk implements storage.Backend, paging through the whole bucket via
// ListObjectsV2.
func (b *Backend) Walk(ctx context.Context, fn func(key string, bytes int64) error) error {
	paginator := s3.NewListObjectsV2Paginator(b.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(b.bucket),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing objects on %s: %w", b.name, err)
		}
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			size := int64(0)
			if obj.Size != nil {
				size = *obj.Size
			}
			if err := fn(*obj.Key, size); err != nil {
				return err
			}
		}
	}
	return nil
}
