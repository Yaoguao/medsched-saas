package service

import (
	"context"
	"errors"
	"fmt"

	"base/internal/domain/models"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreateDoctor(ctx context.Context, doctor models.Doctor) (models.Doctor, error) {
	if doctor.FullName == "" {
		return models.Doctor{}, fmt.Errorf("%w: full_name is required", repository.ErrValidation)
	}
	if doctor.Specialty == "" {
		return models.Doctor{}, fmt.Errorf("%w: specialty is required", repository.ErrValidation)
	}
	if err := s.store.CreateDoctor(ctx, &doctor); err != nil {
		return models.Doctor{}, err
	}
	return doctor, nil
}

func (s *Service) ListDoctors(ctx context.Context, page pagination.Page, filter repository.DoctorListFilter) ([]models.Doctor, error) {
	return s.store.ListDoctors(ctx, page, filter)
}

func (s *Service) ListDoctorSpecialties(ctx context.Context, activeOnly bool) ([]string, error) {
	return s.store.ListDoctorSpecialties(ctx, activeOnly)
}

func (s *Service) GetDoctor(ctx context.Context, id uuid.UUID) (models.Doctor, error) {
	return s.store.GetDoctor(ctx, id)
}

func (s *Service) UpdateDoctor(ctx context.Context, id uuid.UUID, patch DoctorPatch) (models.Doctor, error) {
	doctor, err := s.store.GetDoctor(ctx, id)
	if err != nil {
		return models.Doctor{}, err
	}

	if patch.UserID != nil {
		doctor.UserID = patch.UserID
	}
	if patch.FullName != nil {
		if *patch.FullName == "" {
			return models.Doctor{}, fmt.Errorf("%w: full_name cannot be empty", repository.ErrValidation)
		}
		doctor.FullName = *patch.FullName
	}
	if patch.Specialty != nil {
		if *patch.Specialty == "" {
			return models.Doctor{}, fmt.Errorf("%w: specialty cannot be empty", repository.ErrValidation)
		}
		doctor.Specialty = *patch.Specialty
	}
	if patch.Cabinet != nil {
		doctor.Cabinet = *patch.Cabinet
	}
	if patch.ExperienceYears != nil {
		if *patch.ExperienceYears < 0 {
			return models.Doctor{}, fmt.Errorf("%w: experience_years cannot be negative", repository.ErrValidation)
		}
		doctor.ExperienceYears = *patch.ExperienceYears
	}
	if patch.Status != nil {
		doctor.Status = *patch.Status
	}

	if err := s.store.UpdateDoctor(ctx, &doctor); err != nil {
		return models.Doctor{}, err
	}
	return doctor, nil
}

func (s *Service) LinkDoctorUser(ctx context.Context, doctorID uuid.UUID, userID uuid.UUID) (models.Doctor, error) {
	var doctor models.Doctor
	err := s.store.WithTx(ctx, func(tx *repository.Store) error {
		currentDoctor, err := tx.GetDoctor(ctx, doctorID)
		if err != nil {
			return err
		}

		user, err := tx.GetUser(ctx, userID)
		if err != nil {
			return err
		}
		if user.Status != "ACTIVE" {
			return fmt.Errorf("%w: user must be active", repository.ErrValidation)
		}

		linkedDoctor, err := tx.GetDoctorByUserID(ctx, userID)
		if err == nil && linkedDoctor.ID != doctorID {
			return fmt.Errorf("%w: user is already linked to another doctor", repository.ErrConflict)
		}
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return err
		}

		currentDoctor.UserID = &userID
		user.Role = models.UserRoleDoctor

		if err := tx.UpdateUser(ctx, &user); err != nil {
			return err
		}
		if err := tx.UpdateDoctor(ctx, &currentDoctor); err != nil {
			return err
		}

		doctor = currentDoctor
		return nil
	})
	if err != nil {
		return models.Doctor{}, err
	}
	return doctor, nil
}
