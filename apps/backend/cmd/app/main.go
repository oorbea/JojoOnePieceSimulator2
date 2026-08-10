package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/application/services"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/auth"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache"
	rediscache "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/cache/redis"
	gameinfra "github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/game"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/gamestore"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/imaging"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/random"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/fallback"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/s3store"
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

	// storageBackends/storageTiers are built in cfg.StorageProviders order -
	// R2 is always first (config.Load enforces this). Each provider speaks
	// the same S3 API (s3store), so only endpoint/region/quota differ.
	storageBackends, storageTiers, err := buildStorageTiers(ctx, cfg)
	if err != nil {
		log.Fatalf("configuring storage backends: %v", err)
	}
	storageLedger := repositories.NewStorageLedger(pool)
	pictureStorage, err := fallback.New(ctx, storageTiers, storageLedger, cfg.StorageQuotaThresholdPct)
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

	// pictureHub fans out PENDING->READY/FAILED transitions to connected SSE
	// clients (events_endpoints.go) - a single in-process pub/sub is enough
	// since there's only ever one backend instance.
	pictureHub := services.NewPictureEventHub()

	pictureWorker := services.NewPictureWorker(imageProcessor, pictures, pictureTargets,
		idgen.UUIDGenerator[powers.PowerID]{}, services.WorkerConfig{
			Workers:        cfg.PictureWorkers,
			QueueSize:      cfg.PictureQueueSize,
			JobTimeout:     cfg.PictureJobTimeout,
			MaxDimension:   cfg.PictureMaxDimension,
			ThumbDimension: cfg.PictureThumbDimension,
			Quality:        cfg.PictureWebPQuality,
		}, pictureHub)
	pictureWorker.Start()

	// The reconciler walks every configured bucket to correct any drift
	// between it and the ledger (Record/Forget on the fallback chain are
	// best-effort). It shares ctx with the server, so it stops on the same
	// shutdown signal - no separate teardown needed.
	storageReconciler := services.NewStorageReconciler(storageBackends, storageLedger, pictureStorage, cfg.StorageReconcileInterval)
	go storageReconciler.Start(ctx)

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

	// Game (Gauntlet/Versus) application layer. Websockets + the routes
	// that would expose this over HTTP are the next tanda (see
	// ObsidianVault/ADR.md) - it is wired here, ready to be handed to a
	// router, but nothing calls into it yet.
	gameStore := gamestore.NewMemoryGameStore()
	gameReaper := gamestore.NewReaper(gameStore, cfg.GameLobbyTTL, cfg.GameLobbyReapInterval)
	go gameReaper.Start(ctx)

	gameRNG := random.NewStdRandomGenerator[string]()
	gameEventHub := services.NewGameEventHub()
	gameService := services.NewGameService(
		gameStore,
		idgen.UUIDGenerator[game.GameID]{},
		idgen.UUIDGenerator[game.ParticipantID]{},
		idgen.UUIDGenerator[game.TeamID]{},
		userRepo,
		gameinfra.NewStaticStageCatalog(),
		gameinfra.NewRepoPowerPool(standRepo, devilFruitRepo),
		gameinfra.NewDefaultWeights(),
		gameinfra.NewCoinFlipTiebreaker(gameRNG),
		// No ports.IGameHistory adapter yet - finished/aborted games are
		// simply dropped from the store once the match ends.
		nil,
		gameRNG,
		gameEventHub,
		services.NewSystemClock(),
		services.VotingPolicy{Window: cfg.GameVotingWindow},
	)
	_ = gameService

	// ctx (cancelled on SIGINT/SIGTERM) lets the stream handler exit
	// promptly on shutdown instead of blocking srv.Shutdown's grace window.
	eventsEndpoints := endpoints.NewEventsEndpoints(pictureHub, tokenIssuer, ctx)

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
		Handler:           endpoints.NewRouter(authEndpoints, standEndpoints, devilFruitEndpoints, userEndpoints, eventsEndpoints, tokenIssuer, corsCfg, rateCfg, cacheCfg),
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

// buildStorageTiers constructs one s3store.Backend per provider listed in
// cfg.StorageProviders, in that order, paired with each provider's quota
// into fallback.Tier. It returns the backends both as the plain slice the
// reconciler walks and as the Tier slice the fallback chain picks from.
func buildStorageTiers(ctx context.Context, cfg *config.Config) ([]ports.IStorageBackend, []fallback.Tier, error) {
	backends := make([]ports.IStorageBackend, 0, len(cfg.StorageProviders))
	tiers := make([]fallback.Tier, 0, len(cfg.StorageProviders))

	for _, name := range cfg.StorageProviders {
		var s3Cfg s3store.Config
		var quota int64

		switch name {
		case "r2":
			s3Cfg = s3store.Config{
				Name: "r2", Endpoint: s3store.R2Endpoint(cfg.R2AccountID), Region: "auto",
				AccessKeyID: cfg.R2AccessKeyID, SecretAccessKey: cfg.R2SecretAccessKey,
				Bucket: cfg.R2Bucket, PresignTTL: cfg.R2PresignTTL,
			}
			quota = cfg.R2QuotaBytes
		case "b2":
			s3Cfg = s3store.Config{
				Name: "b2", Endpoint: cfg.B2Endpoint, Region: cfg.B2Region,
				AccessKeyID: cfg.B2AccessKeyID, SecretAccessKey: cfg.B2SecretAccessKey,
				Bucket: cfg.B2Bucket, PresignTTL: cfg.R2PresignTTL,
			}
			quota = cfg.B2QuotaBytes
		case "supabase":
			s3Cfg = s3store.Config{
				Name: "supabase", Endpoint: cfg.SupabaseEndpoint, Region: cfg.SupabaseRegion,
				AccessKeyID: cfg.SupabaseAccessKeyID, SecretAccessKey: cfg.SupabaseSecretAccessKey,
				Bucket: cfg.SupabaseBucket, PresignTTL: cfg.R2PresignTTL,
			}
			quota = cfg.SupabaseQuotaBytes
		default:
			// Unreachable: config.Load already rejects unknown providers.
			return nil, nil, fmt.Errorf("unknown storage provider %q", name)
		}

		backend, err := s3store.New(ctx, s3Cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("configuring %s storage backend: %w", name, err)
		}
		backends = append(backends, backend)
		tiers = append(tiers, fallback.Tier{Backend: backend, QuotaBytes: quota})
	}

	return backends, tiers, nil
}
