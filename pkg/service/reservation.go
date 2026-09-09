// Package service holds business logic: the booking flow orchestration on top
// of the model stores and locking layers.
package service

import (
	"context"
	"time"

	"distributed-ticket-booking/models"
	"distributed-ticket-booking/pkg/locking"
)

// ReservationService orchestrates the seat-hold -> reserve -> book flow using
// the configured locking strategy.
type ReservationService struct {
	locker  locking.SeatLocker
	seats   models.SeatStore
	holds   models.ReservationStore
	holdTTL time.Duration
}

// NewReservationService wires the reservation engine.
func NewReservationService(
	locker locking.SeatLocker,
	seats models.SeatStore,
	holds models.ReservationStore,
	holdTTL time.Duration,
) *ReservationService {
	return &ReservationService{locker: locker, seats: seats, holds: holds, holdTTL: holdTTL}
}

// HoldSeat acquires a lock via the configured strategy, which flips the seat to
// RESERVED, then persists a Reservation row tracking the hold expiry. On success
// it returns the created reservation.
func (s *ReservationService) HoldSeat(ctx context.Context, eventID, seatID, userID int64) (*models.Reservation, error) {
	if _, err := s.locker.Acquire(ctx, seatID, userID); err != nil {
		return nil, err
	}

	res := &models.Reservation{
		SeatID:    seatID,
		EventID:   eventID,
		UserID:    userID,
		Status:    models.ReservationHeld,
		ExpiresAt: time.Now().Add(s.holdTTL),
	}
	if err := s.holds.Create(ctx, res); err != nil {
		// Best-effort rollback of the seat hold so we don't strand it.
		_ = s.locker.Release(ctx, seatID, "")
		return nil, err
	}
	return res, nil
}

// Strategy reports which locking strategy is active.
func (s *ReservationService) Strategy() string { return s.locker.Name() }
