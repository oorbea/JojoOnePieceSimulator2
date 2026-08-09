package s3store_test

import (
	"context"
	"testing"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/s3store"
)

func TestR2Endpoint(t *testing.T) {
	got := s3store.R2Endpoint("abc123")
	want := "https://abc123.r2.cloudflarestorage.com"
	if got != want {
		t.Errorf("R2Endpoint(%q) = %q, want %q", "abc123", got, want)
	}
}

// TestNew_BuildsBackendForEachProviderShape exercises New with the endpoint
// shape each configured provider actually uses (R2's own helper, and the
// literal endpoints B2/Supabase display in their dashboards). No network
// call happens here - aws-sdk-go-v2's LoadDefaultConfig with static
// credentials doesn't touch the network, so this only checks that Config is
// wired into a working *Backend, not that the bucket is reachable.
func TestNew_BuildsBackendForEachProviderShape(t *testing.T) {
	cases := []struct {
		name string
		cfg  s3store.Config
	}{
		{"r2", s3store.Config{
			Name: "r2", Endpoint: s3store.R2Endpoint("acct"), Region: "auto",
			AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "bucket", PresignTTL: 15 * time.Minute,
		}},
		{"b2", s3store.Config{
			Name: "b2", Endpoint: "https://s3.us-west-004.backblazeb2.com", Region: "us-west-004",
			AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "bucket", PresignTTL: 15 * time.Minute,
		}},
		{"supabase", s3store.Config{
			Name: "supabase", Endpoint: "https://xyzcompany.supabase.co/storage/v1/s3", Region: "eu-west-1",
			AccessKeyID: "id", SecretAccessKey: "secret", Bucket: "bucket", PresignTTL: 15 * time.Minute,
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			backend, err := s3store.New(context.Background(), tc.cfg)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			if backend.Name() != tc.cfg.Name {
				t.Errorf("Name() = %q, want %q", backend.Name(), tc.cfg.Name)
			}
			var _ ports.IStorageBackend = backend
		})
	}
}
