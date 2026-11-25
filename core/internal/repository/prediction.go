package repository

import (
	"context"

	"base/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

func (s *Store) CreatePredictionResult(ctx context.Context, prediction *models.PredictionResult) error {
	return s.db.WithContext(ctx).Create(prediction).Error
}

func (s *Store) UpsertPredictionResult(ctx context.Context, prediction *models.PredictionResult) error {
	return mapGormError(s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "appointment_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"probability",
			"risk_level",
			"model_version",
			"requested_at",
			"received_at",
			"updated_at",
		}),
	}).Create(prediction).Error)
}

func (s *Store) GetPredictionResult(ctx context.Context, id uuid.UUID) (models.PredictionResult, error) {
	var prediction models.PredictionResult
	err := s.db.WithContext(ctx).First(&prediction, "id = ?", id).Error
	return prediction, mapGormError(err)
}

func (s *Store) GetPredictionResultByAppointment(ctx context.Context, appointmentID uuid.UUID) (models.PredictionResult, error) {
	var prediction models.PredictionResult
	err := s.db.WithContext(ctx).First(&prediction, "appointment_id = ?", appointmentID).Error
	return prediction, mapGormError(err)
}
