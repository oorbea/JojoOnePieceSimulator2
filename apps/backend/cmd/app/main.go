package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/auth"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache"
	rediscache "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache/redis"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/imaging"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/r2"
)

// @title JojoOnePieceSimulator2 API
// @version 1.0
// @description Backend API for JojoOnePieceSimulator2 - Google-only auth, Stands and Devil Fruits catalogues.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT access token.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	pictureStorage, err := r2.NewPictureStorage(ctx, r2.Config{
		AccountID:       cfg.R2AccountID,
		AccessKeyID:     cfg.R2AccessKeyID,
		SecretAccessKey: cfg.R2SecretAccessKey,
		Bucket:          cfg.R2Bucket,
		PresignTTL:      cfg.R2PresignTTL,
	})
	if err != nil {
		log.Fatalf("configuring picture storage: %v", err)
	}

	imageProcessor, closeImaging, err := imaging.New(imaging.Config{Concurrency: 1})
	if err != nil {
		log.Fatalf("configuring image processor: %v", err)
	}

	standRepository := repositories.NewStandRepository(pool)
	devilFruitRepository := repositories.NewDevilFruitRepository(pool)

	// standRepo/devilFruitRepo/pictures start as the undecorated adapters;
	// all three get wrapped with a Redis-backed cache below when one is
	// configured. The picture worker and each catalogue's service must
	// receive the *same* decorated repo, or a background transcode
	// publishing READY/FAILED would invalidate one copy of the cache while
	// readers kept hitting the other.
	var standRepo ports.IStandRepository = standRepository
	var devilFruitRepo ports.IDevilFruitRepository = devilFruitRepository
	var pictures ports.IPictureStorage = pictureStorage

	if cfg.CacheEnabled && cfg.RedisURL != "" {
		redisCache, err := rediscache.New(ctx, rediscache.Config{
			URL:         cfg.RedisURL,
			DialTimeout: cfg.RedisDialTimeout,
			OpTimeout:   cfg.RedisOpTimeout,
		})
		if err != nil {
			log.Fatalf("connecting to redis: %v", err)
		}
		defer func() {
			if err := redisCache.Close(); err != nil {
				log.Printf("closing redis connection: %v", err)
			}
		}()

		standRepo = cache.NewStandRepository(standRepo, redisCache, cfg.CacheStandTTL, cfg.CacheNotFoundTTL)
		devilFruitRepo = cache.NewDevilFruitRepository(devilFruitRepo, redisCache, cfg.CacheDevilFruitTTL, cfg.CacheNotFoundTTL)
		pictures = cache.NewPictureStorage(pictures, redisCache, cfg.CachePresignTTL)
	}

	userRepository := repositories.NewUserRepository(pool)
	var userRepo ports.IUserRepository = userRepository

	pictureTargets := map[enums.PictureSubjectKind]services.PictureTarget{
		enums.StandSubject:      {Publisher: services.NewStandPicturePublisher(standRepo), KeyPrefix: "stands"},
		enums.DevilFruitSubject: {Publisher: services.NewDevilFruitPicturePublisher(devilFruitRepo), KeyPrefix: "devil-fruits"},
		enums.UserSubject:       {Publisher: services.NewUserPicturePublisher(userRepo), KeyPrefix: "users"},
	}

	pictureWorker := services.NewPictureWorker(imageProcessor, pictures, pictureTargets,
		idgen.UUIDGenerator[powers.PowerID]{}, services.WorkerConfig{
			Workers:        cfg.PictureWorkers,
			QueueSize:      cfg.PictureQueueSize,
			JobTimeout:     cfg.PictureJobTimeout,
			MaxDimension:   cfg.PictureMaxDimension,
			ThumbDimension: cfg.PictureThumbDimension,
			Quality:        cfg.PictureWebPQuality,
		})
	pictureWorker.Start()

	picturePolicy := services.PicturePolicy{
		MaxBytes:     cfg.PictureMaxBytes,
		AllowedTypes: cfg.PictureAllowedTypes,
		MaxPixels:    cfg.PictureMaxPixels,
	}

	standService := services.NewStandService(standRepo, idgen.UUIDGenerator[powers.PowerID]{}, pictures,
		imageProcessor, pictureWorker, picturePolicy)
	standEndpoints := endpoints.NewStandEndpoints(standService)

	devilFruitService := services.NewDevilFruitService(devilFruitRepo, idgen.UUIDGenerator[powers.PowerID]{}, pictures,
		imageProcessor, pictureWorker, picturePolicy)
	devilFruitEndpoints := endpoints.NewDevilFruitEndpoints(devilFruitService)

	googleVerifier := auth.NewGoogleVerifier(cfg.GoogleClientID)
	tokenIssuer := auth.NewJWTIssuer([]byte(cfg.JWTSecret), cfg.JWTIssuer, cfg.JWTTTL)
	authService := services.NewAuthService(
		userRepo,
		idgen.UUIDGenerator[user.UserID]{},
		googleVerifier,
		tokenIssuer,
		cfg.AdminEmails,
		pictures,
	)
	authEndpoints := endpoints.NewAuthEndpoints(authService)

	userService := services.NewUserService(userRepo, pictures, imageProcessor, pictureWorker, picturePolicy)
	userEndpoints := endpoints.NewUserEndpoints(userService)

	corsCfg := endpoints.CORSConfig{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		AllowCredentials: cfg.CORSAllowCredentials,
		MaxAge:           cfg.CORSMaxAge,
	}

	rateCfg := endpoints.RateLimitConfig{
		Enabled:      cfg.RateLimitEnabled,
		Window:       cfg.RateLimitWindow,
		GlobalPerIP:  cfg.RateLimitGlobalPerIP,
		LoginPerIP:   cfg.RateLimitLoginPerIP,
		ReadPerUser:  cfg.RateLimitReadPerUser,
		WritePerUser: cfg.RateLimitWritePerUser,
	}

	// The ETag/Cache-Control layer is independent of Redis - it stays on
	// even with REDIS_URL unset.
	cacheCfg := endpoints.CacheConfig{
		HTTPMaxAge: int(cfg.CacheHTTPMaxAge.Seconds()),
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           endpoints.NewRouter(authEndpoints, standEndpoints, devilFruitEndpoints, userEndpoints, tokenIssuer, corsCfg, rateCfg, cacheCfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("JojoOnePieceSimulator2 backend listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serving: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.PictureJobTimeout+10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutting down http server: %v", err)
	}
	// Let in-flight transcodes finish before libvips is torn down - any job
	// still queued (not yet picked up by a worker) is left PENDING and must
	// be retried by re-uploading, since there is no durable job queue.
	if err := pictureWorker.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutting down picture worker: %v", err)
	}
	closeImaging()
}
