package service

import (
	"time"

	"base/internal/domain/models"

	"github.com/google/uuid"
)

type PatientPatch struct {
	FullName      *string
	Gender        *string
	BirthDate     *time.Time
	Age           *int
	Phone         *string
	Email         *string
	Neighbourhood *string
	Scholarship   *bool
}

type PatientMedicalProfilePatch struct {
	ChronicConditions *string
	RiskFactors       *string
	Allergies         *string
	Notes             *string
	HasHypertension   *bool
	HasDiabetes       *bool
	HasAlcoholism     *bool
	HasHandicap       *bool
}

type DoctorPatch struct {
	UserID          *uuid.UUID
	FullName        *string
	Specialty       *string
	Cabinet         *string
	ExperienceYears *int
	Status          *string
}

type SchedulePatch struct {
	StartDate     *time.Time
	EndDate       *time.Time
	WorkingDays   *string
	WorkStartTime *string
	WorkEndTime   *string
	Status        *string
}

type TimeSlotPatch struct {
	StartAt   *time.Time
	EndAt     *time.Time
	Status    *models.SlotStatus
	IsBlocked *bool
}

type CreateAppointmentRequest struct {
	PatientID      uuid.UUID
	TimeSlotID     uuid.UUID
	BookingChannel models.BookingChannel
	VisitType      models.VisitType
	ChangedBy      *uuid.UUID
}

type CreateUserRequest struct {
	FullName string
	Email    string
	Phone    string
	Password string
}

type RegisteredPatientUser struct {
	User    models.User
	Patient models.Patient
}

type AuthToken struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
	User        models.User
}

type NoShowTrainingRow struct {
	Age                       int
	Gender                    string
	Neighbourhood             string
	Scholarship               bool
	HasHypertension           bool
	HasDiabetes               bool
	HasAlcoholism             bool
	HasHandicap               bool
	SMSReceived               bool
	LeadTimeDays              int
	AppointmentWeekday        int
	AppointmentHour           int
	PreviousAppointmentsCount int
	PreviousNoShowCount       int
	PreviousNoShowRate        float64
	BookingChannel            string
	VisitType                 string
	DoctorSpecialty           string
	NoShow                    bool
}
