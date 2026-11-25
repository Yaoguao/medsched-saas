package reports

import (
	"net/http"
	"time"

	"base/internal/domain/models"
	"base/internal/http-server/handlers/common"
	authmiddleware "base/internal/http-server/middleware/auth"
	"base/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r gin.IRouter) {
	allowed := authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager, models.UserRoleITSupport)
	r.GET("/reports/appointments", allowed, h.appointments)
	r.GET("/reports/no-show", allowed, h.noShow)
	r.GET("/reports/load", allowed, h.load)
	r.GET("/reports/dashboard", allowed, h.dashboard)
}

func (h *Handler) appointments(c *gin.Context) {
	from, to, ok := reportRange(c)
	if !ok {
		return
	}
	report, err := h.svc.AppointmentReport(c.Request.Context(), from, to)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *Handler) noShow(c *gin.Context) {
	from, to, ok := reportRange(c)
	if !ok {
		return
	}
	report, err := h.svc.NoShowReport(c.Request.Context(), from, to)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *Handler) load(c *gin.Context) {
	from, to, ok := reportRange(c)
	if !ok {
		return
	}
	report, err := h.svc.LoadReport(c.Request.Context(), from, to)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *Handler) dashboard(c *gin.Context) {
	from, to, ok := reportRange(c)
	if !ok {
		return
	}
	report, err := h.svc.DashboardReport(c.Request.Context(), from, to)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, report)
}

func reportRange(c *gin.Context) (from, to time.Time, ok bool) {
	var err error
	from, err = common.ParseOptionalTime(c.Query("from"))
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid from")
		return time.Time{}, time.Time{}, false
	}
	to, err = common.ParseOptionalTime(c.Query("to"))
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid to")
		return time.Time{}, time.Time{}, false
	}
	return from, to, true
}
