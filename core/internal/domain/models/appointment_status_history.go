package models

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentStatusHistory struct {
	ID            uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppointmentID uuid.UUID         `gorm:"type:uuid;not null;index"`
	OldStatus     AppointmentStatus `gorm:"type:varchar(32)"`
	NewStatus     AppointmentStatus `gorm:"type:varchar(32);not null"`
	ChangedBy     *uuid.UUID        `gorm:"type:uuid"`
	Reason        string            `gorm:"type:text"`
	ChangedAt     time.Time         `gorm:"not null;index"`

	Appointment Appointment `gorm:"foreignKey:AppointmentID"`

	BaseModel
}

func (AppointmentStatusHistory) TableName() string {
	return "appointment_status_history"
}
