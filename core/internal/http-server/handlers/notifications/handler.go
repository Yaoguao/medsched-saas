package notifications

import (
	"net/http"
	"time"

	"base/internal/domain/models"
	"base/internal/http-server/handlers/common"
	"base/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/notifications/send", h.create)
	r.GET("/notifications/:notification_id", h.get)
	r.GET("/appointments/:appointment_id/notifications", h.listByAppointment)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}
	notification := models.Notification{
		AppointmentID: req.AppointmentID,
		PatientID:     req.PatientID,
		Channel:       req.Channel,
		Type:          req.Type,
		Status:        req.Status,
		ScheduledAt:   req.ScheduledAt,
		Provider:      req.Provider,
		ExternalID:    req.ExternalID,
		ErrorMessage:  req.ErrorMessage,
	}
	created, err := h.svc.CreateNotification(c.Request.Context(), notification)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := common.PathUUID(c, "notification_id")
	if !ok {
		return
	}
	notification, err := h.svc.GetNotification(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, notification)
}

func (h *Handler) listByAppointment(c *gin.Context) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}
	page := common.Pagination(c)
	notifications, err := h.svc.ListAppointmentNotifications(c.Request.Context(), id, page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, notifications, page)
}

type createRequest struct {
	AppointmentID uuid.UUID                  `json:"appointment_id" binding:"required"`
	PatientID     uuid.UUID                  `json:"patient_id" binding:"required"`
	Channel       models.NotificationChannel `json:"channel" binding:"required"`
	Type          models.NotificationType    `json:"type" binding:"required"`
	Status        models.NotificationStatus  `json:"status"`
	ScheduledAt   time.Time                  `json:"scheduled_at"`
	Provider      string                     `json:"provider"`
	ExternalID    string                     `json:"external_id"`
	ErrorMessage  string                     `json:"error_message"`
}
