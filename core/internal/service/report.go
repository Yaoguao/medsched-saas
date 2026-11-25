package service

import (
	"context"
	"fmt"
	"time"

	"base/internal/domain/models"
	"base/internal/repository"
)

type ReportRange struct {
	From time.Time
	To   time.Time
}

type AppointmentStatusReport struct {
	Range        ReportRange                        `json:"range"`
	StatusCounts map[models.AppointmentStatus]int64 `json:"status_counts"`
	Total        int64                              `json:"total"`
	NoShow       int64                              `json:"no_show"`
	Cancelled    int64                              `json:"cancelled"`
	Completed    int64                              `json:"completed"`
	NoShowRate   float64                            `json:"no_show_rate"`
}

type DoctorNoShowReportRow struct {
	DoctorID   string  `json:"doctor_id"`
	DoctorName string  `json:"doctor_name"`
	Specialty  string  `json:"specialty"`
	Total      int64   `json:"total"`
	NoShow     int64   `json:"no_show"`
	Cancelled  int64   `json:"cancelled"`
	Completed  int64   `json:"completed"`
	NoShowRate float64 `json:"no_show_rate"`
}

type NoShowReport struct {
	Range   ReportRange             `json:"range"`
	Summary AppointmentStatusReport `json:"summary"`
	Doctors []DoctorNoShowReportRow `json:"doctors"`
}

type DoctorLoadReportRow struct {
	DoctorID     string  `json:"doctor_id"`
	DoctorName   string  `json:"doctor_name"`
	Specialty    string  `json:"specialty"`
	SlotsTotal   int64   `json:"slots_total"`
	FreeSlots    int64   `json:"free_slots"`
	BookedSlots  int64   `json:"booked_slots"`
	BlockedSlots int64   `json:"blocked_slots"`
	LoadRate     float64 `json:"load_rate"`
}

type LoadReport struct {
	Range       ReportRange           `json:"range"`
	Doctors     []DoctorLoadReportRow `json:"doctors"`
	SlotsTotal  int64                 `json:"slots_total"`
	BookedSlots int64                 `json:"booked_slots"`
	LoadRate    float64               `json:"load_rate"`
}

type DashboardReport struct {
	Range        ReportRange             `json:"range"`
	Appointments AppointmentStatusReport `json:"appointments"`
	NoShow       NoShowReport            `json:"no_show"`
	Load         LoadReport              `json:"load"`
}

func (s *Service) AppointmentStatusReport(ctx context.Context, from, to time.Time) (map[models.AppointmentStatus]int64, error) {
	report, err := s.AppointmentReport(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return report.StatusCounts, nil
}

func (s *Service) AppointmentReport(ctx context.Context, from, to time.Time) (AppointmentStatusReport, error) {
	from, to, err := s.reportRange(from, to)
	if err != nil {
		return AppointmentStatusReport{}, err
	}
	statusCounts, err := s.store.CountAppointmentsByStatus(ctx, from, to)
	if err != nil {
		return AppointmentStatusReport{}, err
	}
	return appointmentStatusReport(from, to, statusCounts), nil
}

func (s *Service) NoShowReport(ctx context.Context, from, to time.Time) (NoShowReport, error) {
	summary, err := s.AppointmentReport(ctx, from, to)
	if err != nil {
		return NoShowReport{}, err
	}
	rows, err := s.store.NoShowReportByDoctor(ctx, summary.Range.From, summary.Range.To)
	if err != nil {
		return NoShowReport{}, err
	}
	doctors := make([]DoctorNoShowReportRow, 0, len(rows))
	for _, row := range rows {
		doctors = append(doctors, DoctorNoShowReportRow{
			DoctorID:   row.DoctorID,
			DoctorName: row.DoctorName,
			Specialty:  row.Specialty,
			Total:      row.Total,
			NoShow:     row.NoShow,
			Cancelled:  row.Cancelled,
			Completed:  row.Completed,
			NoShowRate: row.NoShowRate,
		})
	}
	return NoShowReport{
		Range:   summary.Range,
		Summary: summary,
		Doctors: doctors,
	}, nil
}

func (s *Service) LoadReport(ctx context.Context, from, to time.Time) (LoadReport, error) {
	from, to, err := s.reportRange(from, to)
	if err != nil {
		return LoadReport{}, err
	}
	rows, err := s.store.LoadReportByDoctor(ctx, from, to)
	if err != nil {
		return LoadReport{}, err
	}
	doctors := make([]DoctorLoadReportRow, 0, len(rows))
	var slotsTotal int64
	var bookedSlots int64
	for _, row := range rows {
		slotsTotal += row.SlotsTotal
		bookedSlots += row.BookedSlots
		doctors = append(doctors, DoctorLoadReportRow{
			DoctorID:     row.DoctorID,
			DoctorName:   row.DoctorName,
			Specialty:    row.Specialty,
			SlotsTotal:   row.SlotsTotal,
			FreeSlots:    row.FreeSlots,
			BookedSlots:  row.BookedSlots,
			BlockedSlots: row.BlockedSlots,
			LoadRate:     row.LoadRate,
		})
	}
	loadRate := 0.0
	if slotsTotal > 0 {
		loadRate = float64(bookedSlots) / float64(slotsTotal)
	}
	return LoadReport{
		Range:       ReportRange{From: from, To: to},
		Doctors:     doctors,
		SlotsTotal:  slotsTotal,
		BookedSlots: bookedSlots,
		LoadRate:    loadRate,
	}, nil
}

func (s *Service) DashboardReport(ctx context.Context, from, to time.Time) (DashboardReport, error) {
	appointments, err := s.AppointmentReport(ctx, from, to)
	if err != nil {
		return DashboardReport{}, err
	}
	noShow, err := s.NoShowReport(ctx, appointments.Range.From, appointments.Range.To)
	if err != nil {
		return DashboardReport{}, err
	}
	load, err := s.LoadReport(ctx, appointments.Range.From, appointments.Range.To)
	if err != nil {
		return DashboardReport{}, err
	}
	return DashboardReport{
		Range:        appointments.Range,
		Appointments: appointments,
		NoShow:       noShow,
		Load:         load,
	}, nil
}

func (s *Service) reportRange(from, to time.Time) (time.Time, time.Time, error) {
	if from.IsZero() {
		from = s.now().AddDate(0, -1, 0)
	}
	if to.IsZero() {
		to = s.now()
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: invalid report date range", repository.ErrValidation)
	}
	return from, to, nil
}

func appointmentStatusReport(from, to time.Time, statusCounts map[models.AppointmentStatus]int64) AppointmentStatusReport {
	var total int64
	for _, count := range statusCounts {
		total += count
	}
	noShow := statusCounts[models.AppointmentStatusNoShow]
	noShowRate := 0.0
	if total > 0 {
		noShowRate = float64(noShow) / float64(total)
	}
	return AppointmentStatusReport{
		Range:        ReportRange{From: from, To: to},
		StatusCounts: statusCounts,
		Total:        total,
		NoShow:       noShow,
		Cancelled:    statusCounts[models.AppointmentStatusCancelled],
		Completed:    statusCounts[models.AppointmentStatusCompleted],
		NoShowRate:   noShowRate,
	}
}
