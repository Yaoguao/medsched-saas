package predictions

import (
	"net/http"
	"strconv"
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
	r.POST("/ml/predictions", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleITSupport), h.create)
	r.GET("/ml/predictions/:prediction_id", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager, models.UserRoleITSupport), h.get)
	r.GET("/ml/model-info", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleManager, models.UserRoleITSupport), h.modelInfo)
	r.POST("/appointments/:appointment_id/predict-no-show", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.predictAppointmentNoShow)
}

func (h *Handler) RegisterTrainingDataset(r gin.IRouter, serviceToken string) {
	r.GET(
		"/ml/training/no-show-dataset",
		authmiddleware.RequireServiceTokenOrRoles(
			h.svc,
			serviceToken,
			models.UserRoleAdmin,
			models.UserRoleManager,
			models.UserRoleITSupport,
		),
		h.trainingNoShowDataset,
	)
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
	if !h.canAccessAppointment(c, id) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
		return
	}
	prediction, err := h.svc.PredictAppointmentNoShow(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, responseFromModel(prediction))
}

func (h *Handler) canAccessAppointment(c *gin.Context, appointmentID uuid.UUID) bool {
	user, ok := authmiddleware.CurrentUser(c)
	if !ok {
		return false
	}
	if user.Role == models.UserRoleAdmin || user.Role == models.UserRoleRegistrator {
		return true
	}
	appointment, err := h.svc.GetAppointment(c.Request.Context(), appointmentID)
	if err != nil {
		return false
	}
	return user.Role == models.UserRoleDoctor && user.Doctor != nil && appointment.DoctorID == user.Doctor.ID
}

func (h *Handler) trainingNoShowDataset(c *gin.Context) {
	rows, err := h.svc.ExportNoShowTrainingDataset(c.Request.Context(), queryInt(c, "limit"))
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, trainingDatasetResponseFromRows(rows))
}

func (h *Handler) modelInfo(c *gin.Context) {
	info, err := h.svc.MLModelInfo(c.Request.Context())
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, modelInfoResponseFromService(info))
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
