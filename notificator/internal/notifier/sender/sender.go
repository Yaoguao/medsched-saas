package sender

import (
	"context"
	"notificator/internal/models"
	"time"

	"github.com/sirupsen/logrus"
)

// Sender interface for pluggable channels
type Sender interface {
	Send(ctx context.Context, n models.Notification) error
}

type EmailSender struct {
	log *logrus.Logger
}

func NewEmailSender(log *logrus.Logger) *EmailSender {
	return &EmailSender{log: log}
}

func (s *EmailSender) Send(ctx context.Context, n models.Notification) error {
	time.Sleep(100 * time.Millisecond)

	s.log.WithFields(logrus.Fields{
		"channel":   "email",
		"id":        n.ID.String(),
		"recipient": n.Recipient,
		"subject":   n.Subject,
	}).Info("sending email (stub)")

	return nil
}

type TelegramSender struct {
	log *logrus.Logger
}

func NewTelegramSender(log *logrus.Logger) *TelegramSender {
	return &TelegramSender{log: log}
}

func (s *TelegramSender) Send(ctx context.Context, n models.Notification) error {
	time.Sleep(100 * time.Millisecond)

	s.log.WithFields(logrus.Fields{
		"channel":   "telegram",
		"id":        n.ID.String(),
		"recipient": n.Recipient,
	}).Info("sending telegram (stub)")

	return nil
}

type SMSSender struct {
	log *logrus.Logger
}

func NewSMSSender(log *logrus.Logger) *SMSSender {
	return &SMSSender{log: log}
}

func (s *SMSSender) Send(ctx context.Context, n models.Notification) error {
	time.Sleep(100 * time.Millisecond)

	s.log.WithFields(logrus.Fields{
		"channel":   "sms",
		"id":        n.ID.String(),
		"recipient": n.Recipient,
	}).Info("sending sms (stub)")

	return nil
}
