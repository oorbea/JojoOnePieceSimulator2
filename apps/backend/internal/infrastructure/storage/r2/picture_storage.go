// Package r2 adapts ports.IPictureStorage onto Cloudflare R2's S3-compatible
// API via aws-sdk-go-v2. This is the only package in the codebase allowed to
// import the AWS SDK or know R2's endpoint shape - everything else talks to
// ports.IPictureStorage.
package r2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Config holds everything needed to reach one R2 bucket.
type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	// PresignTTL bounds how long a URL returned by PresignGetURL stays valid.
	PresignTTL time.Duration
}

// PictureStorage is the R2-backed ports.IPictureStorage implementation.
type PictureStorage struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
	ttl     time.Duration
}

var _ ports.IPictureStorage = (*PictureStorage)(nil)

// NewPictureStorage builds a PictureStorage talking to the R2 account/bucket
// described by cfg.
func NewPictureStorage(ctx context.Context, cfg Config) (*PictureStorage, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		// R2 has no notion of regions; "auto" is what it expects for SigV4.
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
		o.UsePathStyle = true
		// R2 rejects the SDK's default trailer-based CRC checksums; only
		// compute/validate one when a request explicitly requires it.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &PictureStorage{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
		ttl:     cfg.PresignTTL,
	}, nil
}

// Upload implements ports.IPictureStorage.
func (s *PictureStorage) Upload(ctx context.Context, key string, pic ports.Picture) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          pic.Content,
		ContentType:   aws.String(pic.ContentType),
		ContentLength: aws.Int64(pic.Size),
	})
	if err != nil {
		return fmt.Errorf("uploading picture %q: %w", key, err)
	}
	return nil
}

// PresignGetURL implements ports.IPictureStorage.
func (s *PictureStorage) PresignGetURL(ctx context.Context, key string) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.ttl))
	if err != nil {
		return "", fmt.Errorf("presigning picture %q: %w", key, err)
	}
	return req.URL, nil
}

// Delete implements ports.IPictureStorage.
func (s *PictureStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting picture %q: %w", key, err)
	}
	return nil
}
