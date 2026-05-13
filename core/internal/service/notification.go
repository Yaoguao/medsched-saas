package service

import (
	"context"
	"fmt"

	"base/internal/domain/models"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreateNotification(ctx context.Context, notification models.Notification) (models.Notification, error) {
	if notification.AppointmentID == uuid.Nil || notification.PatientID == uuid.Nil {
		return models.Notification{}, fmt.Errorf("%w: appointment_id and patient_id are required", repository.ErrValidation)
	}
	if notification.Status == "" {
		notification.Status = models.NotificationStatusPending
	}
	if notification.ScheduledAt.IsZero() {
		notification.ScheduledAt = s.now()
	}
	if err := s.store.CreateNotification(ctx, &notification); err != nil {
		return models.Notification{}, err
	}
	return notification, nil
}

func (s *Service) GetNotification(ctx context.Context, id uuid.UUID) (models.Notification, error) {
	return s.store.GetNotification(ctx, id)
}

func (s *Service) ListAppointmentNotifications(ctx context.Context, appointmentID uuid.UUID, page pagination.Page) ([]models.Notification, error) {
	return s.store.ListAppointmentNotifications(ctx, appointmentID, page)
}
