package appointments

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type appointmentResponse struct {
	ID               uuid.UUID                       `json:"id"`
	PatientID        uuid.UUID                       `json:"patient_id"`
	DoctorID         uuid.UUID                       `json:"doctor_id"`
	TimeSlotID       uuid.UUID                       `json:"time_slot_id"`
	Patient          *patientSummaryResponse         `json:"patient,omitempty"`
	Doctor           *doctorSummaryResponse          `json:"doctor,omitempty"`
	TimeSlot         *timeSlotSummaryResponse        `json:"time_slot,omitempty"`
	ScheduledAt      time.Time                       `json:"scheduled_at"`
	AppointmentAt    time.Time                       `json:"appointment_at"`
	Status           models.AppointmentStatus        `json:"status"`
	BookingChannel   models.BookingChannel           `json:"booking_channel"`
	VisitType        models.VisitType                `json:"visit_type"`
	ReminderSent     bool                            `json:"reminder_sent"`
	ShowedUp         *bool                           `json:"showed_up,omitempty"`
	ConfirmedAt      *time.Time                      `json:"confirmed_at,omitempty"`
	CancelledAt      *time.Time                      `json:"cancelled_at,omitempty"`
	CompletedAt      *time.Time                      `json:"completed_at,omitempty"`
	StatusHistory    []appointmentStatusHistoryEntry `json:"status_history,omitempty"`
	PredictionResult *predictionSummaryResponse      `json:"prediction_result,omitempty"`
	CreatedAt        time.Time                       `json:"created_at"`
	UpdatedAt        time.Time                       `json:"updated_at"`
}

type patientSummaryResponse struct {
	ID             uuid.UUID                      `json:"id"`
	UserID         *uuid.UUID                     `json:"user_id,omitempty"`
	FullName       string                         `json:"full_name"`
	Gender         string                         `json:"gender"`
	BirthDate      time.Time                      `json:"birth_date"`
	Age            int                            `json:"age"`
	Phone          string                         `json:"phone"`
	Email          string                         `json:"email"`
	Neighbourhood  string                         `json:"neighbourhood"`
	Scholarship    bool                           `json:"scholarship"`
	MedicalProfile *medicalProfileSummaryResponse `json:"medical_profile,omitempty"`
}

type medicalProfileSummaryResponse struct {
	ID                uuid.UUID `json:"id"`
	PatientID         uuid.UUID `json:"patient_id"`
	ChronicConditions string    `json:"chronic_conditions"`
	RiskFactors       string    `json:"risk_factors"`
	Allergies         string    `json:"allergies"`
	Notes             string    `json:"notes"`
	HasHypertension   bool      `json:"has_hypertension"`
	HasDiabetes       bool      `json:"has_diabetes"`
	HasAlcoholism     bool      `json:"has_alcoholism"`
	HasHandicap       bool      `json:"has_handicap"`
	UpdatedAtProfile  time.Time `json:"updated_at_profile"`
}

type doctorSummaryResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	FullName        string     `json:"full_name"`
	Specialty       string     `json:"specialty"`
	Cabinet         string     `json:"cabinet"`
	ExperienceYears int        `json:"experience_years"`
	Status          string     `json:"status"`
}

type timeSlotSummaryResponse struct {
	ID         uuid.UUID         `json:"id"`
	ScheduleID uuid.UUID         `json:"schedule_id"`
	DoctorID   uuid.UUID         `json:"doctor_id"`
	StartAt    time.Time         `json:"start_at"`
	EndAt      time.Time         `json:"end_at"`
	Status     models.SlotStatus `json:"status"`
	IsBlocked  bool              `json:"is_blocked"`
}

type rescheduleProposalResponse struct {
	ID                 uuid.UUID                       `json:"id"`
	AppointmentID      uuid.UUID                       `json:"appointment_id"`
	ProposedTimeSlotID uuid.UUID                       `json:"proposed_time_slot_id"`
	RequestedBy        *uuid.UUID                      `json:"requested_by,omitempty"`
	RespondedBy        *uuid.UUID                      `json:"responded_by,omitempty"`
	Reason             string                          `json:"reason"`
	ResponseReason     string                          `json:"response_reason,omitempty"`
	Status             models.RescheduleProposalStatus `json:"status"`
	RespondedAt        *time.Time                      `json:"responded_at,omitempty"`
	Appointment        appointmentResponse             `json:"appointment"`
	ProposedTimeSlot   *timeSlotSummaryResponse        `json:"proposed_time_slot,omitempty"`
	CreatedAt          time.Time                       `json:"created_at"`
	UpdatedAt          time.Time                       `json:"updated_at"`
}

type predictionSummaryResponse struct {
	ID            uuid.UUID                  `json:"id"`
	AppointmentID uuid.UUID                  `json:"appointment_id"`
	Probability   float64                    `json:"probability"`
	RiskLevel     models.PredictionRiskLevel `json:"risk_level"`
	ModelVersion  string                     `json:"model_version"`
	RequestedAt   time.Time                  `json:"requested_at"`
	ReceivedAt    time.Time                  `json:"received_at"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
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

func RescheduleProposalResponsesFromModels(proposals []models.AppointmentRescheduleProposal) []rescheduleProposalResponse {
	responses := make([]rescheduleProposalResponse, 0, len(proposals))
	for _, proposal := range proposals {
		responses = append(responses, RescheduleProposalResponseFromModel(proposal))
	}
	return responses
}

func RescheduleProposalResponseFromModel(proposal models.AppointmentRescheduleProposal) rescheduleProposalResponse {
	return rescheduleProposalResponse{
		ID:                 proposal.ID,
		AppointmentID:      proposal.AppointmentID,
		ProposedTimeSlotID: proposal.ProposedTimeSlotID,
		RequestedBy:        proposal.RequestedBy,
		RespondedBy:        proposal.RespondedBy,
		Reason:             proposal.Reason,
		ResponseReason:     proposal.ResponseReason,
		Status:             proposal.Status,
		RespondedAt:        proposal.RespondedAt,
		Appointment:        ResponseFromModel(proposal.Appointment),
		ProposedTimeSlot:   timeSlotSummaryFromModel(proposal.ProposedTimeSlot),
		CreatedAt:          proposal.CreatedAt,
		UpdatedAt:          proposal.UpdatedAt,
	}
}

func ResponseFromModel(appointment models.Appointment) appointmentResponse {
	return appointmentResponse{
		ID:               appointment.ID,
		PatientID:        appointment.PatientID,
		DoctorID:         appointment.DoctorID,
		TimeSlotID:       appointment.TimeSlotID,
		Patient:          patientSummaryFromModel(appointment.Patient),
		Doctor:           doctorSummaryFromModel(appointment.Doctor),
		TimeSlot:         timeSlotSummaryFromModel(appointment.TimeSlot),
		ScheduledAt:      appointment.ScheduledAt,
		AppointmentAt:    appointment.AppointmentAt,
		Status:           appointment.Status,
		BookingChannel:   appointment.BookingChannel,
		VisitType:        appointment.VisitType,
		ReminderSent:     appointment.ReminderSent,
		ShowedUp:         appointment.ShowedUp,
		ConfirmedAt:      appointment.ConfirmedAt,
		CancelledAt:      appointment.CancelledAt,
		CompletedAt:      appointment.CompletedAt,
		StatusHistory:    statusHistoryFromModels(appointment.StatusHistory),
		PredictionResult: predictionSummaryFromModel(appointment.PredictionResult),
		CreatedAt:        appointment.CreatedAt,
		UpdatedAt:        appointment.UpdatedAt,
	}
}

func patientSummaryFromModel(patient models.Patient) *patientSummaryResponse {
	if patient.ID == uuid.Nil {
		return nil
	}
	return &patientSummaryResponse{
		ID:             patient.ID,
		UserID:         patient.UserID,
		FullName:       patient.FullName,
		Gender:         patient.Gender,
		BirthDate:      patient.BirthDate,
		Age:            patient.Age,
		Phone:          patient.Phone,
		Email:          patient.Email,
		Neighbourhood:  patient.Neighbourhood,
		Scholarship:    patient.Scholarship,
		MedicalProfile: medicalProfileSummaryFromModel(patient.MedicalProfile),
	}
}

func medicalProfileSummaryFromModel(profile *models.PatientMedicalProfile) *medicalProfileSummaryResponse {
	if profile == nil {
		return nil
	}
	return &medicalProfileSummaryResponse{
		ID:                profile.ID,
		PatientID:         profile.PatientID,
		ChronicConditions: profile.ChronicConditions,
		RiskFactors:       profile.RiskFactors,
		Allergies:         profile.Allergies,
		Notes:             profile.Notes,
		HasHypertension:   profile.HasHypertension,
		HasDiabetes:       profile.HasDiabetes,
		HasAlcoholism:     profile.HasAlcoholism,
		HasHandicap:       profile.HasHandicap,
		UpdatedAtProfile:  profile.UpdatedAtProfile,
	}
}

func doctorSummaryFromModel(doctor models.Doctor) *doctorSummaryResponse {
	if doctor.ID == uuid.Nil {
		return nil
	}
	return &doctorSummaryResponse{
		ID:              doctor.ID,
		UserID:          doctor.UserID,
		FullName:        doctor.FullName,
		Specialty:       doctor.Specialty,
		Cabinet:         doctor.Cabinet,
		ExperienceYears: doctor.ExperienceYears,
		Status:          doctor.Status,
	}
}

func timeSlotSummaryFromModel(slot models.TimeSlot) *timeSlotSummaryResponse {
	if slot.ID == uuid.Nil {
		return nil
	}
	return &timeSlotSummaryResponse{
		ID:         slot.ID,
		ScheduleID: slot.ScheduleID,
		DoctorID:   slot.DoctorID,
		StartAt:    slot.StartAt,
		EndAt:      slot.EndAt,
		Status:     slot.Status,
		IsBlocked:  slot.IsBlocked,
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

func predictionSummaryFromModel(prediction *models.PredictionResult) *predictionSummaryResponse {
	if prediction == nil || prediction.ID == uuid.Nil {
		return nil
	}
	return &predictionSummaryResponse{
		ID:            prediction.ID,
		AppointmentID: prediction.AppointmentID,
		Probability:   prediction.Probability,
		RiskLevel:     prediction.RiskLevel,
		ModelVersion:  prediction.ModelVersion,
		RequestedAt:   prediction.RequestedAt,
		ReceivedAt:    prediction.ReceivedAt,
		CreatedAt:     prediction.CreatedAt,
		UpdatedAt:     prediction.UpdatedAt,
	}
}
