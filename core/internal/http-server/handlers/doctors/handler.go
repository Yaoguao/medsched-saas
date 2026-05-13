package doctors

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
	r.GET("/doctors", h.list)
	r.POST("/doctors", h.create)
	r.GET("/doctors/:doctor_id", h.get)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}
	doctor := models.Doctor{
		UserID:          req.UserID,
		FullName:        req.FullName,
		Specialty:       req.Specialty,
		Cabinet:         req.Cabinet,
		ExperienceYears: req.ExperienceYears,
		Status:          req.Status,
	}
	created, err := h.svc.CreateDoctor(c.Request.Context(), doctor)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) list(c *gin.Context) {
	page := common.Pagination(c)
	doctors, err := h.svc.ListDoctors(c.Request.Context(), page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, doctors, page)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := common.PathUUID(c, "doctor_id")
	if !ok {
		return
	}
	doctor, err := h.svc.GetDoctor(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, doctor)
}

type createRequest struct {
	UserID          *uuid.UUID `json:"user_id"`
	FullName        string     `json:"full_name" binding:"required"`
	Specialty       string     `json:"specialty" binding:"required"`
	Cabinet         string     `json:"cabinet"`
	ExperienceYears int        `json:"experience_years"`
	Status          string     `json:"status"`
}
