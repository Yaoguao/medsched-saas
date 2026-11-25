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

type DoctorNoShowRow struct {
	DoctorID   string
	DoctorName string
	Specialty  string
	Total      int64
	NoShow     int64
	Cancelled  int64
	Completed  int64
	NoShowRate float64
}

func (s *Store) NoShowReportByDoctor(ctx context.Context, from, to time.Time) ([]DoctorNoShowRow, error) {
	var rows []DoctorNoShowRow
	err := s.db.WithContext(ctx).
		Table("appointments").
		Select(`
			appointments.doctor_id::text as doctor_id,
			doctors.full_name as doctor_name,
			doctors.specialty as specialty,
			count(*) as total,
			sum(case when appointments.status = ? then 1 else 0 end) as no_show,
			sum(case when appointments.status = ? then 1 else 0 end) as cancelled,
			sum(case when appointments.status = ? then 1 else 0 end) as completed,
			case when count(*) = 0 then 0 else sum(case when appointments.status = ? then 1 else 0 end)::float / count(*) end as no_show_rate
		`, models.AppointmentStatusNoShow, models.AppointmentStatusCancelled, models.AppointmentStatusCompleted, models.AppointmentStatusNoShow).
		Joins("LEFT JOIN doctors ON doctors.id = appointments.doctor_id").
		Where("appointments.appointment_at >= ? AND appointments.appointment_at <= ?", from, to).
		Group("appointments.doctor_id, doctors.full_name, doctors.specialty").
		Order("no_show_rate desc, total desc").
		Scan(&rows).Error
	return rows, err
}

type DoctorLoadRow struct {
	DoctorID     string
	DoctorName   string
	Specialty    string
	SlotsTotal   int64
	FreeSlots    int64
	BookedSlots  int64
	BlockedSlots int64
	LoadRate     float64
}

func (s *Store) LoadReportByDoctor(ctx context.Context, from, to time.Time) ([]DoctorLoadRow, error) {
	var rows []DoctorLoadRow
	err := s.db.WithContext(ctx).
		Table("time_slots").
		Select(`
			time_slots.doctor_id::text as doctor_id,
			doctors.full_name as doctor_name,
			doctors.specialty as specialty,
			count(*) as slots_total,
			sum(case when time_slots.status = ? and time_slots.is_blocked = false then 1 else 0 end) as free_slots,
			sum(case when time_slots.status = ? then 1 else 0 end) as booked_slots,
			sum(case when time_slots.status = ? or time_slots.is_blocked = true then 1 else 0 end) as blocked_slots,
			case when count(*) = 0 then 0 else sum(case when time_slots.status = ? then 1 else 0 end)::float / count(*) end as load_rate
		`, models.SlotStatusFree, models.SlotStatusBooked, models.SlotStatusBlocked, models.SlotStatusBooked).
		Joins("LEFT JOIN doctors ON doctors.id = time_slots.doctor_id").
		Where("time_slots.start_at >= ? AND time_slots.start_at <= ?", from, to).
		Group("time_slots.doctor_id, doctors.full_name, doctors.specialty").
		Order("load_rate desc, slots_total desc").
		Scan(&rows).Error
	return rows, err
}
