package controllers

import (
	"encoding/json"
	"net/http"

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
	SeatID  int64 `json:"seat_id"`
	UserID  int64 `json:"user_id"`
}

// Hold places a temporary hold on a seat using the configured lock strategy.
// POST /api/v1/reservations
func (c *ReservationController) Hold(w http.ResponseWriter, r *http.Request) {
	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.SeatID == 0 || req.UserID == 0 {
		writeError(w, http.StatusBadRequest, "seat_id and user_id are required")
		return
	}

	if err := c.svc.HoldSeat(r.Context(), req.EventID, req.SeatID, req.UserID); err != nil {
		// TODO(phase-6): map domain errors (ErrSeatUnavailable -> 409) precisely.
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status":   "HELD",
		"seat_id":  req.SeatID,
		"strategy": c.svc.Strategy(),
	})
}
