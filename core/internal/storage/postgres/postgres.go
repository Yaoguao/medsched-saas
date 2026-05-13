package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"base/internal/config"
	"base/internal/domain/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(ctx context.Context, cfg config.StorageConfig, log *slog.Logger) (*gorm.DB, error) {
	attempts := cfg.Postgres.ConnAttempts
	if attempts <= 0 {
		attempts = 1
	}

	delay := cfg.Postgres.BaseRetryDelay
	if delay <= 0 {
		delay = 100 * time.Millisecond
	}

	var lastErr error
	for attempt := int32(1); attempt <= attempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		db, err := gorm.Open(postgres.New(postgres.Config{
			DSN: cfg.Postgres.DSN,
		}), &gorm.Config{
			Logger: logger.NewSlogLogger(log, logger.Config{LogLevel: logger.Warn}),
		})
		if err == nil {
			sqlDB, err := db.DB()
			if err != nil {
				return nil, fmt.Errorf("get sql db: %w", err)
			}
			sqlDB.SetMaxOpenConns(int(cfg.Postgres.MaxOpenConns))
			sqlDB.SetMaxIdleConns(int(cfg.Postgres.MaxIdleConns))
			sqlDB.SetConnMaxIdleTime(cfg.Postgres.MaxIdleTime)
			if err := sqlDB.PingContext(ctx); err != nil {
				lastErr = err
				_ = sqlDB.Close()
			} else {
				return db, nil
			}
		} else {
			lastErr = err
		}

		log.Warn("postgres connection attempt failed", "attempt", attempt, "error", lastErr)
		time.Sleep(delay)
		if cfg.Postgres.MaxRetryDelay > 0 && delay < cfg.Postgres.MaxRetryDelay {
			delay *= 2
			if delay > cfg.Postgres.MaxRetryDelay {
				delay = cfg.Postgres.MaxRetryDelay
			}
		}
	}

	return nil, fmt.Errorf("connect postgres after %d attempts: %w", attempts, lastErr)
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Patient{},
		&models.PatientMedicalProfile{},
		&models.Doctor{},
		&models.Schedule{},
		&models.TimeSlot{},
		&models.Appointment{},
		&models.AppointmentStatusHistory{},
		&models.Notification{},
		&models.PredictionResult{},
	)
}
