package repository

import (
	"context"
	"time"

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
		Preload("Patient.MedicalProfile").
		Preload("Doctor").
		Preload("TimeSlot").
		Preload("PredictionResult").
		Limit(page.Limit).
		Offset(page.Offset).
		Order("appointment_at desc").
		Find(&appointments).Error
	return appointments, err
}

func (s *Store) ListDoctorAppointments(ctx context.Context, doctorID uuid.UUID, page pagination.Page) ([]models.Appointment, error) {
	page = pagination.Normalize(page)

	var appointments []models.Appointment
	err := s.db.WithContext(ctx).
		Preload("Patient").
		Preload("Patient.MedicalProfile").
		Preload("Doctor").
		Preload("TimeSlot").
		Preload("PredictionResult").
		Where("doctor_id = ?", doctorID).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("appointment_at desc").
		Find(&appointments).Error
	return appointments, err
}

func (s *Store) ListOverdueActiveAppointments(ctx context.Context, before time.Time, limit int) ([]models.Appointment, error) {
	if limit <= 0 {
		limit = 100
	}

	var appointments []models.Appointment
	err := s.db.WithContext(ctx).
		Where("status IN ? AND appointment_at < ?", []models.AppointmentStatus{
			models.AppointmentStatusCreated,
			models.AppointmentStatusConfirmed,
		}, before).
		Order("appointment_at asc").
		Limit(limit).
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
		Preload("Notifications").
		Where("showed_up IS NOT NULL").
		Order("appointment_at asc").
		Limit(limit).
		Find(&appointments).Error
	return appointments, err
}

func (s *Store) PatientAttendanceHistory(
	ctx context.Context,
	patientID uuid.UUID,
	before time.Time,
	excludeAppointmentID uuid.UUID,
) (int, int, error) {
	var result struct {
		AppointmentsCount int
		NoShowCount       int
	}
	query := s.db.WithContext(ctx).
		Model(&models.Appointment{}).
		Select(
			"COUNT(*) AS appointments_count, "+
				"COALESCE(SUM(CASE WHEN showed_up = false THEN 1 ELSE 0 END), 0) AS no_show_count",
		).
		Where("patient_id = ? AND showed_up IS NOT NULL AND appointment_at < ?", patientID, before)
	if excludeAppointmentID != uuid.Nil {
		query = query.Where("id <> ?", excludeAppointmentID)
	}
	if err := query.Scan(&result).Error; err != nil {
		return 0, 0, err
	}
	return result.AppointmentsCount, result.NoShowCount, nil
}

func (s *Store) CreateAppointmentStatusHistory(ctx context.Context, history *models.AppointmentStatusHistory) error {
	return s.db.WithContext(ctx).Create(history).Error
}
