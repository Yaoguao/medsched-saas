package patients

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type patientResponse struct {
	ID             uuid.UUID               `json:"id"`
	FullName       string                  `json:"full_name"`
	Gender         string                  `json:"gender"`
	BirthDate      time.Time               `json:"birth_date"`
	Age            int                     `json:"age"`
	Phone          string                  `json:"phone"`
	Email          string                  `json:"email"`
	Neighbourhood  string                  `json:"neighbourhood"`
	Scholarship    bool                    `json:"scholarship"`
	MedicalProfile *medicalProfileResponse `json:"medical_profile,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type medicalProfileResponse struct {
	ID                uuid.UUID `json:"id"`
	PatientID         uuid.UUID `json:"patient_id"`
	ChronicConditions string    `json:"chronic_conditions"`
	RiskFactors       string    `json:"risk_factors"`
	Allergies         string    `json:"allergies"`
	Notes             string    `json:"notes"`
	HasHypertension   bool      `json:"has_hypertension"`
	HasDiabetes       bool      `json:"has_diabetes"`
	HasAlcoholism     bool      `json:"has_alcoholism"`
	HasHandicap       bool      `json:"has_handicap"`
	UpdatedAtProfile  time.Time `json:"updated_at_profile"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func patientResponsesFromModels(patients []models.Patient) []patientResponse {
	responses := make([]patientResponse, 0, len(patients))
	for _, patient := range patients {
		responses = append(responses, patientResponseFromModel(patient))
	}
	return responses
}

func patientResponseFromModel(patient models.Patient) patientResponse {
	return patientResponse{
		ID:             patient.ID,
		FullName:       patient.FullName,
		Gender:         patient.Gender,
		BirthDate:      patient.BirthDate,
		Age:            patient.Age,
		Phone:          patient.Phone,
		Email:          patient.Email,
		Neighbourhood:  patient.Neighbourhood,
		Scholarship:    patient.Scholarship,
		MedicalProfile: medicalProfileResponseFromModel(patient.MedicalProfile),
		CreatedAt:      patient.CreatedAt,
		UpdatedAt:      patient.UpdatedAt,
	}
}

func medicalProfileResponseFromModel(profile *models.PatientMedicalProfile) *medicalProfileResponse {
	if profile == nil {
		return nil
	}
	return &medicalProfileResponse{
		ID:                profile.ID,
		PatientID:         profile.PatientID,
		ChronicConditions: profile.ChronicConditions,
		RiskFactors:       profile.RiskFactors,
		Allergies:         profile.Allergies,
		Notes:             profile.Notes,
		HasHypertension:   profile.HasHypertension,
		HasDiabetes:       profile.HasDiabetes,
		HasAlcoholism:     profile.HasAlcoholism,
		HasHandicap:       profile.HasHandicap,
		UpdatedAtProfile:  profile.UpdatedAtProfile,
		CreatedAt:         profile.CreatedAt,
		UpdatedAt:         profile.UpdatedAt,
	}
}
