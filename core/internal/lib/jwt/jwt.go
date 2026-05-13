package jwt

import (
	"fmt"
	"time"

	"base/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID
	Role      models.UserRole
	Email     string
	ExpiresAt time.Time
}

func NewToken(user *models.User, duration time.Duration, jwtSecret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID.String()
	claims["role"] = string(user.Role)
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string, jwtSecret string) (Claims, error) {
	var result Claims

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return result, err
	}
	if !token.Valid {
		return result, fmt.Errorf("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return result, fmt.Errorf("invalid claims")
	}

	uid, err := mapClaimString(mapClaims, "uid")
	if err != nil {
		return result, err
	}
	userID, err := uuid.Parse(uid)
	if err != nil {
		return result, err
	}

	email, err := mapClaimString(mapClaims, "email")
	if err != nil {
		return result, err
	}

	role, err := mapClaimString(mapClaims, "role")
	if err != nil {
		return result, err
	}

	expiresAt, err := mapClaims.GetExpirationTime()
	if err != nil {
		return result, err
	}
	if expiresAt == nil {
		return result, fmt.Errorf("missing exp")
	}

	return Claims{
		UserID:    userID,
		Role:      models.UserRole(role),
		Email:     email,
		ExpiresAt: expiresAt.Time,
	}, nil
}

func mapClaimString(claims jwt.MapClaims, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("missing %s", key)
	}
	str, ok := value.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("invalid %s", key)
	}
	return str, nil
}
