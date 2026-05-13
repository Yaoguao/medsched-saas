package doctors

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type doctorResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	FullName        string     `json:"full_name"`
	Specialty       string     `json:"specialty"`
	Cabinet         string     `json:"cabinet"`
	ExperienceYears int        `json:"experience_years"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func responsesFromModels(doctors []models.Doctor) []doctorResponse {
	responses := make([]doctorResponse, 0, len(doctors))
	for _, doctor := range doctors {
		responses = append(responses, responseFromModel(doctor))
	}
	return responses
}

func responseFromModel(doctor models.Doctor) doctorResponse {
	return doctorResponse{
		ID:              doctor.ID,
		UserID:          doctor.UserID,
		FullName:        doctor.FullName,
		Specialty:       doctor.Specialty,
		Cabinet:         doctor.Cabinet,
		ExperienceYears: doctor.ExperienceYears,
		Status:          doctor.Status,
		CreatedAt:       doctor.CreatedAt,
		UpdatedAt:       doctor.UpdatedAt,
	}
}
