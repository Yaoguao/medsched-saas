package service

import (
	"context"
	"fmt"

	"base/internal/domain/models"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreateRescheduleProposal(ctx context.Context, appointmentID uuid.UUID, req CreateRescheduleProposalRequest) (models.AppointmentRescheduleProposal, error) {
	if req.TimeSlotID == uuid.Nil {
		return models.AppointmentRescheduleProposal{}, fmt.Errorf("%w: time_slot_id is required", repository.ErrValidation)
	}

	var created models.AppointmentRescheduleProposal
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		appointment, err := tx.GetAppointmentForUpdate(ctx, appointmentID)
		if err != nil {
			return err
		}
		if appointment.Status == models.AppointmentStatusCancelled ||
			appointment.Status == models.AppointmentStatusCompleted ||
			appointment.Status == models.AppointmentStatusNoShow {
			return fmt.Errorf("%w: terminal appointment cannot be rescheduled", repository.ErrConflict)
		}
		if appointment.TimeSlotID == req.TimeSlotID {
			return fmt.Errorf("%w: proposed slot is already used by appointment", repository.ErrValidation)
		}
		exists, err := tx.PendingRescheduleProposalExists(ctx, appointment.ID)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: pending reschedule proposal already exists", repository.ErrConflict)
		}

		slot, err := tx.GetTimeSlotForUpdate(ctx, req.TimeSlotID)
		if err != nil {
			return err
		}
		if slot.IsBlocked || slot.Status != models.SlotStatusFree {
			return fmt.Errorf("%w: target slot is not available", repository.ErrConflict)
		}
		if slot.StartAt.Before(s.now()) {
			return fmt.Errorf("%w: cannot propose past slot", repository.ErrValidation)
		}

		reason := req.Reason
		if reason == "" {
			reason = "doctor proposed appointment reschedule"
		}
		created = models.AppointmentRescheduleProposal{
			AppointmentID:      appointment.ID,
			ProposedTimeSlotID: slot.ID,
			RequestedBy:        req.RequestedBy,
			Reason:             reason,
			Status:             models.RescheduleProposalPending,
		}
		return tx.CreateRescheduleProposal(ctx, &created)
	})
	if err != nil {
		return models.AppointmentRescheduleProposal{}, err
	}
	return s.store.GetRescheduleProposal(ctx, created.ID)
}

func (s *Service) ListAppointmentRescheduleProposals(ctx context.Context, appointmentID uuid.UUID, page pagination.Page) ([]models.AppointmentRescheduleProposal, error) {
	return s.store.ListAppointmentRescheduleProposals(ctx, appointmentID, page)
}

func (s *Service) ListPatientRescheduleProposals(ctx context.Context, patientID uuid.UUID, page pagination.Page) ([]models.AppointmentRescheduleProposal, error) {
	return s.store.ListPatientRescheduleProposals(ctx, patientID, page)
}

func (s *Service) GetRescheduleProposal(ctx context.Context, proposalID uuid.UUID) (models.AppointmentRescheduleProposal, error) {
	return s.store.GetRescheduleProposal(ctx, proposalID)
}

func (s *Service) AcceptRescheduleProposal(ctx context.Context, proposalID uuid.UUID, req RespondRescheduleProposalRequest) (models.AppointmentRescheduleProposal, error) {
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		proposal, err := tx.GetRescheduleProposalForUpdate(ctx, proposalID)
		if err != nil {
			return err
		}
		if proposal.Status != models.RescheduleProposalPending {
			return fmt.Errorf("%w: reschedule proposal is not pending", repository.ErrConflict)
		}

		appointment, err := tx.GetAppointmentForUpdate(ctx, proposal.AppointmentID)
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
		newSlot, err := tx.GetTimeSlotForUpdate(ctx, proposal.ProposedTimeSlotID)
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
		now := s.now()
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

		proposal.Status = models.RescheduleProposalAccepted
		proposal.RespondedBy = req.RespondedBy
		proposal.RespondedAt = &now

		reason := req.Reason
		if reason == "" {
			reason = "reschedule proposal accepted"
		}
		proposal.ResponseReason = reason

		if err := tx.UpdateAppointment(ctx, &appointment); err != nil {
			return err
		}
		if err := tx.UpdateTimeSlot(ctx, &oldSlot); err != nil {
			return err
		}
		if err := tx.UpdateTimeSlot(ctx, &newSlot); err != nil {
			return err
		}
		if err := tx.UpdateRescheduleProposal(ctx, &proposal); err != nil {
			return err
		}

		return tx.CreateAppointmentStatusHistory(ctx, &models.AppointmentStatusHistory{
			AppointmentID: appointment.ID,
			OldStatus:     oldStatus,
			NewStatus:     appointment.Status,
			ChangedBy:     req.RespondedBy,
			Reason:        reason,
			ChangedAt:     now,
		})
	})
	if err != nil {
		return models.AppointmentRescheduleProposal{}, err
	}
	proposal, err := s.store.GetRescheduleProposal(ctx, proposalID)
	if err != nil {
		return models.AppointmentRescheduleProposal{}, err
	}
	_ = s.store.MarkPendingAppointmentNotificationsFailed(ctx, proposal.AppointmentID, "appointment rescheduled before scheduled send", s.now())
	appointment, err := s.store.GetAppointment(ctx, proposal.AppointmentID)
	if err != nil {
		return models.AppointmentRescheduleProposal{}, err
	}
	s.scheduleAppointmentNotifications(ctx, appointment)
	s.tryPredictAppointmentNoShow(ctx, proposal.AppointmentID)
	return s.store.GetRescheduleProposal(ctx, proposalID)
}

func (s *Service) DeclineRescheduleProposal(ctx context.Context, proposalID uuid.UUID, req RespondRescheduleProposalRequest) (models.AppointmentRescheduleProposal, error) {
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		proposal, err := tx.GetRescheduleProposalForUpdate(ctx, proposalID)
		if err != nil {
			return err
		}
		if proposal.Status != models.RescheduleProposalPending {
			return fmt.Errorf("%w: reschedule proposal is not pending", repository.ErrConflict)
		}
		now := s.now()
		proposal.Status = models.RescheduleProposalDeclined
		proposal.RespondedBy = req.RespondedBy
		proposal.RespondedAt = &now
		if req.Reason == "" {
			proposal.ResponseReason = "reschedule proposal declined"
		} else {
			proposal.ResponseReason = req.Reason
		}
		return tx.UpdateRescheduleProposal(ctx, &proposal)
	})
	if err != nil {
		return models.AppointmentRescheduleProposal{}, err
	}
	return s.store.GetRescheduleProposal(ctx, proposalID)
}
