package router

import (
	"base/internal/http-server/handlers/appointments"
	"base/internal/http-server/handlers/doctors"
	"base/internal/http-server/handlers/health"
	"base/internal/http-server/handlers/integrations"
	"base/internal/http-server/handlers/mlmanagement"
	"base/internal/http-server/handlers/notifications"
	"base/internal/http-server/handlers/patients"
	"base/internal/http-server/handlers/predictions"
	"base/internal/http-server/handlers/reports"
	"base/internal/http-server/handlers/schedules"
	"base/internal/http-server/handlers/users"
	authmiddleware "base/internal/http-server/middleware/auth"
	"base/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func New(svc *service.Service, env string, notifierCallbackSecret string, mlDatasetServiceToken ...string) *gin.Engine {
	if env != "dev" && env != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), cors())

	health.New().Register(r)

	api := r.Group("/api/v1")
	health.New().Register(api)

	usersHandler := users.New(svc)
	usersHandler.RegisterPublic(api)
	notificationsHandler := notifications.New(svc, notifierCallbackSecret)
	notificationsHandler.RegisterCallbacks(api)
	predictionsHandler := predictions.New(svc)
	datasetServiceToken := ""
	if len(mlDatasetServiceToken) > 0 {
		datasetServiceToken = mlDatasetServiceToken[0]
	}
	predictionsHandler.RegisterTrainingDataset(api, datasetServiceToken)

	protected := api.Group("", authmiddleware.Authenticate(svc))
	usersHandler.RegisterProtected(protected)
	patients.New(svc).Register(protected)
	doctors.New(svc).Register(protected)
	schedules.New(svc).Register(protected)
	appointments.New(svc).Register(protected)
	notificationsHandler.Register(protected)
	predictionsHandler.Register(protected)
	mlmanagement.New(svc).Register(protected)
	integrations.New(svc).Register(protected)
	reports.New(svc).Register(protected)

	return r
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
