package service

import (
	"context"
	"fmt"
	"time"

	"base/internal/domain/models"
	"base/internal/repository"
)

func (s *Service) AppointmentStatusReport(ctx context.Context, from, to time.Time) (map[models.AppointmentStatus]int64, error) {
	if from.IsZero() {
		from = s.now().AddDate(0, -1, 0)
	}
	if to.IsZero() {
		to = s.now()
	}
	if to.Before(from) {
		return nil, fmt.Errorf("%w: invalid report date range", repository.ErrValidation)
	}
	return s.store.CountAppointmentsByStatus(ctx, from, to)
}
