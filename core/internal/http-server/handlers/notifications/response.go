package notifications

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type notificationResponse struct {
	ID            uuid.UUID                  `json:"id"`
	AppointmentID uuid.UUID                  `json:"appointment_id"`
	PatientID     uuid.UUID                  `json:"patient_id"`
	Channel       models.NotificationChannel `json:"channel"`
	Type          models.NotificationType    `json:"type"`
	Status        models.NotificationStatus  `json:"status"`
	ScheduledAt   time.Time                  `json:"scheduled_at"`
	SentAt        *time.Time                 `json:"sent_at,omitempty"`
	Provider      string                     `json:"provider"`
	ExternalID    string                     `json:"external_id"`
	ErrorMessage  string                     `json:"error_message"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
}

func responsesFromModels(notifications []models.Notification) []notificationResponse {
	responses := make([]notificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		responses = append(responses, responseFromModel(notification))
	}
	return responses
}

func responseFromModel(notification models.Notification) notificationResponse {
	return notificationResponse{
		ID:            notification.ID,
		AppointmentID: notification.AppointmentID,
		PatientID:     notification.PatientID,
		Channel:       notification.Channel,
		Type:          notification.Type,
		Status:        notification.Status,
		ScheduledAt:   notification.ScheduledAt,
		SentAt:        notification.SentAt,
		Provider:      notification.Provider,
		ExternalID:    notification.ExternalID,
		ErrorMessage:  notification.ErrorMessage,
		CreatedAt:     notification.CreatedAt,
		UpdatedAt:     notification.UpdatedAt,
	}
}
