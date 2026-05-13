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
	Role      models.UserRole `json:"role"`
	FullName  string          `json:"full_name"`
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

type registeredUserResponse struct {
	User    userResponse              `json:"user"`
	Patient registeredPatientResponse `json:"patient"`
}

type registeredPatientResponse struct {
	ID            uuid.UUID  `json:"id"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	FullName      string     `json:"full_name"`
	Gender        string     `json:"gender"`
	BirthDate     time.Time  `json:"birth_date"`
	Age           int        `json:"age"`
	Phone         string     `json:"phone"`
	Email         string     `json:"email"`
	Neighbourhood string     `json:"neighbourhood"`
	Scholarship   bool       `json:"scholarship"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
	return userResponse{
		ID:        user.ID,
		PatientID: patientID,
		Role:      user.Role,
		FullName:  user.FullName,
		Email:     user.Email,
		Phone:     user.Phone,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func registeredUserResponseFromModel(registered service.RegisteredPatientUser) registeredUserResponse {
	user := registered.User
	user.Patient = &registered.Patient

	return registeredUserResponse{
		User:    userResponseFromModel(user),
		Patient: registeredPatientResponseFromModel(registered.Patient),
	}
}

func registeredPatientResponseFromModel(patient models.Patient) registeredPatientResponse {
	return registeredPatientResponse{
		ID:            patient.ID,
		UserID:        patient.UserID,
		FullName:      patient.FullName,
		Gender:        patient.Gender,
		BirthDate:     patient.BirthDate,
		Age:           patient.Age,
		Phone:         patient.Phone,
		Email:         patient.Email,
		Neighbourhood: patient.Neighbourhood,
		Scholarship:   patient.Scholarship,
		CreatedAt:     patient.CreatedAt,
		UpdatedAt:     patient.UpdatedAt,
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
