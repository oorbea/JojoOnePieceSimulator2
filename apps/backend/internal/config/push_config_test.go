package config_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
)

func TestLoad_VAPID_UnsetDisablesPush(t *testing.T) {
	baseStorageEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.VAPIDPublicKey != "" || cfg.VAPIDPrivateKey != "" || cfg.VAPIDSubject != "" {
		t.Errorf("VAPID = %q/%q/%q, want all empty when unset", cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject)
	}
}

func TestLoad_VAPID_FullyConfigured(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("VAPID_PUBLIC_KEY", "public-key")
	t.Setenv("VAPID_PRIVATE_KEY", "private-key")
	t.Setenv("VAPID_SUBJECT", "mailto:ops@example.com")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.VAPIDPublicKey != "public-key" || cfg.VAPIDPrivateKey != "private-key" || cfg.VAPIDSubject != "mailto:ops@example.com" {
		t.Errorf("VAPID = %q/%q/%q, want the configured values", cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubject)
	}
}

func TestLoad_VAPID_HttpsSubjectAccepted(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("VAPID_PUBLIC_KEY", "public-key")
	t.Setenv("VAPID_PRIVATE_KEY", "private-key")
	t.Setenv("VAPID_SUBJECT", "https://jojo-one-piece-simulator.duckdns.org")

	if _, err := config.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestLoad_VAPID_PartialConfigIsRejected(t *testing.T) {
	cases := map[string]map[string]string{
		"only public":  {"VAPID_PUBLIC_KEY": "p"},
		"only private": {"VAPID_PRIVATE_KEY": "p"},
		"only subject": {"VAPID_SUBJECT": "mailto:ops@example.com"},
		"no subject":   {"VAPID_PUBLIC_KEY": "p", "VAPID_PRIVATE_KEY": "k"},
		"no private":   {"VAPID_PUBLIC_KEY": "p", "VAPID_SUBJECT": "mailto:ops@example.com"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			baseStorageEnv(t)
			for key, value := range env {
				t.Setenv(key, value)
			}
			if _, err := config.Load(); err == nil {
				t.Fatal("Load: want error for a partial VAPID configuration, got nil")
			}
		})
	}
}

func TestLoad_VAPID_BadSubjectIsRejected(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("VAPID_PUBLIC_KEY", "public-key")
	t.Setenv("VAPID_PRIVATE_KEY", "private-key")
	t.Setenv("VAPID_SUBJECT", "ops@example.com")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error when VAPID_SUBJECT is neither mailto: nor https://, got nil")
	}
}
