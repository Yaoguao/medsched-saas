package repository

import (
	"context"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
)

func (s *Store) CreateNotification(ctx context.Context, notification *models.Notification) error {
	return s.db.WithContext(ctx).Create(notification).Error
}

func (s *Store) GetNotification(ctx context.Context, id uuid.UUID) (models.Notification, error) {
	var notification models.Notification
	err := s.db.WithContext(ctx).First(&notification, "id = ?", id).Error
	return notification, mapGormError(err)
}

func (s *Store) ListAppointmentNotifications(ctx context.Context, appointmentID uuid.UUID, page pagination.Page) ([]models.Notification, error) {
	page = pagination.Normalize(page)

	var notifications []models.Notification
	err := s.db.WithContext(ctx).
		Where("appointment_id = ?", appointmentID).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("scheduled_at asc").
		Find(&notifications).Error
	return notifications, err
}
