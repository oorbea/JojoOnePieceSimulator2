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
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/api/endpoints"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/auth"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/idgen"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
)

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

	standRepository := repositories.NewStandRepository(pool)
	standService := services.NewStandService(standRepository, idgen.UUIDGenerator[powers.PowerID]{})
	standEndpoints := endpoints.NewStandEndpoints(standService)

	userRepository := repositories.NewUserRepository(pool)
	googleVerifier := auth.NewGoogleVerifier(cfg.GoogleClientID)
	tokenIssuer := auth.NewJWTIssuer([]byte(cfg.JWTSecret), cfg.JWTIssuer, cfg.JWTTTL)
	authService := services.NewAuthService(
		userRepository,
		idgen.UUIDGenerator[user.UserID]{},
		googleVerifier,
		tokenIssuer,
		cfg.AdminEmails,
	)
	authEndpoints := endpoints.NewAuthEndpoints(authService)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           endpoints.NewRouter(authEndpoints, standEndpoints, tokenIssuer),
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutting down: %v", err)
	}
}
