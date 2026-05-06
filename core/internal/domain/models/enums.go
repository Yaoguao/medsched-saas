package models

type UserRole string

const (
	UserRoleAdmin       UserRole = "ADMIN"
	UserRoleRegistrator UserRole = "REGISTRATOR"
	UserRoleDoctor      UserRole = "DOCTOR"
	UserRolePatient     UserRole = "PATIENT"
	UserRoleManager     UserRole = "MANAGER"
	UserRoleITSupport   UserRole = "IT_SUPPORT"
)

type SlotStatus string

const (
	SlotStatusFree    SlotStatus = "FREE"
	SlotStatusBooked  SlotStatus = "BOOKED"
	SlotStatusBlocked SlotStatus = "BLOCKED"
	SlotStatusBreak   SlotStatus = "BREAK"
)

type AppointmentStatus string

const (
	AppointmentStatusCreated   AppointmentStatus = "CREATED"
	AppointmentStatusConfirmed AppointmentStatus = "CONFIRMED"
	AppointmentStatusCancelled AppointmentStatus = "CANCELLED"
	AppointmentStatusCompleted AppointmentStatus = "COMPLETED"
	AppointmentStatusNoShow    AppointmentStatus = "NO_SHOW"
)

type BookingChannel string

const (
	BookingChannelWeb      BookingChannel = "WEB"
	BookingChannelRegistry BookingChannel = "REGISTRY"
	BookingChannelPhone    BookingChannel = "PHONE"
)

type VisitType string

const (
	VisitTypePrimary      VisitType = "PRIMARY"
	VisitTypeRepeat       VisitType = "REPEAT"
	VisitTypeDiagnostic   VisitType = "DIAGNOSTIC"
	VisitTypeConsultation VisitType = "CONSULTATION"
)

type NotificationChannel string

const (
	NotificationChannelSMS   NotificationChannel = "SMS"
	NotificationChannelEmail NotificationChannel = "EMAIL"
	NotificationChannelPush  NotificationChannel = "PUSH"
)

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "PENDING"
	NotificationStatusSent      NotificationStatus = "SENT"
	NotificationStatusFailed    NotificationStatus = "FAILED"
	NotificationStatusDelivered NotificationStatus = "DELIVERED"
)

type NotificationType string

const (
	NotificationTypeConfirmation NotificationType = "CONFIRMATION"
	NotificationTypeReminder     NotificationType = "REMINDER"
	NotificationTypeCancellation NotificationType = "CANCELLATION"
)

type PredictionRiskLevel string

const (
	PredictionRiskLow    PredictionRiskLevel = "LOW"
	PredictionRiskMedium PredictionRiskLevel = "MEDIUM"
	PredictionRiskHigh   PredictionRiskLevel = "HIGH"
)
