package notification

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"notificator/internal/models"
	"notificator/internal/service"
)

type Handler struct {
	svc *service.NotifierService
	log *logrus.Logger
}

func NewHandler(svc *service.NotifierService, log *logrus.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Register(r chi.Router) {
	r.Post("/notify", h.Create)
	r.Get("/notify", h.List)
	r.Get("/notify/{id}", h.Get)
	r.Post("/notify/{id}/send", h.SendNow)
	r.Delete("/notify/{id}", h.Cancel)

	r.Post("/templates", h.UpsertTemplate)
	r.Get("/templates", h.ListTemplates)
	r.Get("/templates/{id}", h.GetTemplate)
	r.Put("/templates/{id}", h.UpsertTemplateByID)
	r.Delete("/templates/{id}", h.DeleteTemplate)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var n models.Notification
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if n.SendAt.IsZero() {
		n.SendAt = time.Now().Add(1 * time.Minute)
	}

	id, err := h.svc.CreateNotification(r.Context(), n)
	if err != nil {
		h.log.WithError(err).Error("create notification failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.svc.GetNotification(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusCreated, map[string]string{"id": id.String()})
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	notifications, err := h.svc.ListNotifications(r.Context())
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, notifications)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	n, err := h.svc.GetNotification(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) SendNow(w http.ResponseWriter, r *http.Request) {
	id, ok := notificationID(w, r)
	if !ok {
		return
	}

	if err := h.svc.SendNow(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	n, err := h.svc.GetNotification(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, ok := notificationID(w, r)
	if !ok {
		return
	}

	if err := h.svc.CancelNotification(r.Context(), id); err != nil {
		http.Error(w, "cancel failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpsertTemplate(w http.ResponseWriter, r *http.Request) {
	var t models.NotificationTemplate
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	h.upsertTemplate(w, r, t)
}

func (h *Handler) UpsertTemplateByID(w http.ResponseWriter, r *http.Request) {
	var t models.NotificationTemplate
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	t.ID = chi.URLParam(r, "id")
	h.upsertTemplate(w, r, t)
}

func (h *Handler) upsertTemplate(w http.ResponseWriter, r *http.Request, t models.NotificationTemplate) {
	template, err := h.svc.UpsertTemplate(r.Context(), t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, template)
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.svc.ListTemplates(r.Context())
	if err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, templates)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	template, err := h.svc.GetTemplate(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, template)
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteTemplate(r.Context(), chi.URLParam(r, "id")); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func notificationID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
