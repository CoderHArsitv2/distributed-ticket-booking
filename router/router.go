// Package router assembles the HTTP routes and middleware. It uses the stdlib
// net/http mux for now; swap for chi/gin in Phase 6 if richer routing is needed.
package router

import (
	"log/slog"
	"net/http"

	"distributed-ticket-booking/controllers"
	"distributed-ticket-booking/middleware"
)

// Deps are the controllers the router wires into routes.
type Deps struct {
	Health      *controllers.HealthController
	Reservation *controllers.ReservationController
	Logger      *slog.Logger
}

// New builds the fully-wired HTTP handler.
func New(d Deps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", d.Health.Healthz)

	// API v1
	mux.HandleFunc("POST /api/v1/reservations", d.Reservation.Hold)
	// TODO(phase-6): GET /api/v1/events, GET /api/v1/events/{id}/seats,
	// POST /api/v1/bookings, etc.

	return middleware.Chain(mux,
		middleware.Recover(d.Logger),
		middleware.Logger(d.Logger),
	)
}
