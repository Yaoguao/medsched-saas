package appointments

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type appointmentResponse struct {
	ID             uuid.UUID                       `json:"id"`
	PatientID      uuid.UUID                       `json:"patient_id"`
	DoctorID       uuid.UUID                       `json:"doctor_id"`
	TimeSlotID     uuid.UUID                       `json:"time_slot_id"`
	ScheduledAt    time.Time                       `json:"scheduled_at"`
	AppointmentAt  time.Time                       `json:"appointment_at"`
	Status         models.AppointmentStatus        `json:"status"`
	BookingChannel models.BookingChannel           `json:"booking_channel"`
	VisitType      models.VisitType                `json:"visit_type"`
	ReminderSent   bool                            `json:"reminder_sent"`
	ShowedUp       *bool                           `json:"showed_up,omitempty"`
	ConfirmedAt    *time.Time                      `json:"confirmed_at,omitempty"`
	CancelledAt    *time.Time                      `json:"cancelled_at,omitempty"`
	CompletedAt    *time.Time                      `json:"completed_at,omitempty"`
	StatusHistory  []appointmentStatusHistoryEntry `json:"status_history,omitempty"`
	CreatedAt      time.Time                       `json:"created_at"`
	UpdatedAt      time.Time                       `json:"updated_at"`
}

type appointmentStatusHistoryEntry struct {
	ID            uuid.UUID                `json:"id"`
	AppointmentID uuid.UUID                `json:"appointment_id"`
	OldStatus     models.AppointmentStatus `json:"old_status,omitempty"`
	NewStatus     models.AppointmentStatus `json:"new_status"`
	ChangedBy     *uuid.UUID               `json:"changed_by,omitempty"`
	Reason        string                   `json:"reason"`
	ChangedAt     time.Time                `json:"changed_at"`
}

func ResponsesFromModels(appointments []models.Appointment) []appointmentResponse {
	responses := make([]appointmentResponse, 0, len(appointments))
	for _, appointment := range appointments {
		responses = append(responses, ResponseFromModel(appointment))
	}
	return responses
}

func ResponseFromModel(appointment models.Appointment) appointmentResponse {
	return appointmentResponse{
		ID:             appointment.ID,
		PatientID:      appointment.PatientID,
		DoctorID:       appointment.DoctorID,
		TimeSlotID:     appointment.TimeSlotID,
		ScheduledAt:    appointment.ScheduledAt,
		AppointmentAt:  appointment.AppointmentAt,
		Status:         appointment.Status,
		BookingChannel: appointment.BookingChannel,
		VisitType:      appointment.VisitType,
		ReminderSent:   appointment.ReminderSent,
		ShowedUp:       appointment.ShowedUp,
		ConfirmedAt:    appointment.ConfirmedAt,
		CancelledAt:    appointment.CancelledAt,
		CompletedAt:    appointment.CompletedAt,
		StatusHistory:  statusHistoryFromModels(appointment.StatusHistory),
		CreatedAt:      appointment.CreatedAt,
		UpdatedAt:      appointment.UpdatedAt,
	}
}

func statusHistoryFromModels(history []models.AppointmentStatusHistory) []appointmentStatusHistoryEntry {
	if len(history) == 0 {
		return nil
	}
	entries := make([]appointmentStatusHistoryEntry, 0, len(history))
	for _, item := range history {
		entries = append(entries, appointmentStatusHistoryEntry{
			ID:            item.ID,
			AppointmentID: item.AppointmentID,
			OldStatus:     item.OldStatus,
			NewStatus:     item.NewStatus,
			ChangedBy:     item.ChangedBy,
			Reason:        item.Reason,
			ChangedAt:     item.ChangedAt,
		})
	}
	return entries
}
