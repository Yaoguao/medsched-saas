package service

import (
	"context"
	"fmt"

	"base/internal/domain/models"
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

	if err := s.store.UpdatePatient(ctx, &patient); err != nil {
		return models.Patient{}, err
	}
	return patient, nil
}

func (s *Service) ListPatientAppointments(ctx context.Context, patientID uuid.UUID, page pagination.Page) ([]models.Appointment, error) {
	if _, err := s.store.GetPatient(ctx, patientID); err != nil {
		return nil, err
	}
	return s.store.ListPatientAppointments(ctx, patientID, page)
}
