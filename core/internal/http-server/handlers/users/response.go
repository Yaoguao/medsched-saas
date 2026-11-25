package users

import (
	"time"

	"base/internal/domain/models"
	"base/internal/service"

	"github.com/google/uuid"
)

type userResponse struct {
	ID        uuid.UUID       `json:"id"`
	PatientID *uuid.UUID      `json:"patient_id,omitempty"`
	DoctorID  *uuid.UUID      `json:"doctor_id,omitempty"`
	Role      models.UserRole `json:"role"`
	Email     string          `json:"email"`
	Phone     string          `json:"phone"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type authResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        userResponse `json:"user"`
}

func userResponsesFromModels(users []models.User) []userResponse {
	responses := make([]userResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, userResponseFromModel(user))
	}
	return responses
}

func userResponseFromModel(user models.User) userResponse {
	var patientID *uuid.UUID
	if user.Patient != nil {
		patientID = &user.Patient.ID
	}
	var doctorID *uuid.UUID
	if user.Doctor != nil {
		doctorID = &user.Doctor.ID
	}
	return userResponse{
		ID:        user.ID,
		PatientID: patientID,
		DoctorID:  doctorID,
		Role:      user.Role,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func authResponseFromToken(token service.AuthToken) authResponse {
	return authResponse{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresAt:   token.ExpiresAt,
		User:        userResponseFromModel(token.User),
	}
}
