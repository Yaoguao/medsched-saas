package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID            uuid.UUID           `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppointmentID uuid.UUID           `gorm:"type:uuid;not null;index"`
	PatientID     uuid.UUID           `gorm:"type:uuid;not null;index"`
	Channel       NotificationChannel `gorm:"type:varchar(32);not null;index"`
	Type          NotificationType    `gorm:"type:varchar(32);not null"`
	Status        NotificationStatus  `gorm:"type:varchar(32);not null;default:'PENDING';index"`
	ScheduledAt   time.Time           `gorm:"not null;index"`
	SentAt        *time.Time
	Provider      string `gorm:"type:varchar(64)"`
	ExternalID    string `gorm:"type:varchar(255)"`
	ErrorMessage  string `gorm:"type:text"`

	Appointment Appointment `gorm:"foreignKey:AppointmentID"`
	Patient     Patient     `gorm:"foreignKey:PatientID"`

	BaseModel
}

func (Notification) TableName() string {
	return "notifications"
}
