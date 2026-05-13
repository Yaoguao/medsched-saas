package appointments

import (
	"net/http"

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
	r.GET("/appointments", h.list)
	r.POST("/appointments", h.create)
	r.GET("/appointments/:appointment_id", h.get)
	r.POST("/appointments/:appointment_id/confirm", h.confirm)
	r.POST("/appointments/:appointment_id/cancel", h.cancel)
	r.POST("/appointments/:appointment_id/check-in", h.checkIn)
	r.POST("/appointments/:appointment_id/no-show", h.noShow)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}
	appointment, err := h.svc.CreateAppointment(c.Request.Context(), service.CreateAppointmentRequest{
		PatientID:      req.PatientID,
		TimeSlotID:     req.TimeSlotID,
		BookingChannel: req.BookingChannel,
		VisitType:      req.VisitType,
		ChangedBy:      req.ChangedBy,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, ResponseFromModel(appointment))
}

func (h *Handler) list(c *gin.Context) {
	page := common.Pagination(c)
	appointments, err := h.svc.ListAppointments(c.Request.Context(), page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, ResponsesFromModels(appointments), page)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}
	appointment, err := h.svc.GetAppointment(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, ResponseFromModel(appointment))
}

func (h *Handler) confirm(c *gin.Context) {
	h.changeStatus(c, models.AppointmentStatusConfirmed)
}

func (h *Handler) cancel(c *gin.Context) {
	h.changeStatus(c, models.AppointmentStatusCancelled)
}

func (h *Handler) checkIn(c *gin.Context) {
	h.changeStatus(c, models.AppointmentStatusCompleted)
}

func (h *Handler) noShow(c *gin.Context) {
	h.changeStatus(c, models.AppointmentStatusNoShow)
}

func (h *Handler) changeStatus(c *gin.Context, status models.AppointmentStatus) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}
	var req changeStatusRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if !common.BindJSON(c, &req) {
			return
		}
	}
	appointment, err := h.svc.ChangeAppointmentStatus(c.Request.Context(), id, status, req.ChangedBy, req.Reason)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, ResponseFromModel(appointment))
}

type createRequest struct {
	PatientID      uuid.UUID             `json:"patient_id" binding:"required"`
	TimeSlotID     uuid.UUID             `json:"time_slot_id" binding:"required"`
	BookingChannel models.BookingChannel `json:"booking_channel"`
	VisitType      models.VisitType      `json:"visit_type"`
	ChangedBy      *uuid.UUID            `json:"changed_by"`
}

type changeStatusRequest struct {
	ChangedBy *uuid.UUID `json:"changed_by"`
	Reason    string     `json:"reason"`
}
