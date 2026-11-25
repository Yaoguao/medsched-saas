package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"base/internal/domain/models"
	"base/internal/integration/notifier"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreateNotification(ctx context.Context, notification models.Notification) (models.Notification, error) {
	if notification.AppointmentID == uuid.Nil || notification.PatientID == uuid.Nil {
		return models.Notification{}, fmt.Errorf("%w: appointment_id and patient_id are required", repository.ErrValidation)
	}
	if notification.Status == "" {
		notification.Status = models.NotificationStatusPending
	}
	if notification.ScheduledAt.IsZero() {
		notification.ScheduledAt = s.now()
	}
	if err := s.store.CreateNotification(ctx, &notification); err != nil {
		return models.Notification{}, err
	}
	s.dispatchNotification(ctx, notification, models.Appointment{}, models.Patient{})
	return notification, nil
}

func (s *Service) scheduleAppointmentNotifications(ctx context.Context, appointment models.Appointment) {
	if appointment.ID == uuid.Nil || appointment.PatientID == uuid.Nil {
		return
	}

	patient := appointment.Patient
	if patient.ID == uuid.Nil {
		loaded, err := s.store.GetPatient(ctx, appointment.PatientID)
		if err != nil {
			return
		}
		patient = loaded
	}

	plans := []notificationPlan{
		{
			key:         "confirmation-email",
			channel:     models.NotificationChannelEmail,
			kind:        models.NotificationTypeConfirmation,
			scheduledAt: s.now(),
			provider:    "core:auto-default",
		},
		{
			key:         "reminder-email-24h",
			channel:     models.NotificationChannelEmail,
			kind:        models.NotificationTypeReminder,
			scheduledAt: appointment.AppointmentAt.Add(-24 * time.Hour),
			provider:    "core:auto-default",
		},
		{
			key:         "reminder-email-2h",
			channel:     models.NotificationChannelEmail,
			kind:        models.NotificationTypeReminder,
			scheduledAt: appointment.AppointmentAt.Add(-2 * time.Hour),
			provider:    "core:auto-default",
		},
	}
	s.createNotificationPlans(ctx, appointment, patient, plans)
}

func (s *Service) scheduleHighRiskNotifications(ctx context.Context, appointment models.Appointment) {
	if appointment.ID == uuid.Nil || appointment.PatientID == uuid.Nil {
		return
	}

	patient := appointment.Patient
	if patient.ID == uuid.Nil {
		loaded, err := s.store.GetPatient(ctx, appointment.PatientID)
		if err != nil {
			return
		}
		patient = loaded
	}

	plans := []notificationPlan{
		{
			key:         "high-risk-sms-24h",
			channel:     models.NotificationChannelSMS,
			kind:        models.NotificationTypeReminder,
			scheduledAt: appointment.AppointmentAt.Add(-24 * time.Hour),
			provider:    "core:high-risk",
		},
		{
			key:         "high-risk-sms-6h",
			channel:     models.NotificationChannelSMS,
			kind:        models.NotificationTypeReminder,
			scheduledAt: appointment.AppointmentAt.Add(-6 * time.Hour),
			provider:    "core:high-risk",
		},
		{
			key:         "high-risk-sms-2h",
			channel:     models.NotificationChannelSMS,
			kind:        models.NotificationTypeReminder,
			scheduledAt: appointment.AppointmentAt.Add(-2 * time.Hour),
			provider:    "core:high-risk",
		},
		{
			key:         "high-risk-email-now",
			channel:     models.NotificationChannelEmail,
			kind:        models.NotificationTypeReminder,
			scheduledAt: s.now(),
			provider:    "core:high-risk",
		},
	}
	s.createNotificationPlans(ctx, appointment, patient, plans)
}

func (s *Service) scheduleCancellationNotification(ctx context.Context, appointment models.Appointment) {
	if appointment.ID == uuid.Nil || appointment.PatientID == uuid.Nil {
		return
	}

	_ = s.store.MarkPendingAppointmentNotificationsFailed(ctx, appointment.ID, "appointment cancelled before scheduled send", s.now())

	patient := appointment.Patient
	if patient.ID == uuid.Nil {
		loaded, err := s.store.GetPatient(ctx, appointment.PatientID)
		if err != nil {
			return
		}
		patient = loaded
	}

	s.createNotificationPlans(ctx, appointment, patient, []notificationPlan{
		{
			key:         "cancellation-email",
			channel:     models.NotificationChannelEmail,
			kind:        models.NotificationTypeCancellation,
			scheduledAt: s.now(),
			provider:    "core:auto-default",
		},
	})
}

func (s *Service) createNotificationPlans(ctx context.Context, appointment models.Appointment, patient models.Patient, plans []notificationPlan) {
	for _, plan := range plans {
		if !s.notificationChannelAvailable(patient, plan.channel) {
			continue
		}

		scheduledAt := plan.scheduledAt
		if scheduledAt.Before(s.now()) {
			scheduledAt = s.now()
		}

		notification := models.Notification{
			AppointmentID: appointment.ID,
			PatientID:     appointment.PatientID,
			Channel:       plan.channel,
			Type:          plan.kind,
			Status:        models.NotificationStatusPending,
			ScheduledAt:   scheduledAt,
			Provider:      plan.provider,
			ExternalID:    notificationExternalID(appointment.ID, plan.key),
		}
		created, _ := s.store.CreateNotificationIfMissing(ctx, &notification)
		if created {
			s.dispatchNotification(ctx, notification, appointment, patient)
		}
	}
}

func (s *Service) dispatchNotification(ctx context.Context, notification models.Notification, appointment models.Appointment, patient models.Patient) {
	if s.notifierClient == nil || notification.Status != models.NotificationStatusPending {
		return
	}

	if appointment.ID == uuid.Nil {
		loaded, err := s.store.GetAppointment(ctx, notification.AppointmentID)
		if err == nil {
			appointment = loaded
		}
	}
	if patient.ID == uuid.Nil {
		loaded, err := s.store.GetPatient(ctx, notification.PatientID)
		if err == nil {
			patient = loaded
		}
	}

	channel, recipient := notifierChannelAndRecipient(notification.Channel, patient)
	if channel == "" || recipient == "" {
		return
	}

	subject, message := notificationMessage(notification, appointment, patient)
	if message == "" {
		return
	}

	_ = s.notifierClient.Schedule(ctx, notifier.ScheduleNotificationRequest{
		ID:            notification.ID,
		ExternalID:    notification.ExternalID,
		AppointmentID: notification.AppointmentID,
		PatientID:     notification.PatientID,
		Channel:       channel,
		Recipient:     recipient,
		Subject:       subject,
		Message:       message,
		SendAt:        notification.ScheduledAt,
		Variables: map[string]string{
			"patient_name":   patient.FullName,
			"doctor_name":    appointment.Doctor.FullName,
			"appointment_at": appointment.AppointmentAt.Format("02.01.2006 15:04"),
			"cabinet":        appointment.Doctor.Cabinet,
		},
	})
}

func notifierChannelAndRecipient(channel models.NotificationChannel, patient models.Patient) (string, string) {
	switch channel {
	case models.NotificationChannelEmail:
		return "email", strings.TrimSpace(patient.Email)
	case models.NotificationChannelSMS:
		return "sms", strings.TrimSpace(patient.Phone)
	default:
		return "", ""
	}
}

func notificationMessage(notification models.Notification, appointment models.Appointment, patient models.Patient) (string, string) {
	patientName := patient.FullName
	if patientName == "" {
		patientName = "пациент"
	}
	doctorName := appointment.Doctor.FullName
	if doctorName == "" {
		doctorName = "врач"
	}
	appointmentAt := appointment.AppointmentAt.Format("02.01.2006 15:04")
	cabinet := strings.TrimSpace(appointment.Doctor.Cabinet)
	if cabinet != "" {
		cabinet = ", кабинет " + cabinet
	}

	switch notification.Type {
	case models.NotificationTypeConfirmation:
		return "Запись создана", fmt.Sprintf("Здравствуйте, %s. Ваша запись на %s к врачу %s%s создана.", patientName, appointmentAt, doctorName, cabinet)
	case models.NotificationTypeCancellation:
		return "Запись отменена", fmt.Sprintf("Здравствуйте, %s. Ваша запись на %s к врачу %s отменена.", patientName, appointmentAt, doctorName)
	case models.NotificationTypeReminder:
		if strings.Contains(notification.Provider, "high-risk") {
			return "Напоминание о приеме", fmt.Sprintf("Здравствуйте, %s. Напоминаем о приеме %s у врача %s%s. Если не сможете прийти, пожалуйста, свяжитесь с клиникой заранее.", patientName, appointmentAt, doctorName, cabinet)
		}
		return "Напоминание о приеме", fmt.Sprintf("Здравствуйте, %s. Напоминаем о приеме %s у врача %s%s.", patientName, appointmentAt, doctorName, cabinet)
	default:
		return "Уведомление MedSched", fmt.Sprintf("Здравствуйте, %s. У вас есть обновление по записи на %s.", patientName, appointmentAt)
	}
}

func (s *Service) notificationChannelAvailable(patient models.Patient, channel models.NotificationChannel) bool {
	switch channel {
	case models.NotificationChannelEmail:
		return strings.TrimSpace(patient.Email) != ""
	case models.NotificationChannelSMS:
		return strings.TrimSpace(patient.Phone) != ""
	case models.NotificationChannelPush:
		return true
	default:
		return false
	}
}

type notificationPlan struct {
	key         string
	channel     models.NotificationChannel
	kind        models.NotificationType
	scheduledAt time.Time
	provider    string
}

func notificationExternalID(appointmentID uuid.UUID, key string) string {
	return fmt.Sprintf("%s:%s", appointmentID.String(), key)
}

func (s *Service) GetNotification(ctx context.Context, id uuid.UUID) (models.Notification, error) {
	return s.store.GetNotification(ctx, id)
}

func (s *Service) ListAppointmentNotifications(ctx context.Context, appointmentID uuid.UUID, page pagination.Page) ([]models.Notification, error) {
	return s.store.ListAppointmentNotifications(ctx, appointmentID, page)
}

func (s *Service) UpdateNotificationDeliveryStatus(ctx context.Context, id uuid.UUID, status models.NotificationStatus, sentAt *time.Time, errorMessage string) (models.Notification, error) {
	if id == uuid.Nil {
		return models.Notification{}, fmt.Errorf("%w: notification_id is required", repository.ErrValidation)
	}
	switch status {
	case models.NotificationStatusSent, models.NotificationStatusFailed, models.NotificationStatusDelivered:
	default:
		return models.Notification{}, fmt.Errorf("%w: unsupported notification status", repository.ErrValidation)
	}
	if status == models.NotificationStatusFailed && strings.TrimSpace(errorMessage) == "" {
		errorMessage = "delivery failed"
	}
	if (status == models.NotificationStatusSent || status == models.NotificationStatusDelivered) && sentAt == nil {
		now := s.now()
		sentAt = &now
	}
	if status == models.NotificationStatusSent || status == models.NotificationStatusDelivered {
		errorMessage = ""
	}
	return s.store.UpdateNotificationDeliveryStatus(ctx, id, status, sentAt, strings.TrimSpace(errorMessage))
}
