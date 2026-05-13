package service

import (
	"base/internal/integration/ml"
	"errors"
	"time"

	"base/internal/repository"
)

type Service struct {
	store       *repository.Store
	now         func() time.Time
	tokenSecret string
	tokenTTL    time.Duration
	mlClient    *ml.Client
}

type Option func(*Service)

func New(store *repository.Store, opts ...Option) *Service {
	s := &Service{
		store:       store,
		now:         time.Now,
		tokenSecret: "dev-secret-change-me",
		tokenTTL:    24 * time.Hour,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithMLClient(client *ml.Client) Option {
	return func(s *Service) {
		s.mlClient = client
	}
}

func WithAuth(secret string, ttl time.Duration) Option {
	return func(s *Service) {
		if secret != "" {
			s.tokenSecret = secret
		}
		if ttl > 0 {
			s.tokenTTL = ttl
		}
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}

func IsConflict(err error) bool {
	return errors.Is(err, repository.ErrConflict)
}

func IsValidation(err error) bool {
	return errors.Is(err, repository.ErrValidation)
}
