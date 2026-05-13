package router

import (
	"base/internal/http-server/handlers/appointments"
	"base/internal/http-server/handlers/doctors"
	"base/internal/http-server/handlers/health"
	"base/internal/http-server/handlers/notifications"
	"base/internal/http-server/handlers/patients"
	"base/internal/http-server/handlers/predictions"
	"base/internal/http-server/handlers/reports"
	"base/internal/http-server/handlers/schedules"
	"base/internal/http-server/handlers/users"
	"base/internal/service"

	"github.com/gin-gonic/gin"
)

func New(svc *service.Service, env string) *gin.Engine {
	if env != "dev" && env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	health.New().Register(r)

	api := r.Group("/api/v1")
	health.New().Register(api)
	users.New(svc).Register(api)
	patients.New(svc).Register(api)
	doctors.New(svc).Register(api)
	schedules.New(svc).Register(api)
	appointments.New(svc).Register(api)
	notifications.New(svc).Register(api)
	predictions.New(svc).Register(api)
	reports.New(svc).Register(api)

	return r
}
