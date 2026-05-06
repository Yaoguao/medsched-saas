package models

import (
	"time"

	"github.com/google/uuid"
)

type Schedule struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	DoctorID      uuid.UUID `gorm:"type:uuid;not null;index"`
	StartDate     time.Time `gorm:"type:date;not null"`
	EndDate       time.Time `gorm:"type:date;not null"`
	WorkingDays   string    `gorm:"type:varchar(64)"`
	WorkStartTime string    `gorm:"type:varchar(16)"`
	WorkEndTime   string    `gorm:"type:varchar(16)"`
	Status        string    `gorm:"type:varchar(32);not null;default:'ACTIVE'"`

	Doctor    Doctor     `gorm:"foreignKey:DoctorID"`
	TimeSlots []TimeSlot `gorm:"foreignKey:ScheduleID"`

	BaseModel
}

func (Schedule) TableName() string {
	return "schedules"
}
