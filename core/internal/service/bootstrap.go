package service

import (
	"context"
	"fmt"
	"strings"

	"base/internal/domain/models"
	"base/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type BootstrapAdminRequest struct {
	Enabled  bool
	Email    string
	Phone    string
	Password string
}

func (s *Service) EnsureBootstrapAdmin(ctx context.Context, req BootstrapAdminRequest) (models.User, bool, error) {
	if !req.Enabled {
		return models.User{}, false, nil
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return models.User{}, false, fmt.Errorf("%w: bootstrap admin email is required", repository.ErrValidation)
	}
	if len(req.Password) < 8 {
		return models.User{}, false, fmt.Errorf("%w: bootstrap admin password must contain at least 8 characters", repository.ErrValidation)
	}

	existing, err := s.store.GetUserByEmail(ctx, email)
	if err == nil {
		changed := false
		if existing.Role != models.UserRoleAdmin {
			existing.Role = models.UserRoleAdmin
			changed = true
		}
		if existing.Status != "ACTIVE" {
			existing.Status = "ACTIVE"
			changed = true
		}
		if changed {
			if err := s.store.UpdateUser(ctx, &existing); err != nil {
				return models.User{}, false, err
			}
			existing, err = s.store.GetUser(ctx, existing.ID)
			if err != nil {
				return models.User{}, false, err
			}
		}
		return existing, false, nil
	}
	if !IsNotFound(err) {
		return models.User{}, false, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, false, err
	}

	user := models.User{
		ID:           uuid.New(),
		Role:         models.UserRoleAdmin,
		Email:        email,
		Phone:        strings.TrimSpace(req.Phone),
		PasswordHash: string(hash),
		Status:       "ACTIVE",
	}
	if err := s.store.CreateUser(ctx, &user); err != nil {
		return models.User{}, false, err
	}
	created, err := s.store.GetUser(ctx, user.ID)
	if err != nil {
		return models.User{}, false, err
	}
	return created, true, nil
}
