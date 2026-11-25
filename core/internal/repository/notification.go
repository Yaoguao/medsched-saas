package repository

import (
	"context"
	"time"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
)

func (s *Store) CreateNotification(ctx context.Context, notification *models.Notification) error {
	return s.db.WithContext(ctx).Create(notification).Error
}

func (s *Store) NotificationExternalIDExists(ctx context.Context, externalID string) (bool, error) {
	if externalID == "" {
		return false, nil
	}

	var count int64
	err := s.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("external_id = ?", externalID).
		Count(&count).Error
	return count > 0, err
}

func (s *Store) CreateNotificationIfMissing(ctx context.Context, notification *models.Notification) (bool, error) {
	exists, err := s.NotificationExternalIDExists(ctx, notification.ExternalID)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	if err := s.CreateNotification(ctx, notification); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) GetNotification(ctx context.Context, id uuid.UUID) (models.Notification, error) {
	var notification models.Notification
	err := s.db.WithContext(ctx).First(&notification, "id = ?", id).Error
	return notification, mapGormError(err)
}

func (s *Store) UpdateNotificationDeliveryStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus, sentAt *time.Time, errorMessage string) (models.Notification, error) {
	updates := map[string]any{
		"status":        status,
		"error_message": errorMessage,
	}
	if sentAt != nil {
		updates["sent_at"] = sentAt
	}

	if err := s.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return models.Notification{}, mapGormError(err)
	}
	return s.GetNotification(ctx, id)
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

func (s *Store) MarkPendingAppointmentNotificationsFailed(ctx context.Context, appointmentID uuid.UUID, reason string, before time.Time) error {
	query := s.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("appointment_id = ? AND status = ?", appointmentID, models.NotificationStatusPending)
	if !before.IsZero() {
		query = query.Where("scheduled_at >= ?", before)
	}
	return mapGormError(query.Updates(map[string]any{
		"status":        models.NotificationStatusFailed,
		"error_message": reason,
	}).Error)
}
