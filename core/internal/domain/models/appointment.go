package models

import (
	"time"

	"github.com/google/uuid"
)

type Appointment struct {
	ID             uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PatientID      uuid.UUID         `gorm:"type:uuid;not null;index"`
	DoctorID       uuid.UUID         `gorm:"type:uuid;not null;index"`
	TimeSlotID     uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex"`
	ScheduledAt    time.Time         `gorm:"not null;index"`
	AppointmentAt  time.Time         `gorm:"not null;index"`
	Status         AppointmentStatus `gorm:"type:varchar(32);not null;default:'CREATED';index"`
	BookingChannel BookingChannel    `gorm:"type:varchar(32);not null"`
	VisitType      VisitType         `gorm:"type:varchar(32);not null"`
	ReminderSent   bool              `gorm:"not null;default:false"`
	ShowedUp       *bool
	ConfirmedAt    *time.Time
	CancelledAt    *time.Time
	CompletedAt    *time.Time

	Patient             Patient                         `gorm:"foreignKey:PatientID"`
	Doctor              Doctor                          `gorm:"foreignKey:DoctorID"`
	TimeSlot            TimeSlot                        `gorm:"foreignKey:TimeSlotID"`
	StatusHistory       []AppointmentStatusHistory      `gorm:"foreignKey:AppointmentID"`
	RescheduleProposals []AppointmentRescheduleProposal `gorm:"foreignKey:AppointmentID"`
	Notifications       []Notification                  `gorm:"foreignKey:AppointmentID"`
	PredictionResult    *PredictionResult               `gorm:"foreignKey:AppointmentID"`

	BaseModel
}

func (Appointment) TableName() string {
	return "appointments"
}
