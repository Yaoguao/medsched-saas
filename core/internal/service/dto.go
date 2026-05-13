package service

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type PatientPatch struct {
	FullName      *string
	Gender        *string
	BirthDate     *time.Time
	Age           *int
	Phone         *string
	Email         *string
	Neighbourhood *string
	Scholarship   *bool
}

type SchedulePatch struct {
	StartDate     *time.Time
	EndDate       *time.Time
	WorkingDays   *string
	WorkStartTime *string
	WorkEndTime   *string
	Status        *string
}

type TimeSlotPatch struct {
	StartAt   *time.Time
	EndAt     *time.Time
	Status    *models.SlotStatus
	IsBlocked *bool
}

type CreateAppointmentRequest struct {
	PatientID      uuid.UUID
	TimeSlotID     uuid.UUID
	BookingChannel models.BookingChannel
	VisitType      models.VisitType
	ChangedBy      *uuid.UUID
}
