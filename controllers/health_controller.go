package controllers

import "net/http"

// HealthController exposes liveness/readiness endpoints.
type HealthController struct{}

// NewHealthController constructs the health controller.
func NewHealthController() *HealthController { return &HealthController{} }

// Healthz reports process liveness.
func (c *HealthController) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
