package jobs

import (
	"context"
	"log/slog"
	"time"

	"base/internal/service"
)

type AutoNoShowConfig struct {
	Enabled  bool
	Interval time.Duration
	Grace    time.Duration
	Limit    int
}

type AutoNoShowRunner struct {
	log *slog.Logger
	svc *service.Service
	cfg AutoNoShowConfig
}

func NewAutoNoShowRunner(log *slog.Logger, svc *service.Service, cfg AutoNoShowConfig) *AutoNoShowRunner {
	if cfg.Interval <= 0 {
		cfg.Interval = time.Minute
	}
	if cfg.Grace < 0 {
		cfg.Grace = 0
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 100
	}
	return &AutoNoShowRunner{
		log: log,
		svc: svc,
		cfg: cfg,
	}
}

func (r *AutoNoShowRunner) Run(ctx context.Context) {
	if !r.cfg.Enabled {
		r.log.Info("auto no-show job disabled")
		return
	}

	r.log.Info("auto no-show job started", "interval", r.cfg.Interval, "grace", r.cfg.Grace, "limit", r.cfg.Limit)
	r.runOnce(ctx)

	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.log.Info("auto no-show job stopped")
			return
		case <-ticker.C:
			r.runOnce(ctx)
		}
	}
}

func (r *AutoNoShowRunner) runOnce(ctx context.Context) {
	changed, err := r.svc.MarkOverdueAppointmentsNoShow(ctx, r.cfg.Grace, r.cfg.Limit)
	if err != nil {
		r.log.Error("auto no-show job failed", "error", err.Error())
		return
	}
	if changed > 0 {
		r.log.Info("auto no-show job marked appointments", "count", changed)
	}
}
