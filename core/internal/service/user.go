package service

import (
	"context"
	"fmt"
	"strings"

	"base/internal/domain/models"
	jwtlib "base/internal/lib/jwt"
	"base/internal/lib/pagination"
	"base/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (RegisteredPatientUser, error) {
	if strings.TrimSpace(req.FullName) == "" {
		return RegisteredPatientUser{}, fmt.Errorf("%w: full_name is required", repository.ErrValidation)
	}
	if strings.TrimSpace(req.Email) == "" {
		return RegisteredPatientUser{}, fmt.Errorf("%w: email is required", repository.ErrValidation)
	}
	if len(req.Password) < 8 {
		return RegisteredPatientUser{}, fmt.Errorf("%w: password must contain at least 8 characters", repository.ErrValidation)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisteredPatientUser{}, err
	}

	user := models.User{
		ID:           uuid.New(),
		Role:         models.UserRolePatient,
		FullName:     req.FullName,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Status:       "ACTIVE",
	}

	var registered RegisteredPatientUser
	err = s.store.WithTx(ctx, func(tx *repository.Store) error {
		if err := tx.CreateUser(ctx, &user); err != nil {
			return err
		}

		userID := user.ID
		patient := models.Patient{
			ID:       uuid.New(),
			UserID:   &userID,
			FullName: user.FullName,
			Gender:   "UNKNOWN",
			Age:      0,
			Phone:    user.Phone,
			Email:    user.Email,
		}
		if err := tx.CreatePatient(ctx, &patient); err != nil {
			return err
		}

		registered = RegisteredPatientUser{
			User:    user,
			Patient: patient,
		}
		return nil
	})
	if err != nil {
		return RegisteredPatientUser{}, err
	}
	return registered, nil
}

func (s *Service) ListUsers(ctx context.Context, page pagination.Page) ([]models.User, error) {
	return s.store.ListUsers(ctx, page)
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	return s.store.GetUser(ctx, id)
}

func (s *Service) Login(ctx context.Context, email, password string) (AuthToken, error) {
	user, err := s.store.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return AuthToken{}, fmt.Errorf("%w: invalid email or password", repository.ErrValidation)
	}
	if user.Status != "ACTIVE" {
		return AuthToken{}, fmt.Errorf("%w: user is not active", repository.ErrValidation)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthToken{}, fmt.Errorf("%w: invalid email or password", repository.ErrValidation)
	}

	token, err := jwtlib.NewToken(&user, s.tokenTTL, s.tokenSecret)
	if err != nil {
		return AuthToken{}, err
	}
	return AuthToken{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   s.now().Add(s.tokenTTL),
		User:        user,
	}, nil
}

func (s *Service) CurrentUser(ctx context.Context, token string) (models.User, error) {
	claims, err := jwtlib.ParseToken(token, s.tokenSecret)
	if err != nil {
		return models.User{}, fmt.Errorf("%w: invalid token", repository.ErrValidation)
	}
	return s.store.GetUser(ctx, claims.UserID)
}

func (s *Service) CurrentUserFromAuthorization(ctx context.Context, authorizationHeader string) (models.User, error) {
	token, err := bearerToken(authorizationHeader)
	if err != nil {
		return models.User{}, fmt.Errorf("%w: %s", repository.ErrValidation, err.Error())
	}
	return s.CurrentUser(ctx, token)
}

func bearerToken(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", fmt.Errorf("authorization header must use Bearer scheme")
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", fmt.Errorf("bearer token is empty")
	}
	return token, nil
}
