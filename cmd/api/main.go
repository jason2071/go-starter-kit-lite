package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	httpHandler "github.com/jason2071/go-starter-kit-lite/internal/handler/http"
	"github.com/jason2071/go-starter-kit-lite/internal/platform"
	"github.com/jason2071/go-starter-kit-lite/internal/repository"
	"github.com/jason2071/go-starter-kit-lite/internal/usecase"
)

func main() {
	_ = godotenv.Load()
	cfg, err := platform.LoadConfig()
	if err != nil {
		panic(err)
	}
	logger := platform.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	db, err := platform.NewDatabase(cfg)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	repo := repository.NewPostgres(db)
	security := platform.NewSecurity(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL)
	authService := usecase.NewAuthService(repo, repo, security, security, cfg.RefreshTTL)
	userService := usecase.NewUserService(repo)
	app := httpHandler.NewApp(httpHandler.Dependencies{Ready: sqlDB.Ping, Auth: authService, Users: userService, Tokens: security, Logger: logger, AllowedOrigins: cfg.AllowedOrigin})

	listenErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
		listenErr <- app.Listen(":" + cfg.AppPort)
	}()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal", "signal", sig.String())
	case err := <-listenErr:
		if err != nil {
			logger.Error("server stopped", "error", err)
		}
		return
	}
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}
