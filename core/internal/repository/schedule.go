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

func (s *Store) CreateSchedule(ctx context.Context, schedule *models.Schedule) error {
	return s.db.WithContext(ctx).Create(schedule).Error
}

func (s *Store) GetSchedule(ctx context.Context, id uuid.UUID) (models.Schedule, error) {
	var schedule models.Schedule
	err := s.db.WithContext(ctx).
		Preload("TimeSlots", func(db *gorm.DB) *gorm.DB {
			return db.Order("start_at asc")
		}).
		First(&schedule, "id = ?", id).Error
	return schedule, mapGormError(err)
}

func (s *Store) GetDoctorSchedules(ctx context.Context, doctorID uuid.UUID, page pagination.Page) ([]models.Schedule, error) {
	page = pagination.Normalize(page)

	var schedules []models.Schedule
	err := s.db.WithContext(ctx).
		Preload("TimeSlots", func(db *gorm.DB) *gorm.DB {
			return db.Order("start_at asc")
		}).
		Where("doctor_id = ?", doctorID).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("start_date asc").
		Find(&schedules).Error
	return schedules, err
}

func (s *Store) FindDoctorScheduleForDate(ctx context.Context, doctorID uuid.UUID, date time.Time) (models.Schedule, error) {
	var schedule models.Schedule
	err := s.db.WithContext(ctx).
		Where("doctor_id = ? AND status = ? AND start_date <= ? AND end_date >= ?", doctorID, "ACTIVE", date, date).
		Order("start_date desc").
		First(&schedule).Error
	return schedule, mapGormError(err)
}

func (s *Store) UpdateSchedule(ctx context.Context, schedule *models.Schedule) error {
	return mapGormError(s.db.WithContext(ctx).Save(schedule).Error)
}

func (s *Store) CreateTimeSlot(ctx context.Context, slot *models.TimeSlot) error {
	return s.db.WithContext(ctx).Create(slot).Error
}

func (s *Store) GetTimeSlot(ctx context.Context, id uuid.UUID) (models.TimeSlot, error) {
	var slot models.TimeSlot
	err := s.db.WithContext(ctx).First(&slot, "id = ?", id).Error
	return slot, mapGormError(err)
}

func (s *Store) GetTimeSlotForUpdate(ctx context.Context, id uuid.UUID) (models.TimeSlot, error) {
	var slot models.TimeSlot
	err := s.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&slot, "id = ?", id).Error
	return slot, mapGormError(err)
}

func (s *Store) UpdateTimeSlot(ctx context.Context, slot *models.TimeSlot) error {
	return mapGormError(s.db.WithContext(ctx).Save(slot).Error)
}

func (s *Store) HasDoctorSlotOverlap(ctx context.Context, doctorID uuid.UUID, startAt, endAt time.Time, excludeSlotID *uuid.UUID) (bool, error) {
	var count int64
	query := s.db.WithContext(ctx).
		Model(&models.TimeSlot{}).
		Where("doctor_id = ? AND start_at < ? AND end_at > ?", doctorID, endAt, startAt)
	if excludeSlotID != nil {
		query = query.Where("id <> ?", *excludeSlotID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) ListDoctorTimeSlotsBetween(ctx context.Context, doctorID uuid.UUID, startAt, endAt time.Time) ([]models.TimeSlot, error) {
	var slots []models.TimeSlot
	err := s.db.WithContext(ctx).
		Where("doctor_id = ? AND start_at < ? AND end_at > ?", doctorID, endAt, startAt).
		Order("start_at asc").
		Find(&slots).Error
	return slots, err
}

func (s *Store) HasActiveAppointmentForTimeSlot(ctx context.Context, slotID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&models.Appointment{}).
		Where("time_slot_id = ? AND status IN ?", slotID, []models.AppointmentStatus{
			models.AppointmentStatusCreated,
			models.AppointmentStatusConfirmed,
		}).
		Count(&count).Error
	return count > 0, err
}
