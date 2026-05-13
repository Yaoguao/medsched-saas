package main

import (
	httpapp "base/internal/app/http"
	"base/internal/config"
	"base/internal/http-server/router"
	"base/internal/lib/logger/sl"
	"base/internal/repository"
	"base/internal/service"
	"base/internal/storage/postgres"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("config loaded", "cfg", cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := postgres.Open(ctx, cfg.StorageConfig, log)
	if err != nil {
		log.Error("failed to connect postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := postgres.AutoMigrate(db); err != nil {
		log.Error("failed to migrate database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("failed to get sql db", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer sqlDB.Close()

	store := repository.NewStore(db)
	coreService := service.New(store)
	handler := router.New(coreService, cfg.Env)

	app := httpapp.New(
		log,
		cfg.HTTPServer.Address,
		cfg.HTTPServer.Timeout,
		cfg.HTTPServer.Timeout,
		cfg.HTTPServer.IdleTimeout,
		handler,
	)

	go func() {
		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			log.Error("http server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Error("http server shutdown failed", slog.String("error", err.Error()))
	}
}
