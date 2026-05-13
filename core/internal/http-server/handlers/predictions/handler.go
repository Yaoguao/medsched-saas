package predictions

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
	r.POST("/ml/predictions", h.create)
	r.GET("/ml/predictions/:prediction_id", h.get)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}
	prediction := models.PredictionResult{
		AppointmentID: req.AppointmentID,
		Probability:   req.Probability,
		RiskLevel:     req.RiskLevel,
		ModelVersion:  req.ModelVersion,
		RequestedAt:   req.RequestedAt,
		ReceivedAt:    req.ReceivedAt,
	}
	created, err := h.svc.CreatePredictionResult(c.Request.Context(), prediction)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := common.PathUUID(c, "prediction_id")
	if !ok {
		return
	}
	prediction, err := h.svc.GetPredictionResult(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, prediction)
}

type createRequest struct {
	AppointmentID uuid.UUID                  `json:"appointment_id" binding:"required"`
	Probability   float64                    `json:"probability" binding:"required"`
	RiskLevel     models.PredictionRiskLevel `json:"risk_level"`
	ModelVersion  string                     `json:"model_version" binding:"required"`
	RequestedAt   time.Time                  `json:"requested_at"`
	ReceivedAt    time.Time                  `json:"received_at"`
}
