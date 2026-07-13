package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/safescope/safescope/apps/api/internal/application"
	"github.com/safescope/safescope/apps/api/internal/infrastructure/cache"
	"github.com/safescope/safescope/apps/api/internal/infrastructure/persistence"
	httpapi "github.com/safescope/safescope/apps/api/internal/interfaces/http"
	"github.com/safescope/safescope/apps/api/internal/platform/config"
	"github.com/safescope/safescope/apps/api/internal/platform/logger"
	"github.com/safescope/safescope/apps/api/internal/platform/security"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		healthcheck()
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := logger.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		return err
	}
	defer log.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := persistence.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	redis, err := cache.New(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("configure redis: %w", err)
	}
	defer redis.Close()
	if err := redis.Ping(ctx); err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}

	tokens := security.JWTManager{Secret: []byte(cfg.JWTSecret), TTL: cfg.JWTTTL}
	users := persistence.UserRepository{DB: db}
	projects := persistence.ProjectRepository{DB: db}
	assets := persistence.AssetRepository{DB: db}
	router := httpapi.NewRouter(httpapi.Dependencies{
		Auth:        application.NewAuthService(users, security.BcryptHasher{}, tokens),
		Users:       application.NewUserService(users),
		Projects:    application.NewProjectService(projects, assets),
		Dashboard:   application.NewDashboardService(db),
		Tokens:      tokens,
		Logger:      log,
		CORSOrigins: cfg.CORSOrigins,
		Health: func() error {
			probe, stop := context.WithTimeout(context.Background(), 2*time.Second)
			defer stop()
			if err := db.Ping(probe); err != nil {
				return err
			}
			return redis.Ping(probe)
		},
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		log.Info("api_started", zap.Int("port", cfg.Port), zap.String("environment", cfg.Environment))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("api_failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	return server.Shutdown(shutdown)
}

func healthcheck() {
	client := &http.Client{Timeout: 2 * time.Second}
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	response, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil || response.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	response.Body.Close()
}
