package predictions

import (
	"net/http"
	"strconv"
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
	r.GET("/ml/training/no-show-dataset", h.trainingNoShowDataset)
	r.POST("/appointments/:appointment_id/predict-no-show", h.predictAppointmentNoShow)
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
	c.JSON(http.StatusCreated, responseFromModel(created))
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
	c.JSON(http.StatusOK, responseFromModel(prediction))
}

func (h *Handler) predictAppointmentNoShow(c *gin.Context) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}
	prediction, err := h.svc.PredictAppointmentNoShow(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, responseFromModel(prediction))
}

func (h *Handler) trainingNoShowDataset(c *gin.Context) {
	rows, err := h.svc.ExportNoShowTrainingDataset(c.Request.Context(), queryInt(c, "limit"))
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, trainingDatasetResponseFromRows(rows))
}

type createRequest struct {
	AppointmentID uuid.UUID                  `json:"appointment_id" binding:"required"`
	Probability   float64                    `json:"probability" binding:"required"`
	RiskLevel     models.PredictionRiskLevel `json:"risk_level"`
	ModelVersion  string                     `json:"model_version" binding:"required"`
	RequestedAt   time.Time                  `json:"requested_at"`
	ReceivedAt    time.Time                  `json:"received_at"`
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
