package schedules

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
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/schedules/import/preview", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.previewImport)
	r.POST("/schedules/import/apply", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.applyImport)
	r.GET("/schedules/:doctor_id", h.listDoctorSchedules)
	r.POST("/schedules", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.createSchedule)
	r.PATCH("/schedules/:schedule_id", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.updateSchedule)
	r.POST("/schedules/:schedule_id/slots", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.createTimeSlot)
	r.PATCH("/slots/:slot_id", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.updateTimeSlot)
	r.POST("/slots/:slot_id/block", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.blockTimeSlot)
	r.POST("/slots/:slot_id/unblock", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager), h.unblockTimeSlot)
}

func (h *Handler) previewImport(c *gin.Context) {
	var req service.ScheduleImportRequest
	if !common.BindJSON(c, &req) {
		return
	}
	result, err := h.svc.PreviewScheduleImport(c.Request.Context(), req)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) applyImport(c *gin.Context) {
	var req service.ScheduleImportRequest
	if !common.BindJSON(c, &req) {
		return
	}
	result, err := h.svc.ApplyScheduleImport(c.Request.Context(), req)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) createSchedule(c *gin.Context) {
	var req createScheduleRequest
	if !common.BindJSON(c, &req) {
		return
	}
	schedule := models.Schedule{
		DoctorID:      req.DoctorID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		WorkingDays:   req.WorkingDays,
		WorkStartTime: req.WorkStartTime,
		WorkEndTime:   req.WorkEndTime,
		Status:        req.Status,
	}
	created, err := h.svc.CreateSchedule(c.Request.Context(), schedule)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, scheduleResponseFromModel(created))
}

func (h *Handler) listDoctorSchedules(c *gin.Context) {
	doctorID, ok := common.PathUUID(c, "doctor_id")
	if !ok {
		return
	}
	page := common.Pagination(c)
	schedules, err := h.svc.GetDoctorSchedules(c.Request.Context(), doctorID, page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, scheduleResponsesFromModels(schedules), page)
}

func (h *Handler) updateSchedule(c *gin.Context) {
	id, ok := common.PathUUID(c, "schedule_id")
	if !ok {
		return
	}
	var req updateScheduleRequest
	if !common.BindJSON(c, &req) {
		return
	}
	schedule, err := h.svc.UpdateSchedule(c.Request.Context(), id, service.SchedulePatch{
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		WorkingDays:   req.WorkingDays,
		WorkStartTime: req.WorkStartTime,
		WorkEndTime:   req.WorkEndTime,
		Status:        req.Status,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, scheduleResponseFromModel(schedule))
}

func (h *Handler) createTimeSlot(c *gin.Context) {
	scheduleID, ok := common.PathUUID(c, "schedule_id")
	if !ok {
		return
	}
	var req createTimeSlotRequest
	if !common.BindJSON(c, &req) {
		return
	}
	slot := models.TimeSlot{
		ScheduleID: scheduleID,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
		Status:     req.Status,
		IsBlocked:  req.IsBlocked,
	}
	created, err := h.svc.CreateTimeSlot(c.Request.Context(), slot)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, timeSlotResponseFromModel(created))
}

func (h *Handler) updateTimeSlot(c *gin.Context) {
	id, ok := common.PathUUID(c, "slot_id")
	if !ok {
		return
	}
	var req updateTimeSlotRequest
	if !common.BindJSON(c, &req) {
		return
	}
	slot, err := h.svc.UpdateTimeSlot(c.Request.Context(), id, service.TimeSlotPatch{
		StartAt:   req.StartAt,
		EndAt:     req.EndAt,
		Status:    req.Status,
		IsBlocked: req.IsBlocked,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, timeSlotResponseFromModel(slot))
}

func (h *Handler) blockTimeSlot(c *gin.Context) {
	h.setSlotBlock(c, true)
}

func (h *Handler) unblockTimeSlot(c *gin.Context) {
	h.setSlotBlock(c, false)
}

func (h *Handler) setSlotBlock(c *gin.Context, blocked bool) {
	id, ok := common.PathUUID(c, "slot_id")
	if !ok {
		return
	}
	var req setSlotBlockRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if !common.BindJSON(c, &req) {
			return
		}
	}
	slot, err := h.svc.BlockTimeSlot(c.Request.Context(), id, blocked, req.Reason)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, timeSlotResponseFromModel(slot))
}

type createScheduleRequest struct {
	DoctorID      uuid.UUID `json:"doctor_id" binding:"required"`
	StartDate     time.Time `json:"start_date" binding:"required"`
	EndDate       time.Time `json:"end_date" binding:"required"`
	WorkingDays   string    `json:"working_days"`
	WorkStartTime string    `json:"work_start_time"`
	WorkEndTime   string    `json:"work_end_time"`
	Status        string    `json:"status"`
}

type updateScheduleRequest struct {
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	WorkingDays   *string    `json:"working_days"`
	WorkStartTime *string    `json:"work_start_time"`
	WorkEndTime   *string    `json:"work_end_time"`
	Status        *string    `json:"status"`
}

type createTimeSlotRequest struct {
	StartAt   time.Time         `json:"start_at" binding:"required"`
	EndAt     time.Time         `json:"end_at" binding:"required"`
	Status    models.SlotStatus `json:"status"`
	IsBlocked bool              `json:"is_blocked"`
}

type updateTimeSlotRequest struct {
	StartAt   *time.Time         `json:"start_at"`
	EndAt     *time.Time         `json:"end_at"`
	Status    *models.SlotStatus `json:"status"`
	IsBlocked *bool              `json:"is_blocked"`
}

type setSlotBlockRequest struct {
	Reason string `json:"reason"`
}
