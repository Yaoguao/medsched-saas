package service

import (
	"errors"
	"time"

	"base/internal/repository"
)

type Service struct {
	store *repository.Store
	now   func() time.Time
}

func New(store *repository.Store) *Service {
	return &Service{
		store: store,
		now:   time.Now,
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
