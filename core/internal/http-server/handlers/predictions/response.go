package predictions

import (
	"time"

	"base/internal/domain/models"
	"base/internal/service"

	"github.com/google/uuid"
)

type predictionResponse struct {
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

type trainingDatasetResponse struct {
	Items []trainingRowResponse `json:"items"`
	Count int                   `json:"count"`
}

type modelInfoResponse struct {
	Loaded       bool           `json:"loaded"`
	ModelVersion string         `json:"model_version"`
	Algorithm    string         `json:"algorithm"`
	ArtifactPath string         `json:"artifact_path"`
	Metrics      map[string]any `json:"metrics"`
}

type trainingRowResponse struct {
	PatientID                 uuid.UUID `json:"patient_id"`
	AppointmentAt             time.Time `json:"appointment_at"`
	Age                       int       `json:"age"`
	Gender                    string    `json:"gender"`
	Neighbourhood             string    `json:"neighbourhood"`
	Scholarship               bool      `json:"scholarship"`
	HasHypertension           bool      `json:"has_hypertension"`
	HasDiabetes               bool      `json:"has_diabetes"`
	HasAlcoholism             bool      `json:"has_alcoholism"`
	HasHandicap               bool      `json:"has_handicap"`
	SMSReceived               bool      `json:"sms_received"`
	NotificationCount         int       `json:"notification_count"`
	SentNotifications         int       `json:"sent_notifications"`
	FailedNotifications       int       `json:"failed_notifications"`
	PendingNotifications      int       `json:"pending_notifications"`
	LeadTimeDays              int       `json:"lead_time_days"`
	AppointmentWeekday        int       `json:"appointment_weekday"`
	AppointmentHour           int       `json:"appointment_hour"`
	PreviousAppointmentsCount int       `json:"previous_appointments_count"`
	PreviousNoShowCount       int       `json:"previous_no_show_count"`
	PreviousNoShowRate        float64   `json:"previous_no_show_rate"`
	BookingChannel            string    `json:"booking_channel"`
	VisitType                 string    `json:"visit_type"`
	DoctorSpecialty           string    `json:"doctor_specialty"`
	NoShow                    bool      `json:"no_show"`
}

func responseFromModel(prediction models.PredictionResult) predictionResponse {
	return predictionResponse{
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

func trainingDatasetResponseFromRows(rows []service.NoShowTrainingRow) trainingDatasetResponse {
	responses := make([]trainingRowResponse, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, trainingRowResponse{
			PatientID:                 row.PatientID,
			AppointmentAt:             row.AppointmentAt,
			Age:                       row.Age,
			Gender:                    row.Gender,
			Neighbourhood:             row.Neighbourhood,
			Scholarship:               row.Scholarship,
			HasHypertension:           row.HasHypertension,
			HasDiabetes:               row.HasDiabetes,
			HasAlcoholism:             row.HasAlcoholism,
			HasHandicap:               row.HasHandicap,
			SMSReceived:               row.SMSReceived,
			NotificationCount:         row.NotificationCount,
			SentNotifications:         row.SentNotifications,
			FailedNotifications:       row.FailedNotifications,
			PendingNotifications:      row.PendingNotifications,
			LeadTimeDays:              row.LeadTimeDays,
			AppointmentWeekday:        row.AppointmentWeekday,
			AppointmentHour:           row.AppointmentHour,
			PreviousAppointmentsCount: row.PreviousAppointmentsCount,
			PreviousNoShowCount:       row.PreviousNoShowCount,
			PreviousNoShowRate:        row.PreviousNoShowRate,
			BookingChannel:            row.BookingChannel,
			VisitType:                 row.VisitType,
			DoctorSpecialty:           row.DoctorSpecialty,
			NoShow:                    row.NoShow,
		})
	}
	return trainingDatasetResponse{
		Items: responses,
		Count: len(responses),
	}
}

func modelInfoResponseFromService(info service.MLModelInfo) modelInfoResponse {
	return modelInfoResponse{
		Loaded:       info.Loaded,
		ModelVersion: info.ModelVersion,
		Algorithm:    info.Algorithm,
		ArtifactPath: info.ArtifactPath,
		Metrics:      info.Metrics,
	}
}
