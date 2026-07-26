package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthController exposes liveness/readiness endpoints.
type HealthController struct{}

// NewHealthController constructs the health controller.
func NewHealthController() *HealthController { return &HealthController{} }

// Healthz reports process liveness.
func (h *HealthController) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
