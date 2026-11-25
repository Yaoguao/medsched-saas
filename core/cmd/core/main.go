package main

import (
	httpapp "base/internal/app/http"
	"base/internal/app/jobs"
	"base/internal/config"
	"base/internal/http-server/router"
	ehrclient "base/internal/integration/ehr"
	mlclient "base/internal/integration/ml"
	notifierclient "base/internal/integration/notifier"
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

	log.Debug(
		"config loaded",
		slog.String("env", cfg.Env),
		slog.String("http_address", cfg.HTTPServer.Address),
		slog.String("ml_base_url", cfg.ML.BaseURL),
		slog.String("notifier_base_url", cfg.Notifier.BaseURL),
	)

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
	mlClient := mlclient.NewClient(cfg.ML.BaseURL, cfg.ML.Timeout, cfg.ML.AdminToken)
	notifierClient := notifierclient.NewClient(cfg.Notifier.BaseURL, cfg.Notifier.Timeout, cfg.Notifier.Enabled)
	var fhirClient ehrclient.Client
	if cfg.FHIR.Enabled {
		fhirClient = ehrclient.NewFHIRClient(cfg.FHIR.BaseURL, cfg.FHIR.Timeout, cfg.FHIR.BearerToken)
	}
	coreService := service.New(
		store,
		service.WithAuth(cfg.Auth.TokenSecret, cfg.Auth.TokenTTL),
		service.WithMLClient(mlClient),
		service.WithNotifierClient(notifierClient),
		service.WithEHRClient(ehrclient.NewMockClient()),
		service.WithFHIRClient(fhirClient),
	)
	if _, err := coreService.EnsureDemoEHRProvider(ctx); err != nil {
		log.Error("failed to ensure Demo EHR provider", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if cfg.FHIR.Enabled {
		if _, err := coreService.EnsureFHIRProvider(ctx, service.FHIRProviderConfig{
			Code: cfg.FHIR.Code, Name: cfg.FHIR.Name, BaseURL: cfg.FHIR.BaseURL, Enabled: true,
		}); err != nil {
			log.Error("failed to ensure FHIR provider", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	adminUser, created, err := coreService.EnsureBootstrapAdmin(ctx, service.BootstrapAdminRequest{
		Enabled:  cfg.Bootstrap.Admin.Enabled,
		Email:    cfg.Bootstrap.Admin.Email,
		Phone:    cfg.Bootstrap.Admin.Phone,
		Password: cfg.Bootstrap.Admin.Password,
	})
	if err != nil {
		log.Error("failed to bootstrap admin user", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if cfg.Bootstrap.Admin.Enabled {
		log.Info("bootstrap admin ready", slog.String("email", adminUser.Email), slog.Bool("created", created))
	}

	handler := router.New(coreService, cfg.Env, cfg.Notifier.CallbackSecret, cfg.ML.DatasetServiceToken)

	runtimeCtx, runtimeCancel := context.WithCancel(context.Background())
	defer runtimeCancel()

	autoNoShowRunner := jobs.NewAutoNoShowRunner(log, coreService, jobs.AutoNoShowConfig{
		Enabled:  cfg.Jobs.AutoNoShow.Enabled,
		Interval: cfg.Jobs.AutoNoShow.Interval,
		Grace:    cfg.Jobs.AutoNoShow.Grace,
		Limit:    cfg.Jobs.AutoNoShow.Limit,
	})
	go autoNoShowRunner.Run(runtimeCtx)
	mlTrainingRunner := jobs.NewMLTrainingRunner(log, coreService, jobs.MLTrainingConfig{
		Enabled:       cfg.Jobs.MLTraining.Enabled,
		Interval:      cfg.Jobs.MLTraining.Interval,
		Timeout:       cfg.Jobs.MLTraining.Timeout,
		PythonPath:    cfg.Jobs.MLTraining.PythonPath,
		WorkDir:       cfg.Jobs.MLTraining.WorkDir,
		CoreURL:       cfg.Jobs.MLTraining.CoreURL,
		ServiceToken:  cfg.ML.DatasetServiceToken,
		DataPath:      cfg.Jobs.MLTraining.DataPath,
		ModelsDir:     cfg.Jobs.MLTraining.ModelsDir,
		MetricsPath:   cfg.Jobs.MLTraining.MetricsPath,
		MaxOutputSize: cfg.Jobs.MLTraining.MaxOutputSize,
	})
	go mlTrainingRunner.Run(runtimeCtx)

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
	runtimeCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Error("http server shutdown failed", slog.String("error", err.Error()))
	}
}
