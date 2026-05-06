package models

import (
	"time"

	"github.com/google/uuid"
)

type PatientMedicalProfile struct {
	ID                uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PatientID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	ChronicConditions string    `gorm:"type:text"`
	RiskFactors       string    `gorm:"type:text"`
	Allergies         string    `gorm:"type:text"`
	Notes             string    `gorm:"type:text"`

	HasHypertension bool `gorm:"not null;default:false"`
	HasDiabetes     bool `gorm:"not null;default:false"`
	HasAlcoholism   bool `gorm:"not null;default:false"`
	HasHandicap     bool `gorm:"not null;default:false"`

	UpdatedAtProfile time.Time `gorm:"column:updated_at_profile;not null"`

	Patient Patient `gorm:"foreignKey:PatientID;constraint:OnDelete:CASCADE"`

	BaseModel
}

func (PatientMedicalProfile) TableName() string {
	return "patient_medical_profiles"
}
