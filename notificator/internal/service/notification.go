package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"notificator/internal/config"
	"notificator/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type StorageRedis interface {
	Get(ctx context.Context, key string) (string, error)
	SetWithTTL(ctx context.Context, key string, val models.Notification, ttl time.Duration) error
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

// CreateNotification stores notification in redis and publishes it to RabbitMQ for processing.
func (s *NotifierService) CreateNotification(ctx context.Context, n models.Notification) (uuid.UUID, error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	n.Status = "pending"
	n.Attempts = 0

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

func (s *NotifierService) GetNotification(ctx context.Context, id uuid.UUID) (models.Notification, error) {
	var n models.Notification
	key := fmt.Sprintf("notif:%s", id.String())
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
	key := fmt.Sprintf("notif:%s", id.String())

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

	key := fmt.Sprintf("notif:%s", n.ID.String())
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
			if n.Attempts > maxRetries {

				if err2 := s.redis.UpdateStatus(ctx, fmt.Sprintf("notif:%s", id.String()), "failed"); err2 != nil {
					s.log.WithError(err2).Error("failed to update status to failed")
				}

				return err
			}

			ttl := time.Until(n.SendAt) + 24*time.Hour
			if ttl <= 0 {
				ttl = 24 * time.Hour
			}
			if err2 := s.redis.SetWithTTL(ctx, fmt.Sprintf("notif:%s", id.String()), n, ttl); err2 != nil {
				s.log.WithError(err2).Error("failed to update attempts")
			}

			backoff := initialDelay * time.Duration(math.Pow(2, float64(n.Attempts-1)))
			time.Sleep(backoff)
			continue
		}

		// success
		if err := s.redis.UpdateStatus(ctx, fmt.Sprintf("notif:%s", id.String()), "sent"); err != nil {
			s.log.WithError(err).Error("failed to update status to sent")
		}
		return nil
	}
}

func (s *NotifierService) ListNotificationKeys(ctx context.Context) ([]string, error) {
	return s.redis.Keys(ctx, "notif:*")
}
