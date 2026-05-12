package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID       uuid.UUID `json:"id,omitempty" validate:"omitempty,uuid4"`
	Channel  string    `json:"channel" validate:"required,oneof=email telegram sms"`
	Message  string    `json:"message" validate:"required,min=1,max=1000"`
	SendAt   time.Time `json:"send_at" validate:"required,gt"`
	Status   string    `json:"status,omitempty" validate:"omitempty,oneof=pending sent failed cancelled"`
	Attempts int       `json:"attempts,omitempty"`
}
