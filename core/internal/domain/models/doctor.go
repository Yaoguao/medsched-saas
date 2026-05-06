package models

import "github.com/google/uuid"

type Doctor struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          *uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	FullName        string     `gorm:"type:varchar(255);not null"`
	Specialty       string     `gorm:"type:varchar(255);not null;index"`
	Cabinet         string     `gorm:"type:varchar(64)"`
	ExperienceYears int        `gorm:"not null;default:0"`
	Status          string     `gorm:"type:varchar(32);not null;default:'ACTIVE'"`

	User         *User         `gorm:"foreignKey:UserID"`
	Schedules    []Schedule    `gorm:"foreignKey:DoctorID"`
	TimeSlots    []TimeSlot    `gorm:"foreignKey:DoctorID"`
	Appointments []Appointment `gorm:"foreignKey:DoctorID"`

	BaseModel
}

func (Doctor) TableName() string {
	return "doctors"
}
