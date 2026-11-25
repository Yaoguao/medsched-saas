package notifications

import (
	"net/http"
	"time"

	"base/internal/domain/models"
	"base/internal/http-server/handlers/common"
	authmiddleware "base/internal/http-server/middleware/auth"
	"base/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc            *service.Service
	callbackSecret string
}

func New(svc *service.Service, callbackSecret ...string) *Handler {
	secret := ""
	if len(callbackSecret) > 0 {
		secret = callbackSecret[0]
	}
	return &Handler{svc: svc, callbackSecret: secret}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/notifications/send", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleITSupport), h.create)
	r.GET("/notifications/:notification_id", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleITSupport), h.get)
	r.GET("/appointments/:appointment_id/notifications", h.listByAppointment)
}

func (h *Handler) RegisterCallbacks(r gin.IRouter) {
	r.POST("/notifications/:notification_id/delivery-status", h.deliveryStatusCallback)
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
	c.JSON(http.StatusCreated, responseFromModel(created))
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
	c.JSON(http.StatusOK, responseFromModel(notification))
}

func (h *Handler) listByAppointment(c *gin.Context) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}
	if !h.canAccessAppointment(c, id) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
		return
	}
	page := common.Pagination(c)
	notifications, err := h.svc.ListAppointmentNotifications(c.Request.Context(), id, page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, responsesFromModels(notifications), page)
}

func (h *Handler) deliveryStatusCallback(c *gin.Context) {
	if h.callbackSecret != "" && c.GetHeader("X-Notifier-Secret") != h.callbackSecret {
		common.RespondError(c, http.StatusUnauthorized, "invalid notifier secret")
		return
	}
	id, ok := common.PathUUID(c, "notification_id")
	if !ok {
		return
	}
	var req deliveryStatusRequest
	if !common.BindJSON(c, &req) {
		return
	}
	notification, err := h.svc.UpdateNotificationDeliveryStatus(c.Request.Context(), id, req.Status, req.SentAt, req.ErrorMessage)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, responseFromModel(notification))
}

func (h *Handler) canAccessAppointment(c *gin.Context, appointmentID uuid.UUID) bool {
	user, ok := authmiddleware.CurrentUser(c)
	if !ok {
		return false
	}
	if user.Role == models.UserRoleAdmin ||
		user.Role == models.UserRoleRegistrator ||
		user.Role == models.UserRoleManager ||
		user.Role == models.UserRoleITSupport {
		return true
	}
	appointment, err := h.svc.GetAppointment(c.Request.Context(), appointmentID)
	if err != nil {
		return false
	}
	if user.Role == models.UserRoleDoctor && user.Doctor != nil {
		return appointment.DoctorID == user.Doctor.ID
	}
	if user.Role == models.UserRolePatient && user.Patient != nil {
		return appointment.PatientID == user.Patient.ID
	}
	return false
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

type deliveryStatusRequest struct {
	Status       models.NotificationStatus `json:"status" binding:"required"`
	SentAt       *time.Time                `json:"sent_at"`
	ErrorMessage string                    `json:"error_message"`
}
