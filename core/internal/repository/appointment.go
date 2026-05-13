package repository

import (
	"context"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Store) CreateAppointment(ctx context.Context, appointment *models.Appointment) error {
	return s.db.WithContext(ctx).Create(appointment).Error
}

func (s *Store) GetAppointment(ctx context.Context, id uuid.UUID) (models.Appointment, error) {
	var appointment models.Appointment
	err := s.db.WithContext(ctx).
		Preload("Patient").
		Preload("Patient.MedicalProfile").
		Preload("Doctor").
		Preload("TimeSlot").
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("changed_at asc")
		}).
		Preload("Notifications").
		Preload("PredictionResult").
		First(&appointment, "id = ?", id).Error
	return appointment, mapGormError(err)
}

func (s *Store) GetAppointmentForUpdate(ctx context.Context, id uuid.UUID) (models.Appointment, error) {
	var appointment models.Appointment
	err := s.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&appointment, "id = ?", id).Error
	return appointment, mapGormError(err)
}

func (s *Store) UpdateAppointment(ctx context.Context, appointment *models.Appointment) error {
	return mapGormError(s.db.WithContext(ctx).Save(appointment).Error)
}

func (s *Store) ListAppointments(ctx context.Context, page pagination.Page) ([]models.Appointment, error) {
	page = pagination.Normalize(page)

	var appointments []models.Appointment
	err := s.db.WithContext(ctx).
		Preload("Patient").
		Preload("Doctor").
		Preload("TimeSlot").
		Limit(page.Limit).
		Offset(page.Offset).
		Order("appointment_at desc").
		Find(&appointments).Error
	return appointments, err
}

func (s *Store) ListCompletedAppointmentsForTraining(ctx context.Context, limit int) ([]models.Appointment, error) {
	if limit <= 0 {
		limit = 5000
	}

	var appointments []models.Appointment
	err := s.db.WithContext(ctx).
		Preload("Patient").
		Preload("Patient.MedicalProfile").
		Preload("Doctor").
		Preload("TimeSlot").
		Where("showed_up IS NOT NULL").
		Order("appointment_at asc").
		Limit(limit).
		Find(&appointments).Error
	return appointments, err
}

func (s *Store) CreateAppointmentStatusHistory(ctx context.Context, history *models.AppointmentStatusHistory) error {
	return s.db.WithContext(ctx).Create(history).Error
}
