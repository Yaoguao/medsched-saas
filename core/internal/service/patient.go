package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"base/internal/domain/models"
	"base/internal/integration/ehr"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreatePatient(ctx context.Context, patient models.Patient) (models.Patient, error) {
	if patient.FullName == "" {
		return models.Patient{}, fmt.Errorf("%w: full_name is required", repository.ErrValidation)
	}
	if patient.ID == uuid.Nil {
		patient.ID = uuid.New()
	}
	if patient.MedicalProfile != nil {
		if patient.MedicalProfile.ID == uuid.Nil {
			patient.MedicalProfile.ID = uuid.New()
		}
		patient.MedicalProfile.PatientID = patient.ID
		if patient.MedicalProfile.UpdatedAtProfile.IsZero() {
			patient.MedicalProfile.UpdatedAtProfile = s.now()
		}
	}
	if err := s.store.CreatePatient(ctx, &patient); err != nil {
		return models.Patient{}, err
	}
	return patient, nil
}

func (s *Service) ListPatients(ctx context.Context, page pagination.Page) ([]models.Patient, error) {
	return s.store.ListPatients(ctx, page)
}

func (s *Service) GetPatient(ctx context.Context, id uuid.UUID) (models.Patient, error) {
	return s.store.GetPatient(ctx, id)
}

func (s *Service) GetPatientMedicalProfile(ctx context.Context, patientID uuid.UUID) (models.PatientMedicalProfile, error) {
	if _, err := s.store.GetPatient(ctx, patientID); err != nil {
		return models.PatientMedicalProfile{}, err
	}
	return s.store.GetPatientMedicalProfile(ctx, patientID)
}

func (s *Service) UpdatePatientMedicalProfile(ctx context.Context, patientID uuid.UUID, patch PatientMedicalProfilePatch) (models.PatientMedicalProfile, error) {
	if _, err := s.store.GetPatient(ctx, patientID); err != nil {
		return models.PatientMedicalProfile{}, err
	}

	profile, err := s.store.GetPatientMedicalProfile(ctx, patientID)
	createProfile := false
	if err != nil {
		if !IsNotFound(err) {
			return models.PatientMedicalProfile{}, err
		}
		createProfile = true
		profile = models.PatientMedicalProfile{
			ID:               uuid.New(),
			PatientID:        patientID,
			UpdatedAtProfile: s.now(),
		}
	}

	if patch.ChronicConditions != nil {
		profile.ChronicConditions = *patch.ChronicConditions
	}
	if patch.RiskFactors != nil {
		profile.RiskFactors = *patch.RiskFactors
	}
	if patch.Allergies != nil {
		profile.Allergies = *patch.Allergies
	}
	if patch.Notes != nil {
		profile.Notes = *patch.Notes
	}
	if patch.HasHypertension != nil {
		profile.HasHypertension = *patch.HasHypertension
	}
	if patch.HasDiabetes != nil {
		profile.HasDiabetes = *patch.HasDiabetes
	}
	if patch.HasAlcoholism != nil {
		profile.HasAlcoholism = *patch.HasAlcoholism
	}
	if patch.HasHandicap != nil {
		profile.HasHandicap = *patch.HasHandicap
	}
	profile.UpdatedAtProfile = s.now()

	if createProfile {
		if err := s.store.CreatePatientMedicalProfile(ctx, &profile); err != nil {
			return models.PatientMedicalProfile{}, err
		}
		return profile, nil
	}

	if err := s.store.UpdatePatientMedicalProfile(ctx, &profile); err != nil {
		return models.PatientMedicalProfile{}, err
	}
	return profile, nil
}

func (s *Service) UpdatePatient(ctx context.Context, id uuid.UUID, patch PatientPatch) (models.Patient, error) {
	patient, err := s.store.GetPatient(ctx, id)
	if err != nil {
		return models.Patient{}, err
	}

	if patch.FullName != nil {
		patient.FullName = *patch.FullName
	}
	if patch.Gender != nil {
		patient.Gender = *patch.Gender
	}
	if patch.BirthDate != nil {
		patient.BirthDate = *patch.BirthDate
	}
	if patch.Age != nil {
		patient.Age = *patch.Age
	}
	if patch.Phone != nil {
		patient.Phone = *patch.Phone
	}
	if patch.Email != nil {
		patient.Email = *patch.Email
	}
	if patch.Neighbourhood != nil {
		patient.Neighbourhood = *patch.Neighbourhood
	}
	if patch.Scholarship != nil {
		patient.Scholarship = *patch.Scholarship
	}
	if patch.ExternalEHRID != nil {
		patient.ExternalEHRID = strings.TrimSpace(*patch.ExternalEHRID)
	}

	if err := s.store.UpdatePatient(ctx, &patient); err != nil {
		return models.Patient{}, err
	}
	return patient, nil
}

func (s *Service) LinkPatientUser(ctx context.Context, patientID uuid.UUID, userID uuid.UUID) (models.Patient, error) {
	var patient models.Patient
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		currentPatient, err := tx.GetPatient(ctx, patientID)
		if err != nil {
			return err
		}

		user, err := tx.GetUser(ctx, userID)
		if err != nil {
			return err
		}
		if user.Status != "ACTIVE" {
			return fmt.Errorf("%w: user must be active", repository.ErrValidation)
		}
		if user.Role != models.UserRolePatient {
			return fmt.Errorf("%w: only PATIENT user can be linked to patient profile", repository.ErrValidation)
		}

		linkedPatient, err := tx.GetPatientByUserID(ctx, userID)
		if err == nil && linkedPatient.ID != patientID {
			return fmt.Errorf("%w: user is already linked to another patient", repository.ErrConflict)
		}
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return err
		}

		currentPatient.UserID = &userID
		if err := tx.UpdatePatient(ctx, &currentPatient); err != nil {
			return err
		}

		patient = currentPatient
		return nil
	})
	if err != nil {
		return models.Patient{}, err
	}
	return s.store.GetPatient(ctx, patient.ID)
}

func (s *Service) ListPatientAppointments(ctx context.Context, patientID uuid.UUID, page pagination.Page) ([]models.Appointment, error) {
	if _, err := s.store.GetPatient(ctx, patientID); err != nil {
		return nil, err
	}
	return s.store.ListPatientAppointments(ctx, patientID, page)
}

func (s *Service) SyncPatientWithEHR(
	ctx context.Context,
	patientID uuid.UUID,
	req SyncPatientEHRRequest,
) (EHRSyncResult, error) {
	fields, err := validateEHRSyncFields(req.Fields)
	if err != nil {
		s.recordEHRSyncFailure(ctx, patientID, req, err)
		return EHRSyncResult{}, err
	}
	preview, err := s.PreviewPatientEHR(ctx, patientID, req.ProviderID)
	if err != nil {
		s.recordEHRSyncFailure(ctx, patientID, req, err)
		return EHRSyncResult{}, err
	}
	patient, err := s.store.GetPatient(ctx, patientID)
	if err != nil {
		s.recordEHRSyncFailure(ctx, patientID, req, err)
		return EHRSyncResult{}, err
	}
	profile := preview.External
	event := models.IntegrationEvent{
		ID:            uuid.New(),
		Provider:      preview.Provider.Code,
		EntityType:    "patient",
		EntityID:      patient.ID,
		Action:        "PROFILE_SYNC",
		ActorID:       req.ActorID,
		AppliedFields: strings.Join(fields, ","),
		OccurredAt:    s.now(),
	}

	now := s.now()
	patient.ExternalEHRID = profile.ExternalID
	patient.LastEHRSyncAt = &now

	medicalProfile, err := s.store.GetPatientMedicalProfile(ctx, patient.ID)
	createProfile := false
	if err != nil {
		if !IsNotFound(err) {
			return EHRSyncResult{}, err
		}
		createProfile = true
		medicalProfile = models.PatientMedicalProfile{
			ID:        uuid.New(),
			PatientID: patient.ID,
		}
	}
	applyEHRFields(&medicalProfile, profile, fields)
	medicalProfile.UpdatedAtProfile = now

	if err := s.store.UpdatePatient(ctx, &patient); err != nil {
		s.recordEHRSyncFailure(ctx, patientID, req, err)
		return EHRSyncResult{}, err
	}
	if createProfile {
		if err := s.store.CreatePatientMedicalProfile(ctx, &medicalProfile); err != nil {
			s.recordEHRSyncFailure(ctx, patientID, req, err)
			return EHRSyncResult{}, err
		}
	} else if err := s.store.UpdatePatientMedicalProfile(ctx, &medicalProfile); err != nil {
		s.recordEHRSyncFailure(ctx, patientID, req, err)
		return EHRSyncResult{}, err
	}

	event.Status = "SUCCESS"
	event.Message = fmt.Sprintf(
		"Медицинский профиль синхронизирован из %s, применено полей: %d",
		preview.Provider.Name,
		len(fields),
	)
	if err := s.store.CreateIntegrationEvent(ctx, &event); err != nil {
		return EHRSyncResult{}, err
	}
	preview.Link.LastSyncedAt = &now
	preview.Link.ExternalVersion = profile.ExternalVersion
	if err := s.store.UpsertExternalEntityLink(ctx, &preview.Link); err != nil {
		return EHRSyncResult{}, err
	}
	if err := s.store.MarkIntegrationProviderSynced(ctx, preview.Provider.ID, now); err != nil {
		return EHRSyncResult{}, err
	}

	patient.MedicalProfile = &medicalProfile
	return EHRSyncResult{Patient: patient, Event: event}, nil
}

var allowedEHRSyncFields = map[string]struct{}{
	"chronic_conditions": {},
	"risk_factors":       {},
	"allergies":          {},
	"notes":              {},
	"has_hypertension":   {},
	"has_diabetes":       {},
	"has_alcoholism":     {},
	"has_handicap":       {},
}

func validateEHRSyncFields(fields []string) ([]string, error) {
	result := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if _, ok := allowedEHRSyncFields[field]; !ok {
			return nil, fmt.Errorf("%w: unsupported EHR field %q", repository.ErrValidation, field)
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		result = append(result, field)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: at least one EHR field must be selected", repository.ErrValidation)
	}
	return result, nil
}

func applyEHRFields(
	target *models.PatientMedicalProfile,
	source ehr.PatientProfile,
	fields []string,
) {
	for _, field := range fields {
		switch field {
		case "chronic_conditions":
			target.ChronicConditions = source.ChronicConditions
		case "risk_factors":
			target.RiskFactors = source.RiskFactors
		case "allergies":
			target.Allergies = source.Allergies
		case "notes":
			target.Notes = source.Notes
		case "has_hypertension":
			target.HasHypertension = source.HasHypertension
		case "has_diabetes":
			target.HasDiabetes = source.HasDiabetes
		case "has_alcoholism":
			target.HasAlcoholism = source.HasAlcoholism
		case "has_handicap":
			target.HasHandicap = source.HasHandicap
		}
	}
}

func (s *Service) recordEHRSyncFailure(
	ctx context.Context,
	patientID uuid.UUID,
	req SyncPatientEHRRequest,
	syncErr error,
) {
	providerCode := req.ProviderID.String()
	if provider, err := s.store.GetIntegrationProvider(ctx, req.ProviderID); err == nil {
		providerCode = provider.Code
	}
	_ = s.store.CreateIntegrationEvent(ctx, &models.IntegrationEvent{
		ID:            uuid.New(),
		Provider:      providerCode,
		EntityType:    "patient",
		EntityID:      patientID,
		Action:        "PROFILE_SYNC",
		Status:        "FAILED",
		Message:       syncErr.Error(),
		ActorID:       req.ActorID,
		AppliedFields: strings.Join(req.Fields, ","),
		OccurredAt:    s.now(),
	})
}
