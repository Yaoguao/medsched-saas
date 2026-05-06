package main

import (
	"base/internal/config"
	"base/internal/domain/models"
	"base/internal/lib/logger/sl"
	"log/slog"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.MustLoad()

	log := sl.InitLogger(cfg.Env, os.Stdout)

	log.Debug("config loaded", "cfg", cfg)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: cfg.StorageConfig.Postgres.DSN,
	}), &gorm.Config{
		Logger: logger.NewSlogLogger(log, logger.Config{LogLevel: logger.Info}),
	})

	if err != nil {
		log.Info("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	//ctx := context.Background()

	err = db.AutoMigrate(
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
	if err != nil {
		log.Error("Failed to migrate models", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
