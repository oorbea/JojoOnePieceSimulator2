package config_test

import (
	"testing"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
)

func TestLoad_DevAuthBypass_DefaultsToFalse(t *testing.T) {
	baseStorageEnv(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DevAuthBypass {
		t.Error("DevAuthBypass = true, want false when DEV_AUTH_BYPASS is unset")
	}
}

func TestLoad_DevAuthBypass_LocalEnvAccepted(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("DEV_AUTH_BYPASS", "true")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:8081")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.DevAuthBypass {
		t.Error("DevAuthBypass = false, want true")
	}
}

func TestLoad_DevAuthBypass_LocalEnvAccepted_NoCORSOriginsSet(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("DEV_AUTH_BYPASS", "true")
	t.Setenv("AUTH_COOKIE_SECURE", "false")

	if _, err := config.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestLoad_DevAuthBypass_RejectsSecureCookie(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("DEV_AUTH_BYPASS", "true")
	t.Setenv("AUTH_COOKIE_SECURE", "true")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error when DEV_AUTH_BYPASS=true and AUTH_COOKIE_SECURE=true, got nil")
	}
}

func TestLoad_DevAuthBypass_RejectsSecureCookieDefault(t *testing.T) {
	// AUTH_COOKIE_SECURE left unset - defaults to true, which is also a
	// prod-shaped environment and must be rejected the same way.
	baseStorageEnv(t)
	t.Setenv("DEV_AUTH_BYPASS", "true")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error when DEV_AUTH_BYPASS=true and AUTH_COOKIE_SECURE is unset (defaults true), got nil")
	}
}

func TestLoad_DevAuthBypass_RejectsNonLocalOrigin(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("DEV_AUTH_BYPASS", "true")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://jojo-one-piece-simulator.duckdns.org")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error for DEV_AUTH_BYPASS=true with a non-local CORS origin, got nil")
	}
}

func TestLoad_DevAuthBypass_RejectsOneNonLocalOriginAmongMany(t *testing.T) {
	baseStorageEnv(t)
	t.Setenv("DEV_AUTH_BYPASS", "true")
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,https://jojo-one-piece-simulator.duckdns.org")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load: want error when any CORS origin is non-local, got nil")
	}
}

func TestLoad_DevAuthBypass_FalseIgnoresProdShapedEnv(t *testing.T) {
	// The guard only fires when the bypass itself is on - a normal prod
	// config (bypass unset/false) must load exactly as before.
	baseStorageEnv(t)
	t.Setenv("AUTH_COOKIE_SECURE", "true")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://jojo-one-piece-simulator.duckdns.org")

	if _, err := config.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
}
