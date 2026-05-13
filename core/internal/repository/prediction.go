package repository

import (
	"context"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

func (s *Store) CreatePredictionResult(ctx context.Context, prediction *models.PredictionResult) error {
	return s.db.WithContext(ctx).Create(prediction).Error
}

func (s *Store) GetPredictionResult(ctx context.Context, id uuid.UUID) (models.PredictionResult, error) {
	var prediction models.PredictionResult
	err := s.db.WithContext(ctx).First(&prediction, "id = ?", id).Error
	return prediction, mapGormError(err)
}
