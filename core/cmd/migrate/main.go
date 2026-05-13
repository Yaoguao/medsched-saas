package main

import (
	"base/internal/config"
	"base/internal/lib/logger/sl"
	"base/internal/storage/postgres"
	"context"
	"log/slog"
	"os"
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
		log.Info("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := postgres.AutoMigrate(db); err != nil {
		log.Error("Failed to migrate models", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
