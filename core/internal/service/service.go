package service

import (
	"base/internal/integration/ehr"
	"base/internal/integration/ml"
	"base/internal/integration/notifier"
	"errors"
	"time"

	"base/internal/repository"
)

type Service struct {
	store          *repository.Store
	now            func() time.Time
	tokenSecret    string
	tokenTTL       time.Duration
	mlClient       *ml.Client
	notifierClient *notifier.Client
	ehrClient      ehr.Client
	fhirClient     ehr.Client
}

func WithEHRClient(client ehr.Client) Option {
	return func(s *Service) {
		s.ehrClient = client
	}
}

func WithFHIRClient(client ehr.Client) Option {
	return func(s *Service) {
		s.fhirClient = client
	}
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

func WithNotifierClient(client *notifier.Client) Option {
	return func(s *Service) {
		s.notifierClient = client
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
