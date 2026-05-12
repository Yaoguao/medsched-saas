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
	r.Get("/notify/{id}", h.Get)
	r.Delete("/notify/{id}", h.Cancel)
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

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.CancelNotification(r.Context(), id); err != nil {
		http.Error(w, "cancel failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
