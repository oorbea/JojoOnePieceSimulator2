package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// defaultJWTIssuer is used when JWT_ISSUER is unset.
const defaultJWTIssuer = "jojo-one-piece-simulator"

// defaultJWTTTL is used when JWT_TTL is unset.
const defaultJWTTTL = 24 * time.Hour

// minJWTSecretLen guards against a signing key short enough to brute-force.
const minJWTSecretLen = 32

type Config struct {
	DatabaseURL string
	Port        string

	GoogleClientID string
	JWTSecret      string
	JWTIssuer      string
	JWTTTL         time.Duration
	AdminEmails    []string
}

// Load reads configuration from the environment. If a .env file is present
// in the working directory it is loaded first (without overriding variables
// already set in the environment); its absence is not an error.
func Load() (*Config, error) {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if len(jwtSecret) < minJWTSecretLen {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLen)
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")
	if jwtIssuer == "" {
		jwtIssuer = defaultJWTIssuer
	}

	jwtTTL := defaultJWTTTL
	if raw := os.Getenv("JWT_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing JWT_TTL: %w", err)
		}
		jwtTTL = parsed
	}

	var adminEmails []string
	if raw := os.Getenv("ADMIN_EMAILS"); raw != "" {
		for _, email := range strings.Split(raw, ",") {
			email = strings.ToLower(strings.TrimSpace(email))
			if email == "" {
				continue
			}
			adminEmails = append(adminEmails, email)
		}
	}

	return &Config{
		DatabaseURL:    dsn,
		Port:           port,
		GoogleClientID: googleClientID,
		JWTSecret:      jwtSecret,
		JWTIssuer:      jwtIssuer,
		JWTTTL:         jwtTTL,
		AdminEmails:    adminEmails,
	}, nil
}
