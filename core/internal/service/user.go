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

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (models.User, error) {
	if strings.TrimSpace(req.Email) == "" {
		return models.User{}, fmt.Errorf("%w: email is required", repository.ErrValidation)
	}
	if len(req.Password) < 8 {
		return models.User{}, fmt.Errorf("%w: password must contain at least 8 characters", repository.ErrValidation)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		ID:           uuid.New(),
		Role:         models.UserRolePatient,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Status:       "ACTIVE",
	}
	if err := s.store.CreateUser(ctx, &user); err != nil {
		return models.User{}, err
	}
	return s.store.GetUser(ctx, user.ID)
}

func (s *Service) CreateStaffUser(ctx context.Context, req CreateStaffUserRequest) (models.User, error) {
	if !isStaffRole(req.Role) {
		return models.User{}, fmt.Errorf("%w: staff role is required", repository.ErrValidation)
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return models.User{}, fmt.Errorf("%w: email is required", repository.ErrValidation)
	}
	phone := strings.TrimSpace(req.Phone)
	if phone == "" {
		return models.User{}, fmt.Errorf("%w: phone is required", repository.ErrValidation)
	}
	if len(req.Password) < 8 {
		return models.User{}, fmt.Errorf("%w: password must contain at least 8 characters", repository.ErrValidation)
	}
	if _, err := s.store.GetUserByEmail(ctx, email); err == nil {
		return models.User{}, fmt.Errorf("%w: user with email already exists", repository.ErrConflict)
	} else if !IsNotFound(err) {
		return models.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		ID:           uuid.New(),
		Role:         req.Role,
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hash),
		Status:       "ACTIVE",
	}
	if err := s.store.CreateUser(ctx, &user); err != nil {
		return models.User{}, err
	}
	return s.store.GetUser(ctx, user.ID)
}

func (s *Service) ListUsers(ctx context.Context, page pagination.Page) ([]models.User, error) {
	return s.store.ListUsers(ctx, page)
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	return s.store.GetUser(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, patch UserPatch) (models.User, error) {
	user, err := s.store.GetUser(ctx, id)
	if err != nil {
		return models.User{}, err
	}
	if patch.Role != nil {
		user.Role = *patch.Role
	}
	if patch.Status != nil {
		if strings.TrimSpace(*patch.Status) == "" {
			return models.User{}, fmt.Errorf("%w: status cannot be empty", repository.ErrValidation)
		}
		user.Status = strings.TrimSpace(*patch.Status)
	}
	if patch.Email != nil {
		if strings.TrimSpace(*patch.Email) == "" {
			return models.User{}, fmt.Errorf("%w: email cannot be empty", repository.ErrValidation)
		}
		user.Email = strings.ToLower(strings.TrimSpace(*patch.Email))
	}
	if patch.Phone != nil {
		user.Phone = strings.TrimSpace(*patch.Phone)
	}
	if err := s.store.UpdateUser(ctx, &user); err != nil {
		return models.User{}, err
	}
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

func isStaffRole(role models.UserRole) bool {
	switch role {
	case models.UserRoleAdmin,
		models.UserRoleRegistrator,
		models.UserRoleDoctor,
		models.UserRoleManager,
		models.UserRoleITSupport:
		return true
	default:
		return false
	}
}
