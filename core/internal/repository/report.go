package repository

import (
	"context"
	"time"

	"base/internal/domain/models"
)

func (s *Store) CountAppointmentsByStatus(ctx context.Context, from, to time.Time) (map[models.AppointmentStatus]int64, error) {
	var rows []struct {
		Status models.AppointmentStatus
		Count  int64
	}

	err := s.db.WithContext(ctx).
		Model(&models.Appointment{}).
		Select("status, count(*) as count").
		Where("appointment_at >= ? AND appointment_at <= ?", from, to).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[models.AppointmentStatus]int64, len(rows))
	for _, row := range rows {
		result[row.Status] = row.Count
	}
	return result, nil
}
