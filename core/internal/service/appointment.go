package service

import (
	"context"
	"fmt"
	"time"

	"base/internal/domain/models"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreateAppointment(ctx context.Context, req CreateAppointmentRequest) (models.Appointment, error) {
	var created models.Appointment
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		slot, err := tx.GetTimeSlotForUpdate(ctx, req.TimeSlotID)
		if err != nil {
			return err
		}
		if slot.IsBlocked || slot.Status != models.SlotStatusFree {
			return fmt.Errorf("%w: slot is not available", repository.ErrConflict)
		}
		if slot.StartAt.Before(s.now()) {
			return fmt.Errorf("%w: cannot book past slot", repository.ErrValidation)
		}
		if _, err := tx.GetPatient(ctx, req.PatientID); err != nil {
			return err
		}
		if _, err := tx.GetDoctor(ctx, slot.DoctorID); err != nil {
			return err
		}

		bookingChannel := req.BookingChannel
		if bookingChannel == "" {
			bookingChannel = models.BookingChannelWeb
		}
		visitType := req.VisitType
		if visitType == "" {
			visitType = models.VisitTypeConsultation
		}

		created = models.Appointment{
			PatientID:      req.PatientID,
			DoctorID:       slot.DoctorID,
			TimeSlotID:     slot.ID,
			ScheduledAt:    s.now(),
			AppointmentAt:  slot.StartAt,
			Status:         models.AppointmentStatusCreated,
			BookingChannel: bookingChannel,
			VisitType:      visitType,
		}
		if err := tx.CreateAppointment(ctx, &created); err != nil {
			return err
		}

		slot.Status = models.SlotStatusBooked
		if err := tx.UpdateTimeSlot(ctx, &slot); err != nil {
			return err
		}

		return tx.CreateAppointmentStatusHistory(ctx, &models.AppointmentStatusHistory{
			AppointmentID: created.ID,
			NewStatus:     models.AppointmentStatusCreated,
			ChangedBy:     req.ChangedBy,
			Reason:        "appointment created",
			ChangedAt:     s.now(),
		})
	})
	if err != nil {
		return models.Appointment{}, err
	}
	createdAppointment, err := s.store.GetAppointment(ctx, created.ID)
	if err != nil {
		return models.Appointment{}, err
	}
	s.scheduleAppointmentNotifications(ctx, createdAppointment)
	s.tryPredictAppointmentNoShow(ctx, created.ID)
	return s.store.GetAppointment(ctx, created.ID)
}

func (s *Service) GetAppointment(ctx context.Context, id uuid.UUID) (models.Appointment, error) {
	return s.store.GetAppointment(ctx, id)
}

func (s *Service) ListAppointments(ctx context.Context, page pagination.Page) ([]models.Appointment, error) {
	return s.store.ListAppointments(ctx, page)
}

func (s *Service) ListDoctorAppointments(ctx context.Context, doctorID uuid.UUID, page pagination.Page) ([]models.Appointment, error) {
	return s.store.ListDoctorAppointments(ctx, doctorID, page)
}

func (s *Service) MarkOverdueAppointmentsNoShow(ctx context.Context, grace time.Duration, limit int) (int, error) {
	if grace < 0 {
		grace = 0
	}
	if limit <= 0 {
		limit = 100
	}

	cutoff := s.now().Add(-grace)
	appointments, err := s.store.ListOverdueActiveAppointments(ctx, cutoff, limit)
	if err != nil {
		return 0, err
	}

	changed := 0
	for _, appointment := range appointments {
		_, err := s.ChangeAppointmentStatus(ctx, appointment.ID, models.AppointmentStatusNoShow, nil, "auto no-show: appointment time expired")
		if err != nil {
			if IsConflict(err) || IsValidation(err) || IsNotFound(err) {
				continue
			}
			return changed, err
		}
		changed++
	}
	return changed, nil
}

func (s *Service) ChangeAppointmentStatus(ctx context.Context, id uuid.UUID, newStatus models.AppointmentStatus, changedBy *uuid.UUID, reason string) (models.Appointment, error) {
	var updated models.Appointment
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		appointment, err := tx.GetAppointmentForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if appointment.AppointmentAt.Before(s.now()) && newStatus != models.AppointmentStatusCompleted && newStatus != models.AppointmentStatusNoShow {
			return fmt.Errorf("%w: past appointment cannot be changed to %s", repository.ErrValidation, newStatus)
		}
		if appointment.Status == models.AppointmentStatusCancelled || appointment.Status == models.AppointmentStatusCompleted || appointment.Status == models.AppointmentStatusNoShow {
			return fmt.Errorf("%w: terminal appointment status cannot be changed", repository.ErrConflict)
		}
		if !allowedStatusTransition(appointment.Status, newStatus) {
			return fmt.Errorf("%w: invalid status transition %s -> %s", repository.ErrValidation, appointment.Status, newStatus)
		}

		now := s.now()
		oldStatus := appointment.Status
		appointment.Status = newStatus
		switch newStatus {
		case models.AppointmentStatusConfirmed:
			appointment.ConfirmedAt = &now
		case models.AppointmentStatusCancelled:
			appointment.CancelledAt = &now
		case models.AppointmentStatusCompleted:
			appointment.CompletedAt = &now
			showedUp := true
			appointment.ShowedUp = &showedUp
		case models.AppointmentStatusNoShow:
			showedUp := false
			appointment.ShowedUp = &showedUp
		}

		if err := tx.UpdateAppointment(ctx, &appointment); err != nil {
			return err
		}
		if newStatus == models.AppointmentStatusCancelled {
			slot, err := tx.GetTimeSlotForUpdate(ctx, appointment.TimeSlotID)
			if err != nil {
				return err
			}
			slot.Status = models.SlotStatusFree
			slot.IsBlocked = false
			if err := tx.UpdateTimeSlot(ctx, &slot); err != nil {
				return err
			}
		}
		if reason == "" {
			reason = "status changed"
		}
		if err := tx.CreateAppointmentStatusHistory(ctx, &models.AppointmentStatusHistory{
			AppointmentID: appointment.ID,
			OldStatus:     oldStatus,
			NewStatus:     newStatus,
			ChangedBy:     changedBy,
			Reason:        reason,
			ChangedAt:     now,
		}); err != nil {
			return err
		}
		updated = appointment
		return nil
	})
	if err != nil {
		return models.Appointment{}, err
	}
	if updated.Status == models.AppointmentStatusCancelled {
		appointment, err := s.store.GetAppointment(ctx, updated.ID)
		if err == nil {
			s.scheduleCancellationNotification(ctx, appointment)
		}
	}
	return s.store.GetAppointment(ctx, updated.ID)
}

func (s *Service) RescheduleAppointment(ctx context.Context, id uuid.UUID, req RescheduleAppointmentRequest) (models.Appointment, error) {
	if req.TimeSlotID == uuid.Nil {
		return models.Appointment{}, fmt.Errorf("%w: time_slot_id is required", repository.ErrValidation)
	}

	var updated models.Appointment
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		appointment, err := tx.GetAppointmentForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if appointment.Status == models.AppointmentStatusCancelled ||
			appointment.Status == models.AppointmentStatusCompleted ||
			appointment.Status == models.AppointmentStatusNoShow {
			return fmt.Errorf("%w: terminal appointment cannot be rescheduled", repository.ErrConflict)
		}

		oldSlot, err := tx.GetTimeSlotForUpdate(ctx, appointment.TimeSlotID)
		if err != nil {
			return err
		}
		newSlot, err := tx.GetTimeSlotForUpdate(ctx, req.TimeSlotID)
		if err != nil {
			return err
		}
		if newSlot.IsBlocked || newSlot.Status != models.SlotStatusFree {
			return fmt.Errorf("%w: target slot is not available", repository.ErrConflict)
		}
		if newSlot.StartAt.Before(s.now()) {
			return fmt.Errorf("%w: cannot reschedule to past slot", repository.ErrValidation)
		}

		oldStatus := appointment.Status
		appointment.TimeSlotID = newSlot.ID
		appointment.DoctorID = newSlot.DoctorID
		appointment.AppointmentAt = newSlot.StartAt
		appointment.Status = models.AppointmentStatusCreated
		appointment.ConfirmedAt = nil
		appointment.CancelledAt = nil
		appointment.CompletedAt = nil
		appointment.ShowedUp = nil

		oldSlot.Status = models.SlotStatusFree
		oldSlot.IsBlocked = false
		newSlot.Status = models.SlotStatusBooked

		if err := tx.UpdateAppointment(ctx, &appointment); err != nil {
			return err
		}
		if err := tx.UpdateTimeSlot(ctx, &oldSlot); err != nil {
			return err
		}
		if err := tx.UpdateTimeSlot(ctx, &newSlot); err != nil {
			return err
		}

		reason := req.Reason
		if reason == "" {
			reason = "appointment rescheduled"
		}
		if err := tx.CreateAppointmentStatusHistory(ctx, &models.AppointmentStatusHistory{
			AppointmentID: appointment.ID,
			OldStatus:     oldStatus,
			NewStatus:     appointment.Status,
			ChangedBy:     req.ChangedBy,
			Reason:        reason,
			ChangedAt:     s.now(),
		}); err != nil {
			return err
		}

		updated = appointment
		return nil
	})
	if err != nil {
		return models.Appointment{}, err
	}
	return s.store.GetAppointment(ctx, updated.ID)
}

func allowedStatusTransition(oldStatus, newStatus models.AppointmentStatus) bool {
	if oldStatus == newStatus {
		return true
	}
	switch oldStatus {
	case models.AppointmentStatusCreated:
		return newStatus == models.AppointmentStatusConfirmed ||
			newStatus == models.AppointmentStatusCancelled ||
			newStatus == models.AppointmentStatusNoShow
	case models.AppointmentStatusConfirmed:
		return newStatus == models.AppointmentStatusCancelled ||
			newStatus == models.AppointmentStatusCompleted ||
			newStatus == models.AppointmentStatusNoShow
	default:
		return false
	}
}
