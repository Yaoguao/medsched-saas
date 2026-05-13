package schedules

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type scheduleResponse struct {
	ID            uuid.UUID          `json:"id"`
	DoctorID      uuid.UUID          `json:"doctor_id"`
	StartDate     time.Time          `json:"start_date"`
	EndDate       time.Time          `json:"end_date"`
	WorkingDays   string             `json:"working_days"`
	WorkStartTime string             `json:"work_start_time"`
	WorkEndTime   string             `json:"work_end_time"`
	Status        string             `json:"status"`
	TimeSlots     []timeSlotResponse `json:"time_slots,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type timeSlotResponse struct {
	ID         uuid.UUID         `json:"id"`
	ScheduleID uuid.UUID         `json:"schedule_id"`
	DoctorID   uuid.UUID         `json:"doctor_id"`
	StartAt    time.Time         `json:"start_at"`
	EndAt      time.Time         `json:"end_at"`
	Status     models.SlotStatus `json:"status"`
	IsBlocked  bool              `json:"is_blocked"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func scheduleResponsesFromModels(schedules []models.Schedule) []scheduleResponse {
	responses := make([]scheduleResponse, 0, len(schedules))
	for _, schedule := range schedules {
		responses = append(responses, scheduleResponseFromModel(schedule))
	}
	return responses
}

func scheduleResponseFromModel(schedule models.Schedule) scheduleResponse {
	return scheduleResponse{
		ID:            schedule.ID,
		DoctorID:      schedule.DoctorID,
		StartDate:     schedule.StartDate,
		EndDate:       schedule.EndDate,
		WorkingDays:   schedule.WorkingDays,
		WorkStartTime: schedule.WorkStartTime,
		WorkEndTime:   schedule.WorkEndTime,
		Status:        schedule.Status,
		TimeSlots:     timeSlotResponsesFromModels(schedule.TimeSlots),
		CreatedAt:     schedule.CreatedAt,
		UpdatedAt:     schedule.UpdatedAt,
	}
}

func timeSlotResponsesFromModels(slots []models.TimeSlot) []timeSlotResponse {
	if len(slots) == 0 {
		return nil
	}
	responses := make([]timeSlotResponse, 0, len(slots))
	for _, slot := range slots {
		responses = append(responses, timeSlotResponseFromModel(slot))
	}
	return responses
}

func timeSlotResponseFromModel(slot models.TimeSlot) timeSlotResponse {
	return timeSlotResponse{
		ID:         slot.ID,
		ScheduleID: slot.ScheduleID,
		DoctorID:   slot.DoctorID,
		StartAt:    slot.StartAt,
		EndAt:      slot.EndAt,
		Status:     slot.Status,
		IsBlocked:  slot.IsBlocked,
		CreatedAt:  slot.CreatedAt,
		UpdatedAt:  slot.UpdatedAt,
	}
}
