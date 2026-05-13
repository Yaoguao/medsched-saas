package repository

import (
	"context"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
)

func (s *Store) CreateDoctor(ctx context.Context, doctor *models.Doctor) error {
	return s.db.WithContext(ctx).Create(doctor).Error
}

func (s *Store) ListDoctors(ctx context.Context, page pagination.Page) ([]models.Doctor, error) {
	page = pagination.Normalize(page)

	var doctors []models.Doctor
	err := s.db.WithContext(ctx).
		Limit(page.Limit).
		Offset(page.Offset).
		Order("full_name asc").
		Find(&doctors).Error
	return doctors, err
}

func (s *Store) GetDoctor(ctx context.Context, id uuid.UUID) (models.Doctor, error) {
	var doctor models.Doctor
	err := s.db.WithContext(ctx).First(&doctor, "id = ?", id).Error
	return doctor, mapGormError(err)
}

func (s *Store) UpdateDoctor(ctx context.Context, doctor *models.Doctor) error {
	return mapGormError(s.db.WithContext(ctx).Save(doctor).Error)
}
