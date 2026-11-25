package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID             uuid.UUID         `json:"id,omitempty" validate:"omitempty,uuid4"`
	ExternalID     string            `json:"external_id,omitempty"`
	AppointmentID  uuid.UUID         `json:"appointment_id,omitempty"`
	PatientID      uuid.UUID         `json:"patient_id,omitempty"`
	Channel        string            `json:"channel" validate:"required,oneof=email telegram sms"`
	Recipient      string            `json:"recipient,omitempty"`
	TemplateID     string            `json:"template_id,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
	Subject        string            `json:"subject,omitempty"`
	Message        string            `json:"message,omitempty" validate:"omitempty,min=1,max=4000"`
	SendAt         time.Time         `json:"send_at" validate:"required,gt"`
	Status         string            `json:"status,omitempty" validate:"omitempty,oneof=pending sent failed cancelled"`
	Attempts       int               `json:"attempts,omitempty"`
	AttemptHistory []DeliveryAttempt `json:"attempt_history,omitempty"`
	LastError      string            `json:"last_error,omitempty"`
	CreatedAt      time.Time         `json:"created_at,omitempty"`
	UpdatedAt      time.Time         `json:"updated_at,omitempty"`
}

type DeliveryAttempt struct {
	AttemptedAt time.Time `json:"attempted_at"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
}

type DeadLetter struct {
	ID             uuid.UUID         `json:"id"`
	NotificationID uuid.UUID         `json:"notification_id"`
	ExternalID     string            `json:"external_id,omitempty"`
	Channel        string            `json:"channel"`
	Recipient      string            `json:"recipient,omitempty"`
	Reason         string            `json:"reason"`
	Attempts       int               `json:"attempts"`
	AttemptHistory []DeliveryAttempt `json:"attempt_history,omitempty"`
	Notification   Notification      `json:"notification"`
	CreatedAt      time.Time         `json:"created_at"`
}

type NotificationTemplate struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Subject   string    `json:"subject,omitempty"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Command struct {
	Type         string                `json:"type"`
	Notification *Notification         `json:"notification,omitempty"`
	Template     *NotificationTemplate `json:"template,omitempty"`
	ID           uuid.UUID             `json:"id,omitempty"`
}
