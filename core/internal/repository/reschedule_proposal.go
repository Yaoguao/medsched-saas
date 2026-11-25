package repository

import (
	"context"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) CreateRescheduleProposal(ctx context.Context, proposal *models.AppointmentRescheduleProposal) error {
	return s.db.WithContext(ctx).Create(proposal).Error
}

func (s *Store) GetRescheduleProposal(ctx context.Context, id uuid.UUID) (models.AppointmentRescheduleProposal, error) {
	var proposal models.AppointmentRescheduleProposal
	err := s.db.WithContext(ctx).
		Preload("Appointment").
		Preload("Appointment.Patient").
		Preload("Appointment.Patient.MedicalProfile").
		Preload("Appointment.Doctor").
		Preload("Appointment.TimeSlot").
		Preload("Appointment.PredictionResult").
		Preload("ProposedTimeSlot").
		First(&proposal, "id = ?", id).Error
	return proposal, mapGormError(err)
}

func (s *Store) GetRescheduleProposalForUpdate(ctx context.Context, id uuid.UUID) (models.AppointmentRescheduleProposal, error) {
	var proposal models.AppointmentRescheduleProposal
	err := s.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&proposal, "id = ?", id).Error
	return proposal, mapGormError(err)
}

func (s *Store) UpdateRescheduleProposal(ctx context.Context, proposal *models.AppointmentRescheduleProposal) error {
	return mapGormError(s.db.WithContext(ctx).Save(proposal).Error)
}

func (s *Store) ListAppointmentRescheduleProposals(ctx context.Context, appointmentID uuid.UUID, page pagination.Page) ([]models.AppointmentRescheduleProposal, error) {
	page = pagination.Normalize(page)

	var proposals []models.AppointmentRescheduleProposal
	err := preloadRescheduleProposalList(s.db.WithContext(ctx)).
		Where("appointment_id = ?", appointmentID).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("created_at desc").
		Find(&proposals).Error
	return proposals, err
}

func (s *Store) ListPatientRescheduleProposals(ctx context.Context, patientID uuid.UUID, page pagination.Page) ([]models.AppointmentRescheduleProposal, error) {
	page = pagination.Normalize(page)

	var proposals []models.AppointmentRescheduleProposal
	err := preloadRescheduleProposalList(s.db.WithContext(ctx)).
		Joins("JOIN appointments ON appointments.id = appointment_reschedule_proposals.appointment_id").
		Where("appointments.patient_id = ?", patientID).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("appointment_reschedule_proposals.created_at desc").
		Find(&proposals).Error
	return proposals, err
}

func (s *Store) PendingRescheduleProposalExists(ctx context.Context, appointmentID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&models.AppointmentRescheduleProposal{}).
		Where("appointment_id = ? AND status = ?", appointmentID, models.RescheduleProposalPending).
		Count(&count).Error
	return count > 0, err
}

func preloadRescheduleProposalList(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Appointment").
		Preload("Appointment.Patient").
		Preload("Appointment.Patient.MedicalProfile").
		Preload("Appointment.Doctor").
		Preload("Appointment.TimeSlot").
		Preload("Appointment.PredictionResult").
		Preload("ProposedTimeSlot")
}
