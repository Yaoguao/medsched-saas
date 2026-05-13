package service

import (
	"context"
	"fmt"

	"base/internal/domain/models"
	"base/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) CreatePredictionResult(ctx context.Context, prediction models.PredictionResult) (models.PredictionResult, error) {
	if prediction.AppointmentID == uuid.Nil {
		return models.PredictionResult{}, fmt.Errorf("%w: appointment_id is required", repository.ErrValidation)
	}
	if prediction.Probability < 0 || prediction.Probability > 1 {
		return models.PredictionResult{}, fmt.Errorf("%w: probability must be between 0 and 1", repository.ErrValidation)
	}
	if prediction.RequestedAt.IsZero() {
		prediction.RequestedAt = s.now()
	}
	if prediction.ReceivedAt.IsZero() {
		prediction.ReceivedAt = s.now()
	}
	if prediction.RiskLevel == "" {
		prediction.RiskLevel = riskLevel(prediction.Probability)
	}
	if err := s.store.CreatePredictionResult(ctx, &prediction); err != nil {
		return models.PredictionResult{}, err
	}
	return prediction, nil
}

func (s *Service) GetPredictionResult(ctx context.Context, id uuid.UUID) (models.PredictionResult, error) {
	return s.store.GetPredictionResult(ctx, id)
}

func riskLevel(probability float64) models.PredictionRiskLevel {
	switch {
	case probability >= 0.7:
		return models.PredictionRiskHigh
	case probability >= 0.4:
		return models.PredictionRiskMedium
	default:
		return models.PredictionRiskLow
	}
}
