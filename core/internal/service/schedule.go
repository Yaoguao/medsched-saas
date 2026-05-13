package service

import (
	"context"
	"fmt"

	"base/internal/domain/models"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreateSchedule(ctx context.Context, schedule models.Schedule) (models.Schedule, error) {
	if schedule.DoctorID == uuid.Nil {
		return models.Schedule{}, fmt.Errorf("%w: doctor_id is required", repository.ErrValidation)
	}
	if schedule.StartDate.IsZero() || schedule.EndDate.IsZero() || schedule.EndDate.Before(schedule.StartDate) {
		return models.Schedule{}, fmt.Errorf("%w: invalid schedule date range", repository.ErrValidation)
	}
	if _, err := s.store.GetDoctor(ctx, schedule.DoctorID); err != nil {
		return models.Schedule{}, err
	}
	if err := s.store.CreateSchedule(ctx, &schedule); err != nil {
		return models.Schedule{}, err
	}
	return schedule, nil
}

func (s *Service) GetDoctorSchedules(ctx context.Context, doctorID uuid.UUID, page pagination.Page) ([]models.Schedule, error) {
	if _, err := s.store.GetDoctor(ctx, doctorID); err != nil {
		return nil, err
	}
	return s.store.GetDoctorSchedules(ctx, doctorID, page)
}

func (s *Service) UpdateSchedule(ctx context.Context, id uuid.UUID, patch SchedulePatch) (models.Schedule, error) {
	schedule, err := s.store.GetSchedule(ctx, id)
	if err != nil {
		return models.Schedule{}, err
	}
	if patch.StartDate != nil {
		schedule.StartDate = *patch.StartDate
	}
	if patch.EndDate != nil {
		schedule.EndDate = *patch.EndDate
	}
	if patch.WorkingDays != nil {
		schedule.WorkingDays = *patch.WorkingDays
	}
	if patch.WorkStartTime != nil {
		schedule.WorkStartTime = *patch.WorkStartTime
	}
	if patch.WorkEndTime != nil {
		schedule.WorkEndTime = *patch.WorkEndTime
	}
	if patch.Status != nil {
		schedule.Status = *patch.Status
	}
	if schedule.EndDate.Before(schedule.StartDate) {
		return models.Schedule{}, fmt.Errorf("%w: invalid schedule date range", repository.ErrValidation)
	}
	if err := s.store.UpdateSchedule(ctx, &schedule); err != nil {
		return models.Schedule{}, err
	}
	return schedule, nil
}

func (s *Service) CreateTimeSlot(ctx context.Context, slot models.TimeSlot) (models.TimeSlot, error) {
	if slot.ScheduleID == uuid.Nil {
		return models.TimeSlot{}, fmt.Errorf("%w: schedule_id is required", repository.ErrValidation)
	}
	if !slot.EndAt.After(slot.StartAt) {
		return models.TimeSlot{}, fmt.Errorf("%w: end_at must be after start_at", repository.ErrValidation)
	}

	schedule, err := s.store.GetSchedule(ctx, slot.ScheduleID)
	if err != nil {
		return models.TimeSlot{}, err
	}
	slot.DoctorID = schedule.DoctorID
	if slot.Status == "" {
		slot.Status = models.SlotStatusFree
	}
	if slot.IsBlocked {
		slot.Status = models.SlotStatusBlocked
	}

	overlap, err := s.store.HasDoctorSlotOverlap(ctx, slot.DoctorID, slot.StartAt, slot.EndAt, nil)
	if err != nil {
		return models.TimeSlot{}, err
	}
	if overlap {
		return models.TimeSlot{}, fmt.Errorf("%w: time slot overlaps existing slot", repository.ErrConflict)
	}

	if err := s.store.CreateTimeSlot(ctx, &slot); err != nil {
		return models.TimeSlot{}, err
	}
	return slot, nil
}

func (s *Service) UpdateTimeSlot(ctx context.Context, id uuid.UUID, patch TimeSlotPatch) (models.TimeSlot, error) {
	slot, err := s.store.GetTimeSlot(ctx, id)
	if err != nil {
		return models.TimeSlot{}, err
	}
	if patch.StartAt != nil {
		slot.StartAt = *patch.StartAt
	}
	if patch.EndAt != nil {
		slot.EndAt = *patch.EndAt
	}
	if patch.Status != nil {
		slot.Status = *patch.Status
	}
	if patch.IsBlocked != nil {
		slot.IsBlocked = *patch.IsBlocked
		if slot.IsBlocked {
			slot.Status = models.SlotStatusBlocked
		}
	}
	if !slot.EndAt.After(slot.StartAt) {
		return models.TimeSlot{}, fmt.Errorf("%w: end_at must be after start_at", repository.ErrValidation)
	}

	overlap, err := s.store.HasDoctorSlotOverlap(ctx, slot.DoctorID, slot.StartAt, slot.EndAt, &id)
	if err != nil {
		return models.TimeSlot{}, err
	}
	if overlap {
		return models.TimeSlot{}, fmt.Errorf("%w: time slot overlaps existing slot", repository.ErrConflict)
	}

	if err := s.store.UpdateTimeSlot(ctx, &slot); err != nil {
		return models.TimeSlot{}, err
	}
	return slot, nil
}

func (s *Service) BlockTimeSlot(ctx context.Context, id uuid.UUID, blocked bool) (models.TimeSlot, error) {
	status := models.SlotStatusFree
	if blocked {
		status = models.SlotStatusBlocked
	}
	return s.UpdateTimeSlot(ctx, id, TimeSlotPatch{
		Status:    &status,
		IsBlocked: &blocked,
	})
}
