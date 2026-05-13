package service

import (
	"context"
	"fmt"

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
	return created, nil
}

func (s *Service) GetAppointment(ctx context.Context, id uuid.UUID) (models.Appointment, error) {
	return s.store.GetAppointment(ctx, id)
}

func (s *Service) ListAppointments(ctx context.Context, page pagination.Page) ([]models.Appointment, error) {
	return s.store.ListAppointments(ctx, page)
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
	return updated, nil
}

func allowedStatusTransition(oldStatus, newStatus models.AppointmentStatus) bool {
	if oldStatus == newStatus {
		return true
	}
	switch oldStatus {
	case models.AppointmentStatusCreated:
		return newStatus == models.AppointmentStatusConfirmed || newStatus == models.AppointmentStatusCancelled
	case models.AppointmentStatusConfirmed:
		return newStatus == models.AppointmentStatusCancelled ||
			newStatus == models.AppointmentStatusCompleted ||
			newStatus == models.AppointmentStatusNoShow
	default:
		return false
	}
}
