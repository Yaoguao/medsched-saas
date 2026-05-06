package models

import (
	"time"

	"github.com/google/uuid"
)

type PredictionResult struct {
	ID            uuid.UUID           `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppointmentID uuid.UUID           `gorm:"type:uuid;not null;uniqueIndex"`
	Probability   float64             `gorm:"type:numeric(5,4);not null"`
	RiskLevel     PredictionRiskLevel `gorm:"type:varchar(16);not null"`
	ModelVersion  string              `gorm:"type:varchar(64);not null"`
	RequestedAt   time.Time           `gorm:"not null"`
	ReceivedAt    time.Time           `gorm:"not null"`

	Appointment Appointment `gorm:"foreignKey:AppointmentID"`

	BaseModel
}

func (PredictionResult) TableName() string {
	return "prediction_results"
}
