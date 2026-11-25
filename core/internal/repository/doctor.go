package repository

import (
	"context"
	"fmt"
	"strings"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
)

func (s *Store) CreateDoctor(ctx context.Context, doctor *models.Doctor) error {
	return s.db.WithContext(ctx).Create(doctor).Error
}

type DoctorListFilter struct {
	Search     string
	Specialty  string
	Sort       string
	ActiveOnly bool
}

func (s *Store) ListDoctors(ctx context.Context, page pagination.Page, filter DoctorListFilter) ([]models.Doctor, error) {
	page = pagination.Normalize(page)

	var doctors []models.Doctor
	query := s.db.WithContext(ctx).Model(&models.Doctor{})
	if search := strings.TrimSpace(filter.Search); search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(full_name) LIKE ? OR LOWER(specialty) LIKE ? OR LOWER(cabinet) LIKE ?",
			pattern,
			pattern,
			pattern,
		)
	}
	if specialty := strings.TrimSpace(filter.Specialty); specialty != "" {
		query = query.Where("LOWER(specialty) = ?", strings.ToLower(specialty))
	}
	if filter.ActiveOnly {
		query = query.Where("status = ?", "ACTIVE")
	}
	order := "full_name asc"
	switch filter.Sort {
	case "experience_desc":
		order = "experience_years desc, full_name asc"
	case "experience_asc":
		order = "experience_years asc, full_name asc"
	case "name_desc":
		order = "full_name desc"
	case "", "name_asc":
	default:
		return nil, fmt.Errorf("%w: unsupported doctor sort", ErrValidation)
	}
	err := query.
		Limit(page.Limit).
		Offset(page.Offset).
		Order(order).
		Find(&doctors).Error
	return doctors, err
}

func (s *Store) ListDoctorSpecialties(ctx context.Context, activeOnly bool) ([]string, error) {
	var specialties []string
	query := s.db.WithContext(ctx).
		Model(&models.Doctor{}).
		Distinct("specialty").
		Where("specialty <> ''")
	if activeOnly {
		query = query.Where("status = ?", "ACTIVE")
	}
	err := query.Order("specialty asc").Pluck("specialty", &specialties).Error
	return specialties, err
}

func (s *Store) GetDoctor(ctx context.Context, id uuid.UUID) (models.Doctor, error) {
	var doctor models.Doctor
	err := s.db.WithContext(ctx).First(&doctor, "id = ?", id).Error
	return doctor, mapGormError(err)
}

func (s *Store) GetDoctorByUserID(ctx context.Context, userID uuid.UUID) (models.Doctor, error) {
	var doctor models.Doctor
	err := s.db.WithContext(ctx).First(&doctor, "user_id = ?", userID).Error
	return doctor, mapGormError(err)
}

func (s *Store) GetDoctorByUserEmail(ctx context.Context, email string) (models.Doctor, error) {
	var doctor models.Doctor
	err := s.db.WithContext(ctx).
		Joins("JOIN users ON users.id = doctors.user_id").
		Where("LOWER(users.email) = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&doctor).Error
	return doctor, mapGormError(err)
}

func (s *Store) UpdateDoctor(ctx context.Context, doctor *models.Doctor) error {
	return mapGormError(s.db.WithContext(ctx).Save(doctor).Error)
}
