package patients

import (
	"net/http"
	"time"

	"base/internal/domain/models"
	appointmenthandler "base/internal/http-server/handlers/appointments"
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
	r.GET("/patients", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleManager, models.UserRoleDoctor), h.list)
	r.POST("/patients", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator), h.create)
	r.GET("/patients/:patient_id", authmiddleware.RequirePatientAccess("patient_id"), h.get)
	r.PATCH("/patients/:patient_id", authmiddleware.RequirePatientAccess("patient_id"), h.update)
	r.PATCH("/patients/:patient_id/user", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator), h.linkUser)
	r.POST("/patients/:patient_id/ehr-link", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator), h.linkEHR)
	r.GET("/patients/:patient_id/ehr-preview", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.previewEHR)
	r.POST("/patients/:patient_id/ehr-sync", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.syncEHR)
	r.GET("/patients/:patient_id/profile", authmiddleware.RequirePatientAccess("patient_id"), h.getMedicalProfile)
	r.PATCH("/patients/:patient_id/profile", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.updateMedicalProfile)
	r.GET("/patients/:patient_id/appointments", authmiddleware.RequirePatientAccess("patient_id"), h.listAppointments)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}

	patient := models.Patient{
		FullName:      req.FullName,
		Gender:        req.Gender,
		BirthDate:     req.BirthDate,
		Age:           req.Age,
		Phone:         req.Phone,
		Email:         req.Email,
		Neighbourhood: req.Neighbourhood,
		Scholarship:   req.Scholarship,
		ExternalEHRID: req.ExternalEHRID,
	}
	if req.MedicalProfile != nil {
		patient.MedicalProfile = &models.PatientMedicalProfile{
			ChronicConditions: req.MedicalProfile.ChronicConditions,
			RiskFactors:       req.MedicalProfile.RiskFactors,
			Allergies:         req.MedicalProfile.Allergies,
			Notes:             req.MedicalProfile.Notes,
			HasHypertension:   req.MedicalProfile.HasHypertension,
			HasDiabetes:       req.MedicalProfile.HasDiabetes,
			HasAlcoholism:     req.MedicalProfile.HasAlcoholism,
			HasHandicap:       req.MedicalProfile.HasHandicap,
		}
	}

	created, err := h.svc.CreatePatient(c.Request.Context(), patient)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, patientResponseFromModel(created))
}

func (h *Handler) list(c *gin.Context) {
	page := common.Pagination(c)
	patients, err := h.svc.ListPatients(c.Request.Context(), page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, patientResponsesFromModels(patients), page)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	patient, err := h.svc.GetPatient(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, patientResponseFromModel(patient))
}

func (h *Handler) update(c *gin.Context) {
	id, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	var req updateRequest
	if !common.BindJSON(c, &req) {
		return
	}
	patient, err := h.svc.UpdatePatient(c.Request.Context(), id, service.PatientPatch{
		FullName:      req.FullName,
		Gender:        req.Gender,
		BirthDate:     req.BirthDate,
		Age:           req.Age,
		Phone:         req.Phone,
		Email:         req.Email,
		Neighbourhood: req.Neighbourhood,
		Scholarship:   req.Scholarship,
		ExternalEHRID: req.ExternalEHRID,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, patientResponseFromModel(patient))
}

func (h *Handler) linkUser(c *gin.Context) {
	patientID, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}

	var req linkUserRequest
	if !common.BindJSON(c, &req) {
		return
	}

	patient, err := h.svc.LinkPatientUser(c.Request.Context(), patientID, req.UserID)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, patientResponseFromModel(patient))
}

func (h *Handler) syncEHR(c *gin.Context) {
	patientID, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	var req ehrProviderRequest
	if !common.BindJSON(c, &req) {
		return
	}
	user, ok := authmiddleware.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	result, err := h.svc.SyncPatientWithEHR(c.Request.Context(), patientID, service.SyncPatientEHRRequest{
		ProviderID: req.ProviderID,
		Fields:     req.Fields,
		ActorID:    &user.ID,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, ehrSyncResponseFromService(result))
}

func (h *Handler) linkEHR(c *gin.Context) {
	patientID, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	var req linkEHRRequest
	if !common.BindJSON(c, &req) {
		return
	}
	link, err := h.svc.LinkPatientToEHR(
		c.Request.Context(),
		patientID,
		req.ProviderID,
		req.ExternalID,
	)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, externalEntityLinkResponseFromModel(link))
}

func (h *Handler) previewEHR(c *gin.Context) {
	patientID, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	providerID, err := uuid.Parse(c.Query("provider_id"))
	if err != nil {
		common.RespondError(c, http.StatusBadRequest, "invalid provider_id")
		return
	}
	preview, err := h.svc.PreviewPatientEHR(c.Request.Context(), patientID, providerID)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, ehrPreviewResponseFromService(preview))
}

func (h *Handler) getMedicalProfile(c *gin.Context) {
	id, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	profile, err := h.svc.GetPatientMedicalProfile(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, medicalProfileResponseFromModel(&profile))
}

func (h *Handler) updateMedicalProfile(c *gin.Context) {
	id, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}

	var req updateMedicalProfileRequest
	if !common.BindJSON(c, &req) {
		return
	}

	profile, err := h.svc.UpdatePatientMedicalProfile(c.Request.Context(), id, service.PatientMedicalProfilePatch{
		ChronicConditions: req.ChronicConditions,
		RiskFactors:       req.RiskFactors,
		Allergies:         req.Allergies,
		Notes:             req.Notes,
		HasHypertension:   req.HasHypertension,
		HasDiabetes:       req.HasDiabetes,
		HasAlcoholism:     req.HasAlcoholism,
		HasHandicap:       req.HasHandicap,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, medicalProfileResponseFromModel(&profile))
}

func (h *Handler) listAppointments(c *gin.Context) {
	id, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	page := common.Pagination(c)
	appointments, err := h.svc.ListPatientAppointments(c.Request.Context(), id, page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, appointmenthandler.ResponsesFromModels(appointments), page)
}

type createRequest struct {
	FullName       string                 `json:"full_name" binding:"required"`
	Gender         string                 `json:"gender"`
	BirthDate      time.Time              `json:"birth_date"`
	Age            int                    `json:"age"`
	Phone          string                 `json:"phone"`
	Email          string                 `json:"email"`
	Neighbourhood  string                 `json:"neighbourhood"`
	Scholarship    bool                   `json:"scholarship"`
	ExternalEHRID  string                 `json:"external_ehr_id"`
	MedicalProfile *medicalProfileRequest `json:"medical_profile"`
}

type updateRequest struct {
	FullName      *string    `json:"full_name"`
	Gender        *string    `json:"gender"`
	BirthDate     *time.Time `json:"birth_date"`
	Age           *int       `json:"age"`
	Phone         *string    `json:"phone"`
	Email         *string    `json:"email"`
	Neighbourhood *string    `json:"neighbourhood"`
	Scholarship   *bool      `json:"scholarship"`
	ExternalEHRID *string    `json:"external_ehr_id"`
}

type linkUserRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type ehrProviderRequest struct {
	ProviderID uuid.UUID `json:"provider_id" binding:"required"`
	Fields     []string  `json:"fields" binding:"required,min=1"`
}

type linkEHRRequest struct {
	ProviderID uuid.UUID `json:"provider_id" binding:"required"`
	ExternalID string    `json:"external_id" binding:"required"`
}

type medicalProfileRequest struct {
	ChronicConditions string `json:"chronic_conditions"`
	RiskFactors       string `json:"risk_factors"`
	Allergies         string `json:"allergies"`
	Notes             string `json:"notes"`
	HasHypertension   bool   `json:"has_hypertension"`
	HasDiabetes       bool   `json:"has_diabetes"`
	HasAlcoholism     bool   `json:"has_alcoholism"`
	HasHandicap       bool   `json:"has_handicap"`
}

type updateMedicalProfileRequest struct {
	ChronicConditions *string `json:"chronic_conditions"`
	RiskFactors       *string `json:"risk_factors"`
	Allergies         *string `json:"allergies"`
	Notes             *string `json:"notes"`
	HasHypertension   *bool   `json:"has_hypertension"`
	HasDiabetes       *bool   `json:"has_diabetes"`
	HasAlcoholism     *bool   `json:"has_alcoholism"`
	HasHandicap       *bool   `json:"has_handicap"`
}
