package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"distributed-ticket-booking/locking"
	"distributed-ticket-booking/service"
)

// ReservationController handles seat-hold and booking endpoints.
type ReservationController struct {
	svc *service.ReservationService
}

// NewReservationController constructs the reservation controller.
func NewReservationController(svc *service.ReservationService) *ReservationController {
	return &ReservationController{svc: svc}
}

type holdRequest struct {
	EventID int64 `json:"event_id"`
	SeatID  int64 `json:"seat_id" binding:"required"`
	UserID  int64 `json:"user_id" binding:"required"`
}

// Hold places a temporary hold on a seat using the configured lock strategy.
// POST /api/v1/reservations
func (rc *ReservationController) Hold(c *gin.Context) {
	var req holdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "seat_id and user_id are required")
		return
	}

	reservation, err := rc.svc.HoldSeat(c.Request.Context(), req.EventID, req.SeatID, req.UserID)
	if err != nil {
		if errors.Is(err, locking.ErrSeatUnavailable) {
			respondError(c, http.StatusConflict, "seat is not available")
			return
		}
		respondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":      "HELD",
		"reservation": reservation,
		"strategy":    rc.svc.Strategy(),
	})
}
