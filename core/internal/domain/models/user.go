package models

import (
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Role         UserRole  `gorm:"type:varchar(32);not null;index"`
	FullName     string    `gorm:"type:varchar(255);not null"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex"`
	Phone        string    `gorm:"type:varchar(32);uniqueIndex"`
	PasswordHash string    `gorm:"type:text;not null"`
	Status       string    `gorm:"type:varchar(32);not null;default:'ACTIVE'"`

	Doctor *Doctor `gorm:"foreignKey:UserID"`

	BaseModel
}

func (User) TableName() string {
	return "users"
}
