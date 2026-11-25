package authmiddleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"base/internal/domain/models"
	"base/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CurrentUserKey = "current_user"

func Authenticate(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := svc.CurrentUserFromAuthorization(c.Request.Context(), c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(CurrentUserKey, user)
		c.Next()
	}
}

func RequireRoles(roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[models.UserRole]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if _, ok := allowed[user.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func RequireServiceTokenOrRoles(svc *service.Service, serviceToken string, roles ...models.UserRole) gin.HandlerFunc {
	allowed := make(map[models.UserRole]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if validServiceToken(authorization, serviceToken) {
			c.Next()
			return
		}

		user, err := svc.CurrentUserFromAuthorization(c.Request.Context(), authorization)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if _, ok := allowed[user.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Set(CurrentUserKey, user)
		c.Next()
	}
}

func RequirePatientAccess(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if isOperationalRole(user.Role) {
			c.Next()
			return
		}
		id, err := uuid.Parse(c.Param(param))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
			return
		}
		if user.Role == models.UserRolePatient && user.Patient != nil && user.Patient.ID == id {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

func RequireDoctorAccess(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if isOperationalRole(user.Role) {
			c.Next()
			return
		}
		id, err := uuid.Parse(c.Param(param))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
			return
		}
		if user.Role == models.UserRoleDoctor && user.Doctor != nil && user.Doctor.ID == id {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

func CurrentUser(c *gin.Context) (models.User, bool) {
	value, ok := c.Get(CurrentUserKey)
	if !ok {
		return models.User{}, false
	}
	user, ok := value.(models.User)
	return user, ok
}

func isOperationalRole(role models.UserRole) bool {
	return role == models.UserRoleAdmin ||
		role == models.UserRoleRegistrator ||
		role == models.UserRoleManager ||
		role == models.UserRoleITSupport
}

func validServiceToken(authorization, expected string) bool {
	const prefix = "Bearer "
	if expected == "" || !strings.HasPrefix(authorization, prefix) {
		return false
	}
	actual := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
	if len(actual) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}
