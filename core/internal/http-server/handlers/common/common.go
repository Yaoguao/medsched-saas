package common

import (
	"net/http"
	"strconv"
	"time"

	"base/internal/lib/pagination"
	"base/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListResponse struct {
	Items      any             `json:"items"`
	Pagination pagination.Page `json:"pagination"`
}

func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func PathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	value := c.Param(name)
	id, err := uuid.Parse(value)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}

func Pagination(c *gin.Context) pagination.Page {
	return pagination.Normalize(pagination.Page{
		Limit:  queryInt(c, "limit"),
		Offset: queryInt(c, "offset"),
	})
}

func RespondList(c *gin.Context, items any, page pagination.Page) {
	c.JSON(http.StatusOK, ListResponse{
		Items:      items,
		Pagination: pagination.Normalize(page),
	})
}

func ParseOptionalTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", value)
}

func HandleError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case service.IsNotFound(err):
		RespondError(c, http.StatusNotFound, err.Error())
	case service.IsConflict(err):
		RespondError(c, http.StatusConflict, err.Error())
	case service.IsValidation(err):
		RespondError(c, http.StatusUnprocessableEntity, err.Error())
	default:
		RespondError(c, http.StatusInternalServerError, err.Error())
	}
	return true
}

func RespondError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": message})
}

func queryInt(c *gin.Context, name string) int {
	value := c.Query(name)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
