package service

import (
	"testing"

	"base/internal/domain/models"
)

func TestAllowedStatusTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from models.AppointmentStatus
		to   models.AppointmentStatus
		want bool
	}{
		{name: "created to confirmed", from: models.AppointmentStatusCreated, to: models.AppointmentStatusConfirmed, want: true},
		{name: "created to cancelled", from: models.AppointmentStatusCreated, to: models.AppointmentStatusCancelled, want: true},
		{name: "created to completed denied", from: models.AppointmentStatusCreated, to: models.AppointmentStatusCompleted, want: false},
		{name: "confirmed to completed", from: models.AppointmentStatusConfirmed, to: models.AppointmentStatusCompleted, want: true},
		{name: "confirmed to no show", from: models.AppointmentStatusConfirmed, to: models.AppointmentStatusNoShow, want: true},
		{name: "cancelled is terminal", from: models.AppointmentStatusCancelled, to: models.AppointmentStatusConfirmed, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := allowedStatusTransition(tt.from, tt.to); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}
}
