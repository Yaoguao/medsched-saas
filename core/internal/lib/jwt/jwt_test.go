package jwt

import (
	"base/internal/domain/models"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewToken(t *testing.T) {
	uid, _ := uuid.NewUUID()
	user := models.User{
		ID:       uid,
		Email:    "test@example.com",
		Role:     models.UserRoleAdmin,
		FullName: "Test User",
	}

	secret := "supersecret"

	duration := time.Hour

	tokenString, err := NewToken(&user, duration, secret)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected token, got empty string")
	}

	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("expected uid %s, got %s", user.ID, claims.UserID)
	}

	if claims.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.Role != user.Role {
		t.Errorf("expected role %s, got %s", user.Role, claims.Role)
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	uid := uuid.New()
	user := models.User{
		ID:    uid,
		Email: "test@example.com",
		Role:  models.UserRoleAdmin,
	}

	tokenString, err := NewToken(&user, time.Hour, "right-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := ParseToken(tokenString, "wrong-secret"); err == nil {
		t.Fatal("expected parse to fail")
	}
}
