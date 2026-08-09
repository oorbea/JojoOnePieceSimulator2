package config_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
)

// baseStorageEnv sets every var Load requires that isn't itself under test,
// so each test only has to touch the storage-specific ones.
func baseStorageEnv(t *testing.T) {
	t.Helper()
	env := map[string]string{
		"DATABASE_URL":         "postgres://user:pass@localhost:5432/db",
		"GOOGLE_CLIENT_ID":     "client-id",
		"JWT_SECRET":           "01234567890123456789012345678901",
		"R2_ACCOUNT_ID":        "acct",
		"R2_ACCESS_KEY_ID":     "id",
		"R2_SECRET_ACCESS_KEY": "secret",
		"R2_BUCKET":            "bucket",
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoad_StorageProviders_DefaultsToR2Only(t *testing.T) {
	baseStorageEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.StorageProviders) != 1 || cfg.StorageProviders[0] != "r2" {
		t.Errorf("StorageProviders = %v, want [r2]", cfg.StorageProviders)
	}
	if cfg.StorageQuotaThresholdPct != 95 {
		t.Errorf("StorageQuotaThresholdPct = %d, want 95", cfg.StorageQuotaThresholdPct)
	}
	if cfg.R2QuotaBytes != 10*1024*1024*1024 {
		t.Errorf("R2QuotaBytes = %d, want 10 GiB", cfg.R2QuotaBytes)
	}
}

func TestLoad_StorageProviders_UnknownProviderRejected(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_PROVIDERS", "r2,dropbox")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error for unknown provider, got nil")
	}
}

func TestLoad_StorageProviders_DuplicateRejected(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_PROVIDERS", "r2,b2,b2")
	t.Setenv("B2_ENDPOINT", "https://s3.example.backblazeb2.com")
	t.Setenv("B2_REGION", "us-west-004")
	t.Setenv("B2_ACCESS_KEY_ID", "id")
	t.Setenv("B2_SECRET_ACCESS_KEY", "secret")
	t.Setenv("B2_BUCKET", "bucket")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error for duplicate provider, got nil")
	}
}

func TestLoad_StorageProviders_R2MustBeFirst(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_PROVIDERS", "b2,r2")
	t.Setenv("B2_ENDPOINT", "https://s3.example.backblazeb2.com")
	t.Setenv("B2_REGION", "us-west-004")
	t.Setenv("B2_ACCESS_KEY_ID", "id")
	t.Setenv("B2_SECRET_ACCESS_KEY", "secret")
	t.Setenv("B2_BUCKET", "bucket")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error when R2 isn't first, got nil")
	}
}

func TestLoad_StorageProviders_EmptyListRejected(t *testing.T) {
	baseStorageEnv(t)
	// STORAGE_PROVIDERS="," splits to an empty list via splitCSV, distinct
	// from leaving the var unset (which defaults to [r2]).
	t.Setenv("STORAGE_PROVIDERS", ",")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error for an empty provider list, got nil")
	}
}

func TestLoad_StorageQuotaThresholdPct_OutOfRangeRejected(t *testing.T) {
	for _, raw := range []string{"0", "101", "-1"} {
		t.Run(raw, func(t *testing.T) {
			baseStorageEnv(t)
			t.Setenv("STORAGE_QUOTA_THRESHOLD_PCT", raw)

			if _, err := config.Load(); err == nil {
				t.Fatalf("Load: want error for STORAGE_QUOTA_THRESHOLD_PCT=%s, got nil", raw)
			}
		})
	}
}

func TestLoad_StorageQuotaThresholdPct_BoundsAccepted(t *testing.T) {
	for _, raw := range []string{"1", "100"} {
		t.Run(raw, func(t *testing.T) {
			baseStorageEnv(t)
			t.Setenv("STORAGE_QUOTA_THRESHOLD_PCT", raw)

			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := cfg.StorageQuotaThresholdPct; got != mustAtoi(t, raw) {
				t.Errorf("StorageQuotaThresholdPct = %d, want %s", got, raw)
			}
		})
	}
}

func TestLoad_StorageReconcileInterval_NegativeRejected(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_RECONCILE_INTERVAL", "-1h")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error for negative STORAGE_RECONCILE_INTERVAL, got nil")
	}
}

func TestLoad_StorageReconcileInterval_ZeroAccepted(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_RECONCILE_INTERVAL", "0")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.StorageReconcileInterval != 0 {
		t.Errorf("StorageReconcileInterval = %v, want 0", cfg.StorageReconcileInterval)
	}
}

func TestLoad_B2_RequiredVarsEnforcedOnlyWhenListed(t *testing.T) {
	// b2 not listed: none of the B2_* vars are required, even though none
	// are set.
	baseStorageEnv(t)
	if _, err := config.Load(); err != nil {
		t.Fatalf("Load without b2 listed: %v", err)
	}
}

func TestLoad_B2_EachRequiredVarMissingRejected(t *testing.T) {
	full := map[string]string{
		"B2_ENDPOINT":          "https://s3.example.backblazeb2.com",
		"B2_REGION":            "us-west-004",
		"B2_ACCESS_KEY_ID":     "id",
		"B2_SECRET_ACCESS_KEY": "secret",
		"B2_BUCKET":            "bucket",
	}
	for missing := range full {
		t.Run("missing_"+missing, func(t *testing.T) {
			baseStorageEnv(t)
			t.Setenv("STORAGE_PROVIDERS", "r2,b2")
			for k, v := range full {
				if k == missing {
					continue
				}
				t.Setenv(k, v)
			}
			if _, err := config.Load(); err == nil {
				t.Fatalf("Load: want error with %s missing, got nil", missing)
			}
		})
	}
}

func TestLoad_B2_QuotaBytes_NegativeRejected(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_PROVIDERS", "r2,b2")
	t.Setenv("B2_ENDPOINT", "https://s3.example.backblazeb2.com")
	t.Setenv("B2_REGION", "us-west-004")
	t.Setenv("B2_ACCESS_KEY_ID", "id")
	t.Setenv("B2_SECRET_ACCESS_KEY", "secret")
	t.Setenv("B2_BUCKET", "bucket")
	t.Setenv("B2_QUOTA_BYTES", "-1")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error for negative B2_QUOTA_BYTES, got nil")
	}
}

func TestLoad_B2_QuotaBytes_DefaultsToUnlimitedWhenUnset(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_PROVIDERS", "r2,b2")
	t.Setenv("B2_ENDPOINT", "https://s3.example.backblazeb2.com")
	t.Setenv("B2_REGION", "us-west-004")
	t.Setenv("B2_ACCESS_KEY_ID", "id")
	t.Setenv("B2_SECRET_ACCESS_KEY", "secret")
	t.Setenv("B2_BUCKET", "bucket")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.B2QuotaBytes != 0 {
		t.Errorf("B2QuotaBytes = %d, want 0 (unlimited)", cfg.B2QuotaBytes)
	}
}

func TestLoad_Supabase_EachRequiredVarMissingRejected(t *testing.T) {
	full := map[string]string{
		"SUPABASE_S3_ENDPOINT":          "https://ref.storage.supabase.co/storage/v1/s3",
		"SUPABASE_S3_REGION":            "eu-west-1",
		"SUPABASE_S3_ACCESS_KEY_ID":     "id",
		"SUPABASE_S3_SECRET_ACCESS_KEY": "secret",
		"SUPABASE_BUCKET":               "bucket",
	}
	for missing := range full {
		t.Run("missing_"+missing, func(t *testing.T) {
			baseStorageEnv(t)
			t.Setenv("STORAGE_PROVIDERS", "r2,supabase")
			for k, v := range full {
				if k == missing {
					continue
				}
				t.Setenv(k, v)
			}
			if _, err := config.Load(); err == nil {
				t.Fatalf("Load: want error with %s missing, got nil", missing)
			}
		})
	}
}

func TestLoad_FullChain_AllThreeProvidersParsedCorrectly(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("STORAGE_PROVIDERS", "r2,b2,supabase")
	t.Setenv("R2_QUOTA_BYTES", "111")
	t.Setenv("B2_ENDPOINT", "https://s3.eu-central-003.backblazeb2.com")
	t.Setenv("B2_REGION", "eu-central-003")
	t.Setenv("B2_ACCESS_KEY_ID", "b2id")
	t.Setenv("B2_SECRET_ACCESS_KEY", "b2secret")
	t.Setenv("B2_BUCKET", "b2bucket")
	t.Setenv("B2_QUOTA_BYTES", "222")
	t.Setenv("SUPABASE_S3_ENDPOINT", "https://ref.storage.supabase.co/storage/v1/s3")
	t.Setenv("SUPABASE_S3_REGION", "eu-west-1")
	t.Setenv("SUPABASE_S3_ACCESS_KEY_ID", "sbid")
	t.Setenv("SUPABASE_S3_SECRET_ACCESS_KEY", "sbsecret")
	t.Setenv("SUPABASE_BUCKET", "sbbucket")
	t.Setenv("SUPABASE_QUOTA_BYTES", "333")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	wantProviders := []string{"r2", "b2", "supabase"}
	if len(cfg.StorageProviders) != len(wantProviders) {
		t.Fatalf("StorageProviders = %v, want %v", cfg.StorageProviders, wantProviders)
	}
	for i, p := range wantProviders {
		if cfg.StorageProviders[i] != p {
			t.Errorf("StorageProviders[%d] = %q, want %q", i, cfg.StorageProviders[i], p)
		}
	}
	if cfg.R2QuotaBytes != 111 {
		t.Errorf("R2QuotaBytes = %d, want 111", cfg.R2QuotaBytes)
	}
	if cfg.B2Endpoint != "https://s3.eu-central-003.backblazeb2.com" || cfg.B2QuotaBytes != 222 || cfg.B2Bucket != "b2bucket" {
		t.Errorf("B2 config = %+v, want endpoint/bucket/quota as set", cfg)
	}
	if cfg.SupabaseEndpoint != "https://ref.storage.supabase.co/storage/v1/s3" || cfg.SupabaseQuotaBytes != 333 || cfg.SupabaseBucket != "sbbucket" {
		t.Errorf("Supabase config = %+v, want endpoint/bucket/quota as set", cfg)
	}
}

func mustAtoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			t.Fatalf("mustAtoi(%q): not a plain non-negative integer", s)
		}
		n = n*10 + int(c-'0')
	}
	return n
}
