package repository

import (
	"context"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
)

func (s *Store) CreatePatient(ctx context.Context, patient *models.Patient) error {
	return s.db.WithContext(ctx).Create(patient).Error
}

func (s *Store) ListPatients(ctx context.Context, page pagination.Page) ([]models.Patient, error) {
	page = pagination.Normalize(page)

	var patients []models.Patient
	err := s.db.WithContext(ctx).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("created_at desc").
		Find(&patients).Error
	return patients, err
}

func (s *Store) GetPatient(ctx context.Context, id uuid.UUID) (models.Patient, error) {
	var patient models.Patient
	err := s.db.WithContext(ctx).
		Preload("MedicalProfile").
		First(&patient, "id = ?", id).Error
	return patient, mapGormError(err)
}

func (s *Store) GetPatientMedicalProfile(ctx context.Context, patientID uuid.UUID) (models.PatientMedicalProfile, error) {
	var profile models.PatientMedicalProfile
	err := s.db.WithContext(ctx).
		First(&profile, "patient_id = ?", patientID).Error
	return profile, mapGormError(err)
}

func (s *Store) UpdatePatient(ctx context.Context, patient *models.Patient) error {
	return mapGormError(s.db.WithContext(ctx).Save(patient).Error)
}

func (s *Store) ListPatientAppointments(ctx context.Context, patientID uuid.UUID, page pagination.Page) ([]models.Appointment, error) {
	page = pagination.Normalize(page)

	var appointments []models.Appointment
	err := s.db.WithContext(ctx).
		Preload("Doctor").
		Preload("TimeSlot").
		Where("patient_id = ?", patientID).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("appointment_at desc").
		Find(&appointments).Error
	return appointments, err
}
