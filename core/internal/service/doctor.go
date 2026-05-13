package service

import (
	"context"
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

func (s *Service) ListDoctors(ctx context.Context, page pagination.Page) ([]models.Doctor, error) {
	return s.store.ListDoctors(ctx, page)
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
