package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/health", h.health)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
