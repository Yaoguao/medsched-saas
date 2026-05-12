package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func NewLoggerMiddleware(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log.WithField("component", "middleware/logger").Info("logger middleware enabled")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			entry := log.WithFields(logrus.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"request_id":  middleware.GetReqID(r.Context()),
			})

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				entry.WithFields(logrus.Fields{
					"status":   ww.Status(),
					"bytes":    ww.BytesWritten(),
					"duration": time.Since(start).String(),
				}).Info("request completed")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
