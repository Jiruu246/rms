package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles GET /health — a liveness probe for external health checkers.
// It lives outside the /api base path, so it is intentionally not in the Swagger spec.
// It doesn't touch the database: a DB outage shouldn't make orchestrators restart the server.
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
