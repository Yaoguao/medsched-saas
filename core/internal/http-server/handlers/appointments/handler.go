package appointments

import (
	"net/http"

	"base/internal/domain/models"
	"base/internal/http-server/handlers/common"
	authmiddleware "base/internal/http-server/middleware/auth"
	"base/internal/lib/pagination"
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
	r.POST("/appointments/:appointment_id/confirm", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.confirm)
	r.POST("/appointments/:appointment_id/cancel", h.cancel)
	r.POST("/appointments/:appointment_id/check-in", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.checkIn)
	r.POST("/appointments/:appointment_id/no-show", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.noShow)
	r.POST("/appointments/:appointment_id/reschedule", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator), h.reschedule)
	r.GET("/appointments/:appointment_id/reschedule-proposals", h.listRescheduleProposals)
	r.POST("/appointments/:appointment_id/reschedule-proposals", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleDoctor), h.createRescheduleProposal)
	r.GET("/patients/:patient_id/reschedule-proposals", authmiddleware.RequirePatientAccess("patient_id"), h.listPatientRescheduleProposals)
	r.POST("/reschedule-proposals/:proposal_id/accept", h.acceptRescheduleProposal)
	r.POST("/reschedule-proposals/:proposal_id/decline", h.declineRescheduleProposal)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}
	user, ok := authmiddleware.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if !canCreateAppointmentForPatient(user, req.PatientID) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
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
	user, ok := authmiddleware.CurrentUser(c)
	if !ok {
		common.RespondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	appointments, err := h.listAppointmentsForUser(c, user, page)
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
	if !canAccessAppointment(c, appointment) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
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

func (h *Handler) reschedule(c *gin.Context) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}

	var req rescheduleRequest
	if !common.BindJSON(c, &req) {
		return
	}
	appointmentForAccess, err := h.svc.GetAppointment(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	if !canAccessAppointment(c, appointmentForAccess) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
		return
	}

	appointment, err := h.svc.RescheduleAppointment(c.Request.Context(), id, service.RescheduleAppointmentRequest{
		TimeSlotID: req.TimeSlotID,
		ChangedBy:  req.ChangedBy,
		Reason:     req.Reason,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, ResponseFromModel(appointment))
}

func (h *Handler) createRescheduleProposal(c *gin.Context) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}

	var req createRescheduleProposalRequest
	if !common.BindJSON(c, &req) {
		return
	}
	appointment, err := h.svc.GetAppointment(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	if !canAccessAppointment(c, appointment) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
		return
	}

	proposal, err := h.svc.CreateRescheduleProposal(c.Request.Context(), id, service.CreateRescheduleProposalRequest{
		TimeSlotID:  req.TimeSlotID,
		RequestedBy: req.RequestedBy,
		Reason:      req.Reason,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, RescheduleProposalResponseFromModel(proposal))
}

func (h *Handler) listRescheduleProposals(c *gin.Context) {
	id, ok := common.PathUUID(c, "appointment_id")
	if !ok {
		return
	}
	page := common.Pagination(c)
	appointment, err := h.svc.GetAppointment(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	if !canAccessAppointment(c, appointment) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
		return
	}
	proposals, err := h.svc.ListAppointmentRescheduleProposals(c.Request.Context(), id, page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, RescheduleProposalResponsesFromModels(proposals), page)
}

func (h *Handler) listPatientRescheduleProposals(c *gin.Context) {
	id, ok := common.PathUUID(c, "patient_id")
	if !ok {
		return
	}
	page := common.Pagination(c)
	proposals, err := h.svc.ListPatientRescheduleProposals(c.Request.Context(), id, page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, RescheduleProposalResponsesFromModels(proposals), page)
}

func (h *Handler) acceptRescheduleProposal(c *gin.Context) {
	h.respondRescheduleProposal(c, true)
}

func (h *Handler) declineRescheduleProposal(c *gin.Context) {
	h.respondRescheduleProposal(c, false)
}

func (h *Handler) respondRescheduleProposal(c *gin.Context, accept bool) {
	id, ok := common.PathUUID(c, "proposal_id")
	if !ok {
		return
	}
	var req respondRescheduleProposalRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if !common.BindJSON(c, &req) {
			return
		}
	}
	existingProposal, err := h.svc.GetRescheduleProposal(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	if !canAccessAppointment(c, existingProposal.Appointment) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
		return
	}
	serviceReq := service.RespondRescheduleProposalRequest{
		RespondedBy: req.RespondedBy,
		Reason:      req.Reason,
	}
	var (
		proposal models.AppointmentRescheduleProposal
	)
	if accept {
		proposal, err = h.svc.AcceptRescheduleProposal(c.Request.Context(), id, serviceReq)
	} else {
		proposal, err = h.svc.DeclineRescheduleProposal(c.Request.Context(), id, serviceReq)
	}
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, RescheduleProposalResponseFromModel(proposal))
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
	appointmentForAccess, err := h.svc.GetAppointment(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	if !canAccessAppointment(c, appointmentForAccess) {
		common.RespondError(c, http.StatusForbidden, "forbidden")
		return
	}
	appointment, err := h.svc.ChangeAppointmentStatus(c.Request.Context(), id, status, req.ChangedBy, req.Reason)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, ResponseFromModel(appointment))
}

func (h *Handler) listAppointmentsForUser(c *gin.Context, user models.User, page pagination.Page) ([]models.Appointment, error) {
	if isOperationalRole(user.Role) {
		return h.svc.ListAppointments(c.Request.Context(), page)
	}
	if user.Role == models.UserRoleDoctor && user.Doctor != nil {
		return h.svc.ListDoctorAppointments(c.Request.Context(), user.Doctor.ID, page)
	}
	if user.Role == models.UserRolePatient && user.Patient != nil {
		return h.svc.ListPatientAppointments(c.Request.Context(), user.Patient.ID, page)
	}
	return nil, nil
}

func canAccessAppointment(c *gin.Context, appointment models.Appointment) bool {
	user, ok := authmiddleware.CurrentUser(c)
	if !ok {
		return false
	}
	if isOperationalRole(user.Role) {
		return true
	}
	if user.Role == models.UserRoleDoctor && user.Doctor != nil {
		return appointment.DoctorID == user.Doctor.ID
	}
	if user.Role == models.UserRolePatient && user.Patient != nil {
		return appointment.PatientID == user.Patient.ID
	}
	return false
}

func canCreateAppointmentForPatient(user models.User, patientID uuid.UUID) bool {
	if user.Role == models.UserRoleAdmin || user.Role == models.UserRoleRegistrator {
		return true
	}
	return user.Role == models.UserRolePatient && user.Patient != nil && user.Patient.ID == patientID
}

func isOperationalRole(role models.UserRole) bool {
	return role == models.UserRoleAdmin ||
		role == models.UserRoleRegistrator ||
		role == models.UserRoleManager ||
		role == models.UserRoleITSupport
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

type rescheduleRequest struct {
	TimeSlotID uuid.UUID  `json:"time_slot_id" binding:"required"`
	ChangedBy  *uuid.UUID `json:"changed_by"`
	Reason     string     `json:"reason"`
}

type createRescheduleProposalRequest struct {
	TimeSlotID  uuid.UUID  `json:"time_slot_id" binding:"required"`
	RequestedBy *uuid.UUID `json:"requested_by"`
	Reason      string     `json:"reason"`
}

type respondRescheduleProposalRequest struct {
	RespondedBy *uuid.UUID `json:"responded_by"`
	Reason      string     `json:"reason"`
}
