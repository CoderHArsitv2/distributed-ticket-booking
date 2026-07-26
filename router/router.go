// Package router assembles the Gin engine: routes plus middleware.
package router

import (
	"github.com/gin-gonic/gin"

	"distributed-ticket-booking/controllers"
)

// Deps are the controllers the router wires into routes.
type Deps struct {
	Health      *controllers.HealthController
	Reservation *controllers.ReservationController
}

// New builds the fully-wired Gin engine (with Logger + Recovery middleware).
func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", d.Health.Healthz)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/reservations", d.Reservation.Hold)
		// TODO(phase-6): GET /events, GET /events/:id/seats, POST /bookings, etc.
	}

	return r
}
