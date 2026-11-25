package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"base/internal/domain/models"
	"base/internal/integration/ml"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreatePredictionResult(ctx context.Context, prediction models.PredictionResult) (models.PredictionResult, error) {
	if prediction.AppointmentID == uuid.Nil {
		return models.PredictionResult{}, fmt.Errorf("%w: appointment_id is required", repository.ErrValidation)
	}
	if prediction.Probability < 0 || prediction.Probability > 1 {
		return models.PredictionResult{}, fmt.Errorf("%w: probability must be between 0 and 1", repository.ErrValidation)
	}
	if prediction.RequestedAt.IsZero() {
		prediction.RequestedAt = s.now()
	}
	if prediction.ReceivedAt.IsZero() {
		prediction.ReceivedAt = s.now()
	}
	if prediction.RiskLevel == "" {
		prediction.RiskLevel = riskLevel(prediction.Probability)
	}
	if err := s.store.UpsertPredictionResult(ctx, &prediction); err != nil {
		return models.PredictionResult{}, err
	}
	return s.store.GetPredictionResultByAppointment(ctx, prediction.AppointmentID)
}

func (s *Service) GetPredictionResult(ctx context.Context, id uuid.UUID) (models.PredictionResult, error) {
	return s.store.GetPredictionResult(ctx, id)
}

func (s *Service) MLModelInfo(ctx context.Context) (MLModelInfo, error) {
	if s.mlClient == nil {
		return MLModelInfo{}, fmt.Errorf("%w: ml client is not configured", repository.ErrValidation)
	}
	info, err := s.mlClient.ModelInfo(ctx)
	if err != nil {
		return MLModelInfo{}, err
	}
	return MLModelInfo{
		Loaded:       info.Loaded,
		ModelVersion: info.ModelVersion,
		Algorithm:    info.Algorithm,
		ArtifactPath: info.ArtifactPath,
		Metrics:      info.Metrics,
	}, nil
}

func (s *Service) PredictAppointmentNoShow(ctx context.Context, appointmentID uuid.UUID) (models.PredictionResult, error) {
	if s.mlClient == nil {
		return models.PredictionResult{}, fmt.Errorf("%w: ml client is not configured", repository.ErrValidation)
	}
	return s.predictAppointmentNoShow(ctx, appointmentID)
}

func (s *Service) tryPredictAppointmentNoShow(ctx context.Context, appointmentID uuid.UUID) {
	if s.mlClient == nil || appointmentID == uuid.Nil {
		return
	}
	_, _ = s.predictAppointmentNoShow(ctx, appointmentID)
}

func (s *Service) predictAppointmentNoShow(ctx context.Context, appointmentID uuid.UUID) (models.PredictionResult, error) {
	if appointmentID == uuid.Nil {
		return models.PredictionResult{}, fmt.Errorf("%w: appointment_id is required", repository.ErrValidation)
	}

	appointment, err := s.store.GetAppointment(ctx, appointmentID)
	if err != nil {
		return models.PredictionResult{}, err
	}
	previousAppointments, previousNoShows, err := s.store.PatientAttendanceHistory(
		ctx,
		appointment.PatientID,
		appointment.AppointmentAt,
		appointment.ID,
	)
	if err != nil {
		return models.PredictionResult{}, err
	}

	requestedAt := s.now()
	response, err := s.mlClient.PredictNoShow(
		ctx,
		noShowPredictionRequest(appointment, previousAppointments, previousNoShows),
	)
	if err != nil {
		return models.PredictionResult{}, err
	}

	prediction, err := s.CreatePredictionResult(ctx, models.PredictionResult{
		AppointmentID: appointmentID,
		Probability:   response.NoShowProbability,
		RiskLevel:     predictionRiskLevel(response.RiskLevel, response.NoShowProbability),
		ModelVersion:  response.ModelVersion,
		RequestedAt:   requestedAt,
		ReceivedAt:    s.now(),
	})
	if err != nil {
		return models.PredictionResult{}, err
	}
	if prediction.RiskLevel == models.PredictionRiskHigh {
		s.scheduleHighRiskNotifications(ctx, appointment)
	}
	return prediction, nil
}

func (s *Service) ExportNoShowTrainingDataset(ctx context.Context, limit int) ([]NoShowTrainingRow, error) {
	appointments, err := s.store.ListCompletedAppointmentsForTraining(ctx, limit)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(appointments, func(i, j int) bool {
		return appointments[i].AppointmentAt.Before(appointments[j].AppointmentAt)
	})

	history := make(map[uuid.UUID]patientHistory)
	rows := make([]NoShowTrainingRow, 0, len(appointments))
	for _, appointment := range appointments {
		if appointment.ShowedUp == nil {
			continue
		}
		patientHistory := history[appointment.PatientID]
		rows = append(rows, noShowTrainingRow(appointment, patientHistory))

		patientHistory.appointmentsCount++
		if !*appointment.ShowedUp {
			patientHistory.noShowCount++
		}
		history[appointment.PatientID] = patientHistory
	}
	return rows, nil
}

func noShowPredictionRequest(
	appointment models.Appointment,
	previousAppointments int,
	previousNoShows int,
) ml.NoShowPredictionRequest {
	profile := appointment.Patient.MedicalProfile
	notifications := notificationStatsFromAppointment(appointment)
	patientFeatures := ml.PatientFeatures{
		Age:           appointment.Patient.Age,
		Gender:        valueOrUnknown(appointment.Patient.Gender),
		Neighbourhood: valueOrUnknown(appointment.Patient.Neighbourhood),
		Scholarship:   appointment.Patient.Scholarship,
	}
	if profile != nil {
		patientFeatures.HasHypertension = profile.HasHypertension
		patientFeatures.HasDiabetes = profile.HasDiabetes
		patientFeatures.HasAlcoholism = profile.HasAlcoholism
		patientFeatures.HasHandicap = profile.HasHandicap
	}

	return ml.NoShowPredictionRequest{
		Patient: patientFeatures,
		Appointment: ml.AppointmentFeatures{
			ScheduledAt:          appointment.ScheduledAt,
			AppointmentAt:        appointment.AppointmentAt,
			BookingChannel:       string(appointment.BookingChannel),
			VisitType:            string(appointment.VisitType),
			DoctorSpecialty:      valueOrUnknown(appointment.Doctor.Specialty),
			SMSReceived:          appointment.ReminderSent,
			NotificationCount:    notifications.total,
			SentNotifications:    notifications.sent,
			FailedNotifications:  notifications.failed,
			PendingNotifications: notifications.pending,
		},
		History: ml.HistoryFeatures{
			PreviousAppointmentsCount: previousAppointments,
			PreviousNoShowCount:       previousNoShows,
		},
	}
}

func noShowTrainingRow(appointment models.Appointment, history patientHistory) NoShowTrainingRow {
	profile := appointment.Patient.MedicalProfile
	notifications := notificationStatsFromAppointment(appointment)
	row := NoShowTrainingRow{
		PatientID:                 appointment.PatientID,
		AppointmentAt:             appointment.AppointmentAt,
		Age:                       appointment.Patient.Age,
		Gender:                    valueOrUnknown(appointment.Patient.Gender),
		Neighbourhood:             valueOrUnknown(appointment.Patient.Neighbourhood),
		Scholarship:               appointment.Patient.Scholarship,
		SMSReceived:               appointment.ReminderSent,
		NotificationCount:         notifications.total,
		SentNotifications:         notifications.sent,
		FailedNotifications:       notifications.failed,
		PendingNotifications:      notifications.pending,
		LeadTimeDays:              leadTimeDays(appointment),
		AppointmentWeekday:        int(appointment.AppointmentAt.Weekday()),
		AppointmentHour:           appointment.AppointmentAt.Hour(),
		PreviousAppointmentsCount: history.appointmentsCount,
		PreviousNoShowCount:       history.noShowCount,
		PreviousNoShowRate:        history.noShowRate(),
		BookingChannel:            string(appointment.BookingChannel),
		VisitType:                 string(appointment.VisitType),
		DoctorSpecialty:           valueOrUnknown(appointment.Doctor.Specialty),
		NoShow:                    appointment.ShowedUp != nil && !*appointment.ShowedUp,
	}
	if profile != nil {
		row.HasHypertension = profile.HasHypertension
		row.HasDiabetes = profile.HasDiabetes
		row.HasAlcoholism = profile.HasAlcoholism
		row.HasHandicap = profile.HasHandicap
	}
	return row
}

type notificationStats struct {
	total   int
	sent    int
	failed  int
	pending int
}

func notificationStatsFromAppointment(appointment models.Appointment) notificationStats {
	stats := notificationStats{total: len(appointment.Notifications)}
	for _, notification := range appointment.Notifications {
		switch notification.Status {
		case models.NotificationStatusSent, models.NotificationStatusDelivered:
			stats.sent++
		case models.NotificationStatusFailed:
			stats.failed++
		case models.NotificationStatusPending:
			stats.pending++
		}
	}
	return stats
}

func leadTimeDays(appointment models.Appointment) int {
	days := int(appointment.AppointmentAt.Sub(appointment.ScheduledAt).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

type patientHistory struct {
	appointmentsCount int
	noShowCount       int
}

func (h patientHistory) noShowRate() float64 {
	if h.appointmentsCount == 0 {
		return 0
	}
	return float64(h.noShowCount) / float64(h.appointmentsCount)
}

func predictionRiskLevel(value string, probability float64) models.PredictionRiskLevel {
	switch models.PredictionRiskLevel(strings.ToUpper(strings.TrimSpace(value))) {
	case models.PredictionRiskLow:
		return models.PredictionRiskLow
	case models.PredictionRiskMedium:
		return models.PredictionRiskMedium
	case models.PredictionRiskHigh:
		return models.PredictionRiskHigh
	default:
		return riskLevel(probability)
	}
}

func valueOrUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "UNKNOWN"
	}
	return value
}

func riskLevel(probability float64) models.PredictionRiskLevel {
	switch {
	case probability >= 0.7:
		return models.PredictionRiskHigh
	case probability >= 0.4:
		return models.PredictionRiskMedium
	default:
		return models.PredictionRiskLow
	}
}
