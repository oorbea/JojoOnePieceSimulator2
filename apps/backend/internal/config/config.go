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

// defaultR2QuotaBytes is R2's free-tier storage cap (10 GiB), used when
// R2_QUOTA_BYTES is unset.
const defaultR2QuotaBytes = 10 * 1024 * 1024 * 1024

// defaultStorageProviders is the fallback chain order used when
// STORAGE_PROVIDERS is unset: R2 only, i.e. today's behavior plus the
// storage ledger.
var defaultStorageProviders = []string{"r2"}

// defaultStorageQuotaThresholdPct is used when STORAGE_QUOTA_THRESHOLD_PCT
// is unset - an upload is only allowed onto a tier if it would stay under
// this percentage of that tier's quota, leaving headroom for ledger drift
// and for one oversized upload landing right at the edge.
const defaultStorageQuotaThresholdPct = 95

// defaultStorageReconcileInterval is used when STORAGE_RECONCILE_INTERVAL is
// unset. 0 disables reconciliation.
const defaultStorageReconcileInterval = 6 * time.Hour

// knownStorageProviders is every provider name the fallback chain knows how
// to build a backend for.
var knownStorageProviders = map[string]bool{"r2": true, "b2": true, "supabase": true}

// defaultPictureMaxBytes is used when PICTURE_MAX_BYTES is unset.
const defaultPictureMaxBytes = 5 * 1024 * 1024

// defaultPictureAllowedTypes is used when PICTURE_ALLOWED_TYPES is unset.
var defaultPictureAllowedTypes = []string{"image/webp", "image/avif", "image/jpeg", "image/png", "image/gif"}

// defaultPicture* configure the background image-compression pipeline: every
// accepted upload is re-encoded to WebP, resized to fit within
// PictureMaxDimension, with a PictureThumbDimension-capped thumbnail
// alongside it.
const defaultPictureMaxDimension = 1024
const defaultPictureThumbDimension = 256
const defaultPictureWebPQuality = 80
const defaultPictureMaxPixels = int64(50_000_000)
const defaultPictureWorkers = 2
const defaultPictureQueueSize = 32
const defaultPictureJobTimeout = 30 * time.Second

// defaultCache*/defaultRedis* configure the read cache in front of the Stand
// repository and picture presign URLs. Caching is entirely off when
// REDIS_URL is unset, regardless of CACHE_ENABLED - see Load.
const defaultCacheEnabled = true
const defaultCacheStandTTL = 5 * time.Minute
const defaultCacheNotFoundTTL = 30 * time.Second

// 0 disables Cache-Control's max-age (cacheHeaders then sends
// "private, no-cache") - ETag/304 revalidation stays on regardless, but a
// shared-proxy-free client can no longer serve a stale list body straight
// from its own HTTP cache for up to HTTPMaxAge after a write, which briefly
// hid freshly created Stands/DevilFruits from their own creator.
const defaultCacheHTTPMaxAge = 0 * time.Second
const defaultRedisDialTimeout = 2 * time.Second
const defaultRedisOpTimeout = 200 * time.Millisecond

type Config struct {
	DatabaseURL    string
	Port           string
	GoogleClientID string
	JWTSecret      string
	JWTIssuer      string
	// R2* configure the Cloudflare R2 bucket Stand pictures are stored in.
	// R2 is always the first tier of the storage fallback chain (see
	// Storage* below) and, unlike B2/Supabase, always required.
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	// R2QuotaBytes/B2QuotaBytes/SupabaseQuotaBytes are each provider's
	// free-tier storage cap; 0 means unlimited (never falls through on
	// quota, only on a runtime error).
	R2QuotaBytes int64
	// StorageProviders is the fallback chain order, e.g. []string{"r2",
	// "b2", "supabase"}. A provider only needs its credentials/bucket/quota
	// set if it's listed here.
	StorageProviders []string
	// StorageQuotaThresholdPct bounds how full (as a percentage of its
	// quota) a tier may get before new uploads skip it in favor of the next
	// tier.
	StorageQuotaThresholdPct int
	// StorageReconcileInterval is how often the storage reconciler re-walks
	// every configured bucket to correct ledger drift. 0 disables it.
	StorageReconcileInterval time.Duration
	// B2* configure the optional Backblaze B2 tier.
	B2Endpoint        string
	B2Region          string
	B2AccessKeyID     string
	B2SecretAccessKey string
	B2Bucket          string
	B2QuotaBytes      int64
	// Supabase* configure the optional Supabase Storage tier.
	SupabaseEndpoint        string
	SupabaseRegion          string
	SupabaseAccessKeyID     string
	SupabaseSecretAccessKey string
	SupabaseBucket          string
	SupabaseQuotaBytes      int64
	// RedisURL is empty by default, which turns caching off entirely (no
	// connection is ever attempted) - keeps `go run`/`make test` working
	// with no Redis around. CacheEnabled is a separate kill switch on top of
	// that, for disabling the cache without unsetting REDIS_URL.
	RedisURL    string
	AdminEmails []string
	// CORSAllowedOrigins is deny-all (no CORS headers added at all) when
	// empty, which is the safe default: the browser blocks cross-origin
	// calls exactly as if the server didn't know about CORS.
	CORSAllowedOrigins    []string
	CORSAllowedMethods    []string
	CORSAllowedHeaders    []string
	PictureAllowedTypes   []string
	JWTTTL                time.Duration
	CORSMaxAge            int
	RateLimitWindow       time.Duration
	RateLimitGlobalPerIP  int
	RateLimitLoginPerIP   int
	RateLimitReadPerUser  int
	RateLimitWritePerUser int
	R2PresignTTL          time.Duration
	// PictureMaxBytes/PictureAllowedTypes bound what PATCH
	// /stands/{id}/picture accepts.
	PictureMaxBytes int64
	// Picture* below configure the background compression pipeline: the
	// resize caps and WebP quality applied by the image processor, and the
	// worker pool that runs it.
	PictureMaxDimension   int
	PictureThumbDimension int
	PictureWebPQuality    int
	PictureMaxPixels      int64
	PictureWorkers        int
	PictureQueueSize      int
	PictureJobTimeout     time.Duration
	RedisDialTimeout      time.Duration
	RedisOpTimeout        time.Duration
	// CacheStandTTL/CacheDevilFruitTTL/CacheNotFoundTTL bound how long the
	// Stand/DevilFruit repository caches can serve stale data if an
	// invalidation is ever missed. CachePresignTTL does the same for cached
	// presigned picture URLs, and is validated to stay safely under
	// R2PresignTTL so a served URL is never close to expiring.
	// CacheHTTPMaxAge configures the response Cache-Control header
	// (independent of Redis); 0 disables it.
	CacheStandTTL        time.Duration
	CacheDevilFruitTTL   time.Duration
	CacheNotFoundTTL     time.Duration
	CachePresignTTL      time.Duration
	CacheHTTPMaxAge      time.Duration
	CORSAllowCredentials bool
	// RateLimitEnabled turns the whole tiered limiter off when false (all
	// other RateLimit* fields are then ignored).
	RateLimitEnabled bool
	CacheEnabled     bool
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

// parseQuotaBytesEnv parses name as an int64 byte count, defaulting to def
// when unset. A negative value is rejected; 0 is allowed and means
// "unlimited" for storage quotas.
func parseQuotaBytesEnv(name string, def int64) (int64, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return def, nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
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

	r2QuotaBytes, err := parseQuotaBytesEnv("R2_QUOTA_BYTES", defaultR2QuotaBytes)
	if err != nil {
		return nil, err
	}

	storageProviders := defaultStorageProviders
	if raw := os.Getenv("STORAGE_PROVIDERS"); raw != "" {
		storageProviders = splitCSV(raw)
	}
	if len(storageProviders) == 0 {
		return nil, fmt.Errorf("STORAGE_PROVIDERS must list at least one provider")
	}
	seenProvider := make(map[string]bool, len(storageProviders))
	for _, p := range storageProviders {
		if !knownStorageProviders[p] {
			return nil, fmt.Errorf("STORAGE_PROVIDERS: unknown provider %q", p)
		}
		if seenProvider[p] {
			return nil, fmt.Errorf("STORAGE_PROVIDERS: %q listed more than once", p)
		}
		seenProvider[p] = true
	}
	if storageProviders[0] != "r2" {
		return nil, fmt.Errorf("STORAGE_PROVIDERS: R2 must be the first tier")
	}

	storageQuotaThresholdPct, err := parsePositiveIntEnv("STORAGE_QUOTA_THRESHOLD_PCT", defaultStorageQuotaThresholdPct)
	if err != nil {
		return nil, err
	}
	if storageQuotaThresholdPct < 1 || storageQuotaThresholdPct > 100 {
		return nil, fmt.Errorf("STORAGE_QUOTA_THRESHOLD_PCT must be between 1 and 100")
	}

	storageReconcileInterval := defaultStorageReconcileInterval
	if raw := os.Getenv("STORAGE_RECONCILE_INTERVAL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing STORAGE_RECONCILE_INTERVAL: %w", err)
		}
		if parsed < 0 {
			return nil, fmt.Errorf("STORAGE_RECONCILE_INTERVAL must not be negative")
		}
		storageReconcileInterval = parsed
	}

	var b2Endpoint, b2Region, b2AccessKeyID, b2SecretAccessKey, b2Bucket string
	var b2QuotaBytes int64
	if seenProvider["b2"] {
		b2Endpoint = os.Getenv("B2_ENDPOINT")
		b2Region = os.Getenv("B2_REGION")
		b2AccessKeyID = os.Getenv("B2_ACCESS_KEY_ID")
		b2SecretAccessKey = os.Getenv("B2_SECRET_ACCESS_KEY")
		b2Bucket = os.Getenv("B2_BUCKET")
		for name, val := range map[string]string{
			"B2_ENDPOINT": b2Endpoint, "B2_REGION": b2Region, "B2_ACCESS_KEY_ID": b2AccessKeyID,
			"B2_SECRET_ACCESS_KEY": b2SecretAccessKey, "B2_BUCKET": b2Bucket,
		} {
			if val == "" {
				return nil, fmt.Errorf("%s is required when STORAGE_PROVIDERS includes \"b2\"", name)
			}
		}
		b2QuotaBytes, err = parseQuotaBytesEnv("B2_QUOTA_BYTES", 0)
		if err != nil {
			return nil, err
		}
	}

	var supabaseEndpoint, supabaseRegion, supabaseAccessKeyID, supabaseSecretAccessKey, supabaseBucket string
	var supabaseQuotaBytes int64
	if seenProvider["supabase"] {
		supabaseEndpoint = os.Getenv("SUPABASE_S3_ENDPOINT")
		supabaseRegion = os.Getenv("SUPABASE_S3_REGION")
		supabaseAccessKeyID = os.Getenv("SUPABASE_S3_ACCESS_KEY_ID")
		supabaseSecretAccessKey = os.Getenv("SUPABASE_S3_SECRET_ACCESS_KEY")
		supabaseBucket = os.Getenv("SUPABASE_BUCKET")
		for name, val := range map[string]string{
			"SUPABASE_S3_ENDPOINT": supabaseEndpoint, "SUPABASE_S3_REGION": supabaseRegion,
			"SUPABASE_S3_ACCESS_KEY_ID": supabaseAccessKeyID, "SUPABASE_S3_SECRET_ACCESS_KEY": supabaseSecretAccessKey,
			"SUPABASE_BUCKET": supabaseBucket,
		} {
			if val == "" {
				return nil, fmt.Errorf("%s is required when STORAGE_PROVIDERS includes \"supabase\"", name)
			}
		}
		supabaseQuotaBytes, err = parseQuotaBytesEnv("SUPABASE_QUOTA_BYTES", 0)
		if err != nil {
			return nil, err
		}
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

	pictureMaxDimension, err := parsePositiveIntEnv("PICTURE_MAX_DIMENSION", defaultPictureMaxDimension)
	if err != nil {
		return nil, err
	}

	pictureThumbDimension, err := parsePositiveIntEnv("PICTURE_THUMB_DIMENSION", defaultPictureThumbDimension)
	if err != nil {
		return nil, err
	}

	pictureWebPQuality, err := parsePositiveIntEnv("PICTURE_WEBP_QUALITY", defaultPictureWebPQuality)
	if err != nil {
		return nil, err
	}
	if pictureWebPQuality < 1 || pictureWebPQuality > 100 {
		return nil, fmt.Errorf("PICTURE_WEBP_QUALITY must be between 1 and 100")
	}

	pictureMaxPixels := defaultPictureMaxPixels
	if raw := os.Getenv("PICTURE_MAX_PIXELS"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing PICTURE_MAX_PIXELS: %w", err)
		}
		if parsed < 0 {
			return nil, fmt.Errorf("PICTURE_MAX_PIXELS must not be negative")
		}
		pictureMaxPixels = parsed
	}

	pictureWorkers, err := parsePositiveIntEnv("PICTURE_WORKERS", defaultPictureWorkers)
	if err != nil {
		return nil, err
	}
	if pictureWorkers < 1 {
		return nil, fmt.Errorf("PICTURE_WORKERS must be at least 1")
	}

	pictureQueueSize, err := parsePositiveIntEnv("PICTURE_QUEUE_SIZE", defaultPictureQueueSize)
	if err != nil {
		return nil, err
	}
	if pictureQueueSize < 1 {
		return nil, fmt.Errorf("PICTURE_QUEUE_SIZE must be at least 1")
	}

	pictureJobTimeout := defaultPictureJobTimeout
	if raw := os.Getenv("PICTURE_JOB_TIMEOUT"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing PICTURE_JOB_TIMEOUT: %w", err)
		}
		pictureJobTimeout = parsed
	}

	redisURL := os.Getenv("REDIS_URL")

	cacheEnabled := defaultCacheEnabled
	if raw := os.Getenv("CACHE_ENABLED"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CACHE_ENABLED: %w", err)
		}
		cacheEnabled = parsed
	}

	redisDialTimeout := defaultRedisDialTimeout
	if raw := os.Getenv("REDIS_DIAL_TIMEOUT"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing REDIS_DIAL_TIMEOUT: %w", err)
		}
		redisDialTimeout = parsed
	}

	redisOpTimeout := defaultRedisOpTimeout
	if raw := os.Getenv("REDIS_OP_TIMEOUT"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing REDIS_OP_TIMEOUT: %w", err)
		}
		redisOpTimeout = parsed
	}

	cacheStandTTL := defaultCacheStandTTL
	if raw := os.Getenv("CACHE_STAND_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CACHE_STAND_TTL: %w", err)
		}
		cacheStandTTL = parsed
	}

	cacheDevilFruitTTL := defaultCacheStandTTL
	if raw := os.Getenv("CACHE_DEVIL_FRUIT_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CACHE_DEVIL_FRUIT_TTL: %w", err)
		}
		cacheDevilFruitTTL = parsed
	}

	cacheNotFoundTTL := defaultCacheNotFoundTTL
	if raw := os.Getenv("CACHE_NOTFOUND_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CACHE_NOTFOUND_TTL: %w", err)
		}
		cacheNotFoundTTL = parsed
	}

	// CachePresignTTL defaults to half of R2PresignTTL so a served URL
	// always has at least half its validity left; an explicit value must
	// stay strictly below R2PresignTTL for the same reason.
	cachePresignTTL := r2PresignTTL / 2
	if raw := os.Getenv("CACHE_PRESIGN_TTL"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CACHE_PRESIGN_TTL: %w", err)
		}
		if parsed >= r2PresignTTL {
			return nil, fmt.Errorf("CACHE_PRESIGN_TTL must be less than R2_PRESIGN_TTL")
		}
		cachePresignTTL = parsed
	}

	cacheHTTPMaxAge := defaultCacheHTTPMaxAge
	if raw := os.Getenv("CACHE_HTTP_MAX_AGE"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parsing CACHE_HTTP_MAX_AGE: %w", err)
		}
		if parsed < 0 {
			return nil, fmt.Errorf("CACHE_HTTP_MAX_AGE must not be negative")
		}
		cacheHTTPMaxAge = parsed
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
		R2QuotaBytes:      r2QuotaBytes,

		StorageProviders:         storageProviders,
		StorageQuotaThresholdPct: storageQuotaThresholdPct,
		StorageReconcileInterval: storageReconcileInterval,

		B2Endpoint:        b2Endpoint,
		B2Region:          b2Region,
		B2AccessKeyID:     b2AccessKeyID,
		B2SecretAccessKey: b2SecretAccessKey,
		B2Bucket:          b2Bucket,
		B2QuotaBytes:      b2QuotaBytes,

		SupabaseEndpoint:        supabaseEndpoint,
		SupabaseRegion:          supabaseRegion,
		SupabaseAccessKeyID:     supabaseAccessKeyID,
		SupabaseSecretAccessKey: supabaseSecretAccessKey,
		SupabaseBucket:          supabaseBucket,
		SupabaseQuotaBytes:      supabaseQuotaBytes,

		PictureMaxBytes:     pictureMaxBytes,
		PictureAllowedTypes: pictureAllowedTypes,

		PictureMaxDimension:   pictureMaxDimension,
		PictureThumbDimension: pictureThumbDimension,
		PictureWebPQuality:    pictureWebPQuality,
		PictureMaxPixels:      pictureMaxPixels,
		PictureWorkers:        pictureWorkers,
		PictureQueueSize:      pictureQueueSize,
		PictureJobTimeout:     pictureJobTimeout,

		RedisURL:           redisURL,
		RedisDialTimeout:   redisDialTimeout,
		RedisOpTimeout:     redisOpTimeout,
		CacheEnabled:       cacheEnabled,
		CacheStandTTL:      cacheStandTTL,
		CacheDevilFruitTTL: cacheDevilFruitTTL,
		CacheNotFoundTTL:   cacheNotFoundTTL,
		CachePresignTTL:    cachePresignTTL,
		CacheHTTPMaxAge:    cacheHTTPMaxAge,
	}, nil
}
