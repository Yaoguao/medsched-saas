package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"notificator/internal/config"
	"notificator/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type StorageRedis interface {
	Get(ctx context.Context, key string) (string, error)
	SetWithTTL(ctx context.Context, key string, val models.Notification, ttl time.Duration) error
	SetDeadLetter(ctx context.Context, key string, val models.DeadLetter, ttl time.Duration) error
	SetTemplate(ctx context.Context, key string, val models.NotificationTemplate) error
	UpdateNotification(ctx context.Context, key string, notification models.Notification) error
	UpdateStatus(ctx context.Context, key, status string) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
}

type RabbitMQClient interface {
	Publish(body []byte) error
}

type Sender interface {
	Send(ctx context.Context, n models.Notification) error
}

type NotifierService struct {
	redis   StorageRedis
	mq      RabbitMQClient
	senders map[string]Sender
	cfg     *config.Config
	log     *logrus.Logger
}

func NewNotifierService(ss StorageRedis, mq RabbitMQClient, senders map[string]Sender, cfg *config.Config, log *logrus.Logger) *NotifierService {
	return &NotifierService{
		redis:   ss,
		mq:      mq,
		senders: senders,
		cfg:     cfg,
		log:     log,
	}
}

func (s *NotifierService) EnsureDefaultTemplates(ctx context.Context) error {
	defaults := []models.NotificationTemplate{
		{
			ID:      "appointment_reminder",
			Channel: "email",
			Subject: "Напоминание о приеме",
			Body:    "Здравствуйте, {{patient_name}}. Напоминаем о приеме {{appointment_at}} у врача {{doctor_name}}.",
		},
		{
			ID:      "appointment_confirmation",
			Channel: "email",
			Subject: "Запись подтверждена",
			Body:    "Здравствуйте, {{patient_name}}. Ваша запись на {{appointment_at}} подтверждена.",
		},
		{
			ID:      "appointment_cancelled",
			Channel: "email",
			Subject: "Запись отменена",
			Body:    "Здравствуйте, {{patient_name}}. Ваша запись на {{appointment_at}} отменена.",
		},
	}

	for _, tmpl := range defaults {
		if _, err := s.GetTemplate(ctx, tmpl.ID); err == nil {
			continue
		}
		if _, err := s.UpsertTemplate(ctx, tmpl); err != nil {
			return err
		}
	}
	return nil
}

// CreateNotification stores notification in redis and publishes it to RabbitMQ for processing.
func (s *NotifierService) CreateNotification(ctx context.Context, n models.Notification) (uuid.UUID, error) {
	if err := s.prepareNotification(ctx, &n); err != nil {
		return uuid.Nil, err
	}
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Status == "" {
		n.Status = "pending"
	}
	now := time.Now()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	n.UpdatedAt = now

	ttl := time.Until(n.SendAt) + 24*time.Hour
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	key := fmt.Sprintf("notif:%s", n.ID.String())
	if err := s.redis.SetWithTTL(ctx, key, n, ttl); err != nil {
		return uuid.Nil, err
	}

	b, err := json.Marshal(n)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.mq.Publish(b); err != nil {
		s.log.WithError(err).Error("failed publish to mq")
		return n.ID, err
	}

	return n.ID, nil
}

func (s *NotifierService) UpsertTemplate(ctx context.Context, t models.NotificationTemplate) (models.NotificationTemplate, error) {
	t.ID = strings.TrimSpace(t.ID)
	t.Channel = strings.TrimSpace(strings.ToLower(t.Channel))
	if t.ID == "" {
		return models.NotificationTemplate{}, errors.New("template id is required")
	}
	if t.Body == "" {
		return models.NotificationTemplate{}, errors.New("template body is required")
	}
	if err := s.validateChannel(t.Channel); err != nil {
		return models.NotificationTemplate{}, err
	}

	now := time.Now()
	existing, err := s.GetTemplate(ctx, t.ID)
	if err == nil {
		t.CreatedAt = existing.CreatedAt
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now

	if err := s.redis.SetTemplate(ctx, templateKey(t.ID), t); err != nil {
		return models.NotificationTemplate{}, err
	}
	return t, nil
}

func (s *NotifierService) GetTemplate(ctx context.Context, id string) (models.NotificationTemplate, error) {
	var t models.NotificationTemplate
	data, err := s.redis.Get(ctx, templateKey(id))
	if err != nil {
		return t, err
	}
	if err := json.Unmarshal([]byte(data), &t); err != nil {
		return t, err
	}
	return t, nil
}

func (s *NotifierService) DeleteTemplate(ctx context.Context, id string) error {
	return s.redis.Delete(ctx, templateKey(id))
}

func (s *NotifierService) ListTemplates(ctx context.Context) ([]models.NotificationTemplate, error) {
	keys, err := s.redis.Keys(ctx, "template:*")
	if err != nil {
		return nil, err
	}
	templates := make([]models.NotificationTemplate, 0, len(keys))
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		var t models.NotificationTemplate
		if err := json.Unmarshal([]byte(data), &t); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (s *NotifierService) GetNotification(ctx context.Context, id uuid.UUID) (models.Notification, error) {
	var n models.Notification
	key := notificationKey(id)
	data, err := s.redis.Get(ctx, key)
	if err != nil {
		return n, err
	}
	if err := json.Unmarshal([]byte(data), &n); err != nil {
		return n, err
	}
	return n, nil
}

func (s *NotifierService) CancelNotification(ctx context.Context, id uuid.UUID) error {
	key := notificationKey(id)

	exists, err := s.redis.Exists(ctx, key)

	if err != nil {
		return err
	}
	if !exists {
		return errors.New("notification not found")
	}
	return s.redis.UpdateStatus(ctx, key, "cancelled")
}

// ScheduleNotification is called by the consumer: it ensures notification exists in redis and schedules a send.
func (s *NotifierService) ScheduleNotification(ctx context.Context, n models.Notification) error {
	if err := s.prepareNotification(ctx, &n); err != nil {
		return err
	}
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Status == "" {
		n.Status = "pending"
	}
	now := time.Now()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = now
	}
	n.UpdatedAt = now

	key := notificationKey(n.ID)
	exists, err := s.redis.Exists(ctx, key)
	if err != nil {
		return err
	}

	if !exists {
		ttl := time.Until(n.SendAt) + 24*time.Hour
		if ttl <= 0 {
			ttl = 24 * time.Hour
		}
		if err := s.redis.SetWithTTL(ctx, key, n, ttl); err != nil {
			return err
		}
	}

	go func(id uuid.UUID, sendAt time.Time) {
		delay := time.Until(sendAt)
		if delay > 0 {
			time.Sleep(delay)
		}
		if err := s.SendNow(context.Background(), id); err != nil {
			s.log.WithError(err).WithField("id", id.String()).Error("send failed after scheduled run")
		}
	}(n.ID, n.SendAt)

	return nil
}

// SendNow attempts to send notification and handles retries with exponential backoff.
func (s *NotifierService) SendNow(ctx context.Context, id uuid.UUID) error {
	n, err := s.GetNotification(ctx, id)
	if err != nil {
		return err
	}
	if n.Status == "cancelled" || n.Status == "sent" {
		return nil
	}

	sender, ok := s.senders[n.Channel]
	if !ok {
		return fmt.Errorf("no sender for channel %s", n.Channel)
	}

	maxRetries := s.cfg.Notifier.MaxRetries
	initialDelay := s.cfg.Notifier.RetryInitialDelay

	for {
		if err := sender.Send(ctx, n); err != nil {
			n.Attempts++
			n.LastError = err.Error()
			n.UpdatedAt = time.Now()
			n.AttemptHistory = appendDeliveryAttempt(n.AttemptHistory, n.UpdatedAt, "failed", err.Error())
			if n.Attempts > maxRetries {
				n.Status = "failed"
				_ = s.redis.UpdateNotification(ctx, notificationKey(id), n)
				s.storeDeadLetter(context.Background(), n, err.Error())
				s.notifyCoreDeliveryStatus(context.Background(), n)
				return err
			}

			ttl := time.Until(n.SendAt) + 24*time.Hour
			if ttl <= 0 {
				ttl = 24 * time.Hour
			}
			if err2 := s.redis.SetWithTTL(ctx, notificationKey(id), n, ttl); err2 != nil {
				s.log.WithError(err2).Error("failed to update attempts")
			}

			backoff := initialDelay * time.Duration(math.Pow(2, float64(n.Attempts-1)))
			time.Sleep(backoff)
			continue
		}

		// success
		n.Status = "sent"
		n.LastError = ""
		n.UpdatedAt = time.Now()
		n.AttemptHistory = appendDeliveryAttempt(n.AttemptHistory, n.UpdatedAt, "sent", "")
		_ = s.redis.UpdateNotification(ctx, notificationKey(id), n)
		s.notifyCoreDeliveryStatus(context.Background(), n)

		n.Status = "delivered"
		n.UpdatedAt = time.Now()
		_ = s.redis.UpdateNotification(ctx, notificationKey(id), n)
		s.notifyCoreDeliveryStatus(context.Background(), n)
		return nil
	}
}

func appendDeliveryAttempt(history []models.DeliveryAttempt, attemptedAt time.Time, status string, err string) []models.DeliveryAttempt {
	if attemptedAt.IsZero() {
		attemptedAt = time.Now()
	}
	return append(history, models.DeliveryAttempt{
		AttemptedAt: attemptedAt,
		Status:      status,
		Error:       err,
	})
}

func (s *NotifierService) storeDeadLetter(ctx context.Context, n models.Notification, reason string) {
	deadLetter := models.DeadLetter{
		ID:             n.ID,
		NotificationID: n.ID,
		ExternalID:     n.ExternalID,
		Channel:        n.Channel,
		Recipient:      n.Recipient,
		Reason:         reason,
		Attempts:       n.Attempts,
		AttemptHistory: n.AttemptHistory,
		Notification:   n,
		CreatedAt:      time.Now(),
	}
	if err := s.redis.SetDeadLetter(ctx, deadLetterKey(n.ID), deadLetter, 7*24*time.Hour); err != nil {
		s.log.WithError(err).WithField("id", n.ID.String()).Error("failed to store dead letter")
	}
}

func (s *NotifierService) notifyCoreDeliveryStatus(ctx context.Context, n models.Notification) {
	baseURL := strings.TrimRight(strings.TrimSpace(s.cfg.Notifier.CoreCallbackURL), "/")
	if baseURL == "" || n.ID == uuid.Nil {
		return
	}

	status := strings.ToUpper(strings.TrimSpace(n.Status))
	payload := deliveryStatusCallback{
		Status:       status,
		ErrorMessage: n.LastError,
	}
	if status == "SENT" || status == "DELIVERED" {
		now := n.UpdatedAt
		if now.IsZero() {
			now = time.Now()
		}
		payload.SentAt = &now
	}

	body, err := json.Marshal(payload)
	if err != nil {
		s.log.WithError(err).WithField("id", n.ID.String()).Error("failed to marshal delivery callback")
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/notifications/"+n.ID.String()+"/delivery-status", bytes.NewReader(body))
	if err != nil {
		s.log.WithError(err).WithField("id", n.ID.String()).Error("failed to build delivery callback")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if secret := strings.TrimSpace(s.cfg.Notifier.CoreCallbackSecret); secret != "" {
		req.Header.Set("X-Notifier-Secret", secret)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.log.WithError(err).WithField("id", n.ID.String()).Error("delivery callback failed")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		s.log.WithFields(logrus.Fields{
			"id":     n.ID.String(),
			"status": resp.StatusCode,
		}).Error("delivery callback returned non-success status")
	}
}

type deliveryStatusCallback struct {
	Status       string     `json:"status"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

func (s *NotifierService) ListNotifications(ctx context.Context) ([]models.Notification, error) {
	keys, err := s.redis.Keys(ctx, "notif:*")
	if err != nil {
		return nil, err
	}
	notifications := make([]models.Notification, 0, len(keys))
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		var n models.Notification
		if err := json.Unmarshal([]byte(data), &n); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func (s *NotifierService) ListDeadLetters(ctx context.Context) ([]models.DeadLetter, error) {
	keys, err := s.redis.Keys(ctx, "deadletter:*")
	if err != nil {
		return nil, err
	}
	deadLetters := make([]models.DeadLetter, 0, len(keys))
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		var deadLetter models.DeadLetter
		if err := json.Unmarshal([]byte(data), &deadLetter); err != nil {
			return nil, err
		}
		deadLetters = append(deadLetters, deadLetter)
	}
	return deadLetters, nil
}

func (s *NotifierService) GetDeadLetter(ctx context.Context, id uuid.UUID) (models.DeadLetter, error) {
	var deadLetter models.DeadLetter
	data, err := s.redis.Get(ctx, deadLetterKey(id))
	if err != nil {
		return deadLetter, err
	}
	if err := json.Unmarshal([]byte(data), &deadLetter); err != nil {
		return deadLetter, err
	}
	return deadLetter, nil
}

func (s *NotifierService) HandleCommand(ctx context.Context, body []byte) error {
	var cmd models.Command
	if err := json.Unmarshal(body, &cmd); err == nil && cmd.Type != "" {
		switch cmd.Type {
		case "notification.create":
			if cmd.Notification == nil {
				return errors.New("notification command requires notification payload")
			}
			return s.ScheduleNotification(ctx, *cmd.Notification)
		case "notification.schedule":
			if cmd.Notification == nil {
				return errors.New("notification command requires notification payload")
			}
			return s.ScheduleNotification(ctx, *cmd.Notification)
		case "notification.cancel":
			if cmd.ID == uuid.Nil {
				return errors.New("notification cancel command requires id")
			}
			return s.CancelNotification(ctx, cmd.ID)
		case "notification.send_now":
			if cmd.ID == uuid.Nil {
				return errors.New("notification send_now command requires id")
			}
			return s.SendNow(ctx, cmd.ID)
		case "template.upsert":
			if cmd.Template == nil {
				return errors.New("template command requires template payload")
			}
			_, err := s.UpsertTemplate(ctx, *cmd.Template)
			return err
		default:
			return fmt.Errorf("unsupported command type %s", cmd.Type)
		}
	}

	var n models.Notification
	if err := json.Unmarshal(body, &n); err != nil {
		return err
	}
	return s.ScheduleNotification(ctx, n)
}

func (s *NotifierService) prepareNotification(ctx context.Context, n *models.Notification) error {
	n.Channel = strings.TrimSpace(strings.ToLower(n.Channel))
	if err := s.validateChannel(n.Channel); err != nil {
		return err
	}
	if n.SendAt.IsZero() {
		n.SendAt = time.Now()
	}
	if n.TemplateID != "" && strings.TrimSpace(n.Message) == "" {
		t, err := s.GetTemplate(ctx, n.TemplateID)
		if err != nil {
			return fmt.Errorf("template %s not found: create it via POST /templates before creating this notification", n.TemplateID)
		}
		if t.Channel != "" && t.Channel != n.Channel {
			return fmt.Errorf("template %s is configured for channel %s", t.ID, t.Channel)
		}
		n.Subject = renderTemplate(t.Subject, n.Variables)
		n.Message = renderTemplate(t.Body, n.Variables)
	}
	if strings.TrimSpace(n.Message) == "" {
		return errors.New("message or template_id is required")
	}
	return nil
}

func (s *NotifierService) validateChannel(channel string) error {
	if channel == "" {
		return errors.New("channel is required")
	}
	for _, allowed := range s.cfg.Notifier.Channels {
		if channel == strings.ToLower(strings.TrimSpace(allowed)) {
			if _, ok := s.senders[channel]; ok {
				return nil
			}
			return fmt.Errorf("channel %s is configured but has no sender", channel)
		}
	}
	return fmt.Errorf("unsupported channel %s", channel)
}

func renderTemplate(template string, variables map[string]string) string {
	rendered := template
	for key, value := range variables {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}
	return rendered
}

func notificationKey(id uuid.UUID) string {
	return fmt.Sprintf("notif:%s", id.String())
}

func deadLetterKey(id uuid.UUID) string {
	return fmt.Sprintf("deadletter:%s", id.String())
}

func templateKey(id string) string {
	return fmt.Sprintf("template:%s", strings.TrimSpace(id))
}
