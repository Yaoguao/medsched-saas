package doctors

import (
	"net/http"

	"base/internal/domain/models"
	"base/internal/http-server/handlers/common"
	authmiddleware "base/internal/http-server/middleware/auth"
	"base/internal/repository"
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
	r.GET("/doctors/specialties", h.listSpecialties)
	r.POST("/doctors", authmiddleware.RequireRoles(models.UserRoleAdmin), h.create)
	r.GET("/doctors/:doctor_id", h.get)
	r.PATCH("/doctors/:doctor_id", authmiddleware.RequireRoles(models.UserRoleAdmin), h.update)
	r.PATCH("/doctors/:doctor_id/user", authmiddleware.RequireRoles(models.UserRoleAdmin), h.linkUser)
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
	c.JSON(http.StatusCreated, responseFromModel(created))
}

func (h *Handler) list(c *gin.Context) {
	page := common.Pagination(c)
	doctors, err := h.svc.ListDoctors(c.Request.Context(), page, repository.DoctorListFilter{
		Search:     c.Query("search"),
		Specialty:  c.Query("specialty"),
		Sort:       c.Query("sort"),
		ActiveOnly: c.Query("active_only") == "true",
	})
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, responsesFromModels(doctors), page)
}

func (h *Handler) listSpecialties(c *gin.Context) {
	specialties, err := h.svc.ListDoctorSpecialties(c.Request.Context(), c.Query("active_only") == "true")
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, specialties)
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
	c.JSON(http.StatusOK, responseFromModel(doctor))
}

func (h *Handler) update(c *gin.Context) {
	id, ok := common.PathUUID(c, "doctor_id")
	if !ok {
		return
	}

	var req updateRequest
	if !common.BindJSON(c, &req) {
		return
	}

	doctor, err := h.svc.UpdateDoctor(c.Request.Context(), id, service.DoctorPatch{
		UserID:          req.UserID,
		FullName:        req.FullName,
		Specialty:       req.Specialty,
		Cabinet:         req.Cabinet,
		ExperienceYears: req.ExperienceYears,
		Status:          req.Status,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, responseFromModel(doctor))
}

func (h *Handler) linkUser(c *gin.Context) {
	doctorID, ok := common.PathUUID(c, "doctor_id")
	if !ok {
		return
	}

	var req linkUserRequest
	if !common.BindJSON(c, &req) {
		return
	}

	doctor, err := h.svc.LinkDoctorUser(c.Request.Context(), doctorID, req.UserID)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, responseFromModel(doctor))
}

type createRequest struct {
	UserID          *uuid.UUID `json:"user_id"`
	FullName        string     `json:"full_name" binding:"required"`
	Specialty       string     `json:"specialty" binding:"required"`
	Cabinet         string     `json:"cabinet"`
	ExperienceYears int        `json:"experience_years"`
	Status          string     `json:"status"`
}

type updateRequest struct {
	UserID          *uuid.UUID `json:"user_id"`
	FullName        *string    `json:"full_name"`
	Specialty       *string    `json:"specialty"`
	Cabinet         *string    `json:"cabinet"`
	ExperienceYears *int       `json:"experience_years"`
	Status          *string    `json:"status"`
}

type linkUserRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}
