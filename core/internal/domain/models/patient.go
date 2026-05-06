package models

import (
	"time"

	"github.com/google/uuid"
)

type Patient struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FullName      string    `gorm:"type:varchar(255);not null"`
	Gender        string    `gorm:"type:varchar(16);not null"`
	BirthDate     time.Time `gorm:"type:date"`
	Age           int       `gorm:"not null"`
	Phone         string    `gorm:"type:varchar(32);index"`
	Email         string    `gorm:"type:varchar(255);index"`
	Neighbourhood string    `gorm:"type:varchar(255)"`
	Scholarship   bool      `gorm:"not null;default:false"`

	MedicalProfile *PatientMedicalProfile `gorm:"foreignKey:PatientID"`
	Appointments   []Appointment          `gorm:"foreignKey:PatientID"`
	Notifications  []Notification         `gorm:"foreignKey:PatientID"`

	BaseModel
}

func (Patient) TableName() string {
	return "patients"
}
