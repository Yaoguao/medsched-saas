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
