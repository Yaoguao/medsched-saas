package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) DB() *gorm.DB {
	return s.db
}

func (s *Store) WithTx(ctx context.Context, fn func(*Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewStore(tx))
	})
}

func mapGormError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return fmt.Errorf("gorm: %w", err)
}
