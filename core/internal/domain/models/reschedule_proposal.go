package models

import (
	"time"

	"github.com/google/uuid"
)

type RescheduleProposalStatus string

const (
	RescheduleProposalPending  RescheduleProposalStatus = "PENDING"
	RescheduleProposalAccepted RescheduleProposalStatus = "ACCEPTED"
	RescheduleProposalDeclined RescheduleProposalStatus = "DECLINED"
)

type AppointmentRescheduleProposal struct {
	ID                 uuid.UUID                `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AppointmentID      uuid.UUID                `gorm:"type:uuid;not null;index"`
	ProposedTimeSlotID uuid.UUID                `gorm:"type:uuid;not null;index"`
	RequestedBy        *uuid.UUID               `gorm:"type:uuid;index"`
	RespondedBy        *uuid.UUID               `gorm:"type:uuid;index"`
	Reason             string                   `gorm:"type:text"`
	ResponseReason     string                   `gorm:"type:text"`
	Status             RescheduleProposalStatus `gorm:"type:varchar(32);not null;default:'PENDING';index"`
	RespondedAt        *time.Time

	Appointment      Appointment `gorm:"foreignKey:AppointmentID"`
	ProposedTimeSlot TimeSlot    `gorm:"foreignKey:ProposedTimeSlotID"`

	BaseModel
}

func (AppointmentRescheduleProposal) TableName() string {
	return "appointment_reschedule_proposals"
}
