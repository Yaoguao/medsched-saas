package repository

import (
	"context"

	"base/internal/domain/models"
	"base/internal/lib/pagination"

	"github.com/google/uuid"
)

func (s *Store) CreateUser(ctx context.Context, user *models.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *Store) ListUsers(ctx context.Context, page pagination.Page) ([]models.User, error) {
	page = pagination.Normalize(page)

	var users []models.User
	err := s.db.WithContext(ctx).
		Preload("Patient").
		Limit(page.Limit).
		Offset(page.Offset).
		Order("created_at desc").
		Find(&users).Error
	return users, err
}

func (s *Store) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("Patient").
		First(&user, "id = ?", id).Error
	return user, mapGormError(err)
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("Patient").
		First(&user, "email = ?", email).Error
	return user, mapGormError(err)
}
