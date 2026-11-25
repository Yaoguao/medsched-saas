package service

import (
	"time"

	"base/internal/domain/models"
	"base/internal/integration/ehr"

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
	ExternalEHRID *string
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

type RescheduleAppointmentRequest struct {
	TimeSlotID uuid.UUID
	ChangedBy  *uuid.UUID
	Reason     string
}

type CreateRescheduleProposalRequest struct {
	TimeSlotID  uuid.UUID
	RequestedBy *uuid.UUID
	Reason      string
}

type RespondRescheduleProposalRequest struct {
	RespondedBy *uuid.UUID
	Reason      string
}

type CreateUserRequest struct {
	Email    string
	Phone    string
	Password string
}

type CreateStaffUserRequest struct {
	Role     models.UserRole
	Email    string
	Phone    string
	Password string
}

type UserPatch struct {
	Role   *models.UserRole
	Status *string
	Email  *string
	Phone  *string
}

type AuthToken struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
	User        models.User
}

type EHRSyncResult struct {
	Patient models.Patient
	Event   models.IntegrationEvent
}

type EHRPatientPreview struct {
	Provider models.IntegrationProvider
	Link     models.ExternalEntityLink
	External ehr.PatientProfile
	Changes  []EHRFieldChange
}

type EHRFieldChange struct {
	Field    string
	Current  any
	External any
	Changed  bool
}

type SyncPatientEHRRequest struct {
	ProviderID uuid.UUID
	Fields     []string
	ActorID    *uuid.UUID
}

type IntegrationEventFilter struct {
	Provider   string
	Status     string
	EntityType string
	EntityID   *uuid.UUID
	From       time.Time
	To         time.Time
}

type NoShowTrainingRow struct {
	PatientID                 uuid.UUID
	AppointmentAt             time.Time
	Age                       int
	Gender                    string
	Neighbourhood             string
	Scholarship               bool
	HasHypertension           bool
	HasDiabetes               bool
	HasAlcoholism             bool
	HasHandicap               bool
	SMSReceived               bool
	NotificationCount         int
	SentNotifications         int
	FailedNotifications       int
	PendingNotifications      int
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

type MLModelInfo struct {
	Loaded       bool
	ModelVersion string
	Algorithm    string
	ArtifactPath string
	Metrics      map[string]any
}
