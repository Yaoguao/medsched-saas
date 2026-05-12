package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"

	"notificator/internal/config"
	"notificator/internal/http-server/handlers/notification"
	"notificator/internal/lib/logger/logr"
	"notificator/internal/models"
	"notificator/internal/notifier/sender"
	svc "notificator/internal/service"
	cons "notificator/internal/storage/rabbitmq/consumer"
	prod "notificator/internal/storage/rabbitmq/producer"
	redisStore "notificator/internal/storage/redis"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.MustLoad()

	log := logr.InitLogrusLog(cfg.Env, os.Stdout)
	log.Info("starting notificator")

	// init redis storage
	rds := redisStore.NewStorageRedis(cfg)

	// init rabbitmq producer
	producer, err := prod.NewProducer(cfg.RabbitMQ.URL, cfg.RabbitMQ.ReadyQueue)
	if err != nil {
		log.WithError(err).Fatal("failed to init producer")
	}
	defer producer.Close()

	// init senders
	senders := map[string]svc.Sender{
		"email":    sender.NewEmailSender(log),
		"telegram": sender.NewTelegramSender(log),
	}

	// init service
	notifier := svc.NewNotifierService(rds, producer, senders, cfg, log)

	// start consumer to schedule incoming messages
	consumer, err := cons.NewConsumer(cfg.RabbitMQ.URL, cfg.RabbitMQ.ReadyQueue)
	if err != nil {
		log.WithError(err).Fatal("failed to init consumer")
	}
	defer consumer.Close()

	// handler for consumer: unmarshal and schedule
	_ = consumer.Start(func(body []byte) error {
		var n models.Notification
		if err := json.Unmarshal(body, &n); err != nil {
			log.WithError(err).Error("invalid message body")
			return nil
		}
		if err := notifier.ScheduleNotification(context.Background(), n); err != nil {
			log.WithError(err).Error("failed to schedule notification")
		}
		return nil
	})

	// HTTP server
	r := chi.NewRouter()
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	notifHandler := notification.NewHandler(notifier, log)
	notifHandler.Register(r)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Infof("http server listening on %s", cfg.HTTPServer.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("http server failed")
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Info("shutting down")
}
