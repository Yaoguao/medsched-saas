package reports

import (
	"net/http"

	"base/internal/http-server/handlers/common"
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
	r.GET("/reports/appointments", h.appointmentStatus)
	r.GET("/reports/no-show", h.appointmentStatus)
	r.GET("/reports/dashboard", h.appointmentStatus)
}

func (h *Handler) appointmentStatus(c *gin.Context) {
	from, err := common.ParseOptionalTime(c.Query("from"))
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid from")
		return
	}
	to, err := common.ParseOptionalTime(c.Query("to"))
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid to")
		return
	}
	report, err := h.svc.AppointmentStatusReport(c.Request.Context(), from, to)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, report)
}
