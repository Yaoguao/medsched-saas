package models

import (
	"time"

	"github.com/google/uuid"
)

type TimeSlot struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ScheduleID uuid.UUID  `gorm:"type:uuid;not null;index"`
	DoctorID   uuid.UUID  `gorm:"type:uuid;not null;index"`
	StartAt    time.Time  `gorm:"not null;index"`
	EndAt      time.Time  `gorm:"not null"`
	Status     SlotStatus `gorm:"type:varchar(32);not null;default:'FREE';index"`
	IsBlocked  bool       `gorm:"not null;default:false"`

	Schedule    Schedule     `gorm:"foreignKey:ScheduleID"`
	Doctor      Doctor       `gorm:"foreignKey:DoctorID"`
	Appointment *Appointment `gorm:"foreignKey:TimeSlotID"`

	BaseModel
}

func (TimeSlot) TableName() string {
	return "time_slots"
}
