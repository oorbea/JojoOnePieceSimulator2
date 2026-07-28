package config

import (
	"fmt"
	"os"
	"strconv"
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

// defaultCORSAllowedMethods/Headers/MaxAge are used when their respective
// env vars are unset, but only take effect once CORS_ALLOWED_ORIGINS is
// non-empty - see Load.
var defaultCORSAllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
var defaultCORSAllowedHeaders = []string{"Content-Type", "Authorization"}

const defaultCORSMaxAge = 300

// defaultRateLimit* are used when their respective env vars are unset.
const defaultRateLimitEnabled = true
const defaultRateLimitWindow = time.Minute
const defaultRateLimitGlobalPerIP = 120
const defaultRateLimitLoginPerIP = 10
const defaultRateLimitReadPerUser = 100
const defaultRateLimitWritePerUser = 30

// defaultR2PresignTTL is used when R2_PRESIGN_TTL is unset.
const defaultR2PresignTTL = 15 * time.Minute

// defaultPictureMaxBytes is used when PICTURE_MAX_BYTES is unset.
const defaultPictureMaxBytes = 5 * 1024 * 1024

// defaultPictureAllowedTypes is used when PICTURE_ALLOWED_TYPES is unset.
var defaultPictureAllowedTypes = []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}

type Config struct {
	DatabaseURL string
	Port        string

	GoogleClientID string
	JWTSecret      string
	JWTIssuer      string
	JWTTTL         time.Duration
	AdminEmails    []string

	// CORSAllowedOrigins is deny-all (no CORS headers added at all) when
	// empty, which is the safe default: the browser blocks cross-origin
	// calls exactly as if the server didn't know about CORS.
	CORSAllowedOrigins   []string
	CORSAllowedMethods   []string
	CORSAllowedHeaders   []string
	CORSAllowCredentials bool
	CORSMaxAge           int

	// RateLimitEnabled turns the whole tiered limiter off when false (all
	// other RateLimit* fields are then ignored).
	RateLimitEnabled      bool
	RateLimitWindow       time.Duration
	RateLimitGlobalPerIP  int
	RateLimitLoginPerIP   int
	RateLimitReadPerUser  int
	RateLimitWritePerUser int

	// R2* configure the Cloudflare R2 bucket Stand pictures are stored in.
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2PresignTTL      time.Duration

	// PictureMaxBytes/PictureAllowedTypes bound what PATCH
	// /stands/{id}/picture accepts.
	PictureMaxBytes     int64
	PictureAllowedTypes []string
}

// splitCSV splits raw on commas, trimming whitespace and dropping empty
// entries. Returns nil (not an empty slice) if nothing is left, so callers
// can treat "unset" and "explicitly empty" the same way.
func splitCSV(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

// parsePositiveIntEnv parses name as an int, defaulting to def when unset.
// Rejects a negative value outright, since a typo'd negative limit would
// otherwise silently disable a rate-limit tier instead of failing loudly.
func parsePositiveIntEnv(name string, def int) (int, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return def, nil
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", name, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("%s must not be negative", name)
	}
	return parsed, nil
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

	// CORS is deny-all unless CORS_ALLOWED_ORIGINS is set: the other
	// CORS_* vars are meaningless (and left unparsed) with no origins to
	// allow, so their defaults only ever apply alongside a configured
	// origin list.
	corsAllowedOrigins := splitCSV(os.Getenv("CORS_ALLOWED_ORIGINS"))

	corsAllowedMethods := defaultCORSAllowedMethods
	if raw := os.Getenv("CORS_ALLOWED_METHODS"); raw != "" {
		corsAllowedMethods = splitCSV(raw)
	}

	corsAllowedHeaders := defaultCORSAllowedHeaders
	if raw := os.Getenv("CORS_ALLOWED_HEADERS"); raw != "" {
		corsAllowedHeaders = splitCSV(raw)
	}

	corsAllowCredentials := false
	if raw := os.Getenv("CORS_ALLOW_CREDENTIALS"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CORS_ALLOW_CREDENTIALS: %w", err)
		}
		corsAllowCredentials = parsed
	}

	corsMaxAge := defaultCORSMaxAge
	if raw := os.Getenv("CORS_MAX_AGE"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CORS_MAX_AGE: %w", err)
		}
		corsMaxAge = parsed
	}

	rateLimitEnabled := defaultRateLimitEnabled
	if raw := os.Getenv("RATE_LIMIT_ENABLED"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing RATE_LIMIT_ENABLED: %w", err)
		}
		rateLimitEnabled = parsed
	}

	rateLimitWindow := defaultRateLimitWindow
	if raw := os.Getenv("RATE_LIMIT_WINDOW"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing RATE_LIMIT_WINDOW: %w", err)
		}
		rateLimitWindow = parsed
	}

	rateLimitGlobalPerIP, err := parsePositiveIntEnv("RATE_LIMIT_GLOBAL_PER_IP", defaultRateLimitGlobalPerIP)
	if err != nil {
		return nil, err
	}

	rateLimitLoginPerIP, err := parsePositiveIntEnv("RATE_LIMIT_LOGIN_PER_IP", defaultRateLimitLoginPerIP)
	if err != nil {
		return nil, err
	}

	rateLimitReadPerUser, err := parsePositiveIntEnv("RATE_LIMIT_READ_PER_USER", defaultRateLimitReadPerUser)
	if err != nil {
		return nil, err
	}

	rateLimitWritePerUser, err := parsePositiveIntEnv("RATE_LIMIT_WRITE_PER_USER", defaultRateLimitWritePerUser)
	if err != nil {
		return nil, err
	}

	r2AccountID := os.Getenv("R2_ACCOUNT_ID")
	if r2AccountID == "" {
		return nil, fmt.Errorf("R2_ACCOUNT_ID is required")
	}

	r2AccessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	if r2AccessKeyID == "" {
		return nil, fmt.Errorf("R2_ACCESS_KEY_ID is required")
	}

	r2SecretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	if r2SecretAccessKey == "" {
		return nil, fmt.Errorf("R2_SECRET_ACCESS_KEY is required")
	}

	r2Bucket := os.Getenv("R2_BUCKET")
	if r2Bucket == "" {
		return nil, fmt.Errorf("R2_BUCKET is required")
	}

	r2PresignTTL := defaultR2PresignTTL
	if raw := os.Getenv("R2_PRESIGN_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing R2_PRESIGN_TTL: %w", err)
		}
		r2PresignTTL = parsed
	}

	pictureMaxBytes := int64(defaultPictureMaxBytes)
	if raw := os.Getenv("PICTURE_MAX_BYTES"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing PICTURE_MAX_BYTES: %w", err)
		}
		if parsed < 0 {
			return nil, fmt.Errorf("PICTURE_MAX_BYTES must not be negative")
		}
		pictureMaxBytes = parsed
	}

	pictureAllowedTypes := defaultPictureAllowedTypes
	if raw := os.Getenv("PICTURE_ALLOWED_TYPES"); raw != "" {
		pictureAllowedTypes = splitCSV(raw)
	}

	return &Config{
		DatabaseURL:          dsn,
		Port:                 port,
		GoogleClientID:       googleClientID,
		JWTSecret:            jwtSecret,
		JWTIssuer:            jwtIssuer,
		JWTTTL:               jwtTTL,
		AdminEmails:          adminEmails,
		CORSAllowedOrigins:   corsAllowedOrigins,
		CORSAllowedMethods:   corsAllowedMethods,
		CORSAllowedHeaders:   corsAllowedHeaders,
		CORSAllowCredentials: corsAllowCredentials,
		CORSMaxAge:           corsMaxAge,

		RateLimitEnabled:      rateLimitEnabled,
		RateLimitWindow:       rateLimitWindow,
		RateLimitGlobalPerIP:  rateLimitGlobalPerIP,
		RateLimitLoginPerIP:   rateLimitLoginPerIP,
		RateLimitReadPerUser:  rateLimitReadPerUser,
		RateLimitWritePerUser: rateLimitWritePerUser,

		R2AccountID:       r2AccountID,
		R2AccessKeyID:     r2AccessKeyID,
		R2SecretAccessKey: r2SecretAccessKey,
		R2Bucket:          r2Bucket,
		R2PresignTTL:      r2PresignTTL,

		PictureMaxBytes:     pictureMaxBytes,
		PictureAllowedTypes: pictureAllowedTypes,
	}, nil
}
