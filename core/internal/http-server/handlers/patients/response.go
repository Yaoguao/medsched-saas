package patients

import (
	"strings"
	"time"

	"base/internal/domain/models"
	"base/internal/service"

	"github.com/google/uuid"
)

type patientResponse struct {
	ID             uuid.UUID               `json:"id"`
	UserID         *uuid.UUID              `json:"user_id,omitempty"`
	FullName       string                  `json:"full_name"`
	Gender         string                  `json:"gender"`
	BirthDate      time.Time               `json:"birth_date"`
	Age            int                     `json:"age"`
	Phone          string                  `json:"phone"`
	Email          string                  `json:"email"`
	Neighbourhood  string                  `json:"neighbourhood"`
	Scholarship    bool                    `json:"scholarship"`
	ExternalEHRID  string                  `json:"external_ehr_id"`
	LastEHRSyncAt  *time.Time              `json:"last_ehr_sync_at,omitempty"`
	MedicalProfile *medicalProfileResponse `json:"medical_profile,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type ehrSyncResponse struct {
	Patient patientResponse          `json:"patient"`
	Event   integrationEventResponse `json:"event"`
}

type integrationEventResponse struct {
	ID            uuid.UUID  `json:"id"`
	Provider      string     `json:"provider"`
	EntityType    string     `json:"entity_type"`
	EntityID      uuid.UUID  `json:"entity_id"`
	Action        string     `json:"action"`
	Status        string     `json:"status"`
	Message       string     `json:"message"`
	ActorID       *uuid.UUID `json:"actor_id,omitempty"`
	AppliedFields []string   `json:"applied_fields"`
	OccurredAt    time.Time  `json:"occurred_at"`
}

type externalEntityLinkResponse struct {
	ID           uuid.UUID  `json:"id"`
	ProviderID   uuid.UUID  `json:"provider_id"`
	EntityType   string     `json:"entity_type"`
	InternalID   uuid.UUID  `json:"internal_id"`
	ExternalID   string     `json:"external_id"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

type ehrPreviewResponse struct {
	ProviderID   uuid.UUID                `json:"provider_id"`
	ProviderName string                   `json:"provider_name"`
	ProviderType string                   `json:"provider_type"`
	ExternalID   string                   `json:"external_id"`
	Changes      []ehrFieldChangeResponse `json:"changes"`
}

type ehrFieldChangeResponse struct {
	Field    string `json:"field"`
	Current  any    `json:"current"`
	External any    `json:"external"`
	Changed  bool   `json:"changed"`
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
		UserID:         patient.UserID,
		FullName:       patient.FullName,
		Gender:         patient.Gender,
		BirthDate:      patient.BirthDate,
		Age:            patient.Age,
		Phone:          patient.Phone,
		Email:          patient.Email,
		Neighbourhood:  patient.Neighbourhood,
		Scholarship:    patient.Scholarship,
		ExternalEHRID:  patient.ExternalEHRID,
		LastEHRSyncAt:  patient.LastEHRSyncAt,
		MedicalProfile: medicalProfileResponseFromModel(patient.MedicalProfile),
		CreatedAt:      patient.CreatedAt,
		UpdatedAt:      patient.UpdatedAt,
	}
}

func ehrSyncResponseFromService(result service.EHRSyncResult) ehrSyncResponse {
	return ehrSyncResponse{
		Patient: patientResponseFromModel(result.Patient),
		Event: integrationEventResponse{
			ID:            result.Event.ID,
			Provider:      result.Event.Provider,
			EntityType:    result.Event.EntityType,
			EntityID:      result.Event.EntityID,
			Action:        result.Event.Action,
			Status:        result.Event.Status,
			Message:       result.Event.Message,
			ActorID:       result.Event.ActorID,
			AppliedFields: splitAppliedFields(result.Event.AppliedFields),
			OccurredAt:    result.Event.OccurredAt,
		},
	}
}

func splitAppliedFields(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}

func externalEntityLinkResponseFromModel(link models.ExternalEntityLink) externalEntityLinkResponse {
	return externalEntityLinkResponse{
		ID:           link.ID,
		ProviderID:   link.ProviderID,
		EntityType:   link.EntityType,
		InternalID:   link.InternalID,
		ExternalID:   link.ExternalID,
		LastSyncedAt: link.LastSyncedAt,
	}
}

func ehrPreviewResponseFromService(preview service.EHRPatientPreview) ehrPreviewResponse {
	changes := make([]ehrFieldChangeResponse, 0, len(preview.Changes))
	for _, change := range preview.Changes {
		changes = append(changes, ehrFieldChangeResponse{
			Field: change.Field, Current: change.Current,
			External: change.External, Changed: change.Changed,
		})
	}
	return ehrPreviewResponse{
		ProviderID: preview.Provider.ID, ProviderName: preview.Provider.Name,
		ProviderType: preview.Provider.Type, ExternalID: preview.Link.ExternalID,
		Changes: changes,
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
