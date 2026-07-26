// Package service holds business logic: the booking flow orchestration on top
// of the repository and locking layers.
package service

import (
	"context"

	"distributed-ticket-booking/locking"
	"distributed-ticket-booking/repository"
)

// ReservationService orchestrates the seat-hold -> reserve -> book flow using
// the configured locking strategy.
type ReservationService struct {
	locker locking.SeatLocker
	seats  repository.SeatRepository
	holds  repository.ReservationRepository
}

// NewReservationService wires the reservation engine.
func NewReservationService(
	locker locking.SeatLocker,
	seats repository.SeatRepository,
	holds repository.ReservationRepository,
) *ReservationService {
	return &ReservationService{locker: locker, seats: seats, holds: holds}
}

// HoldSeat acquires a lock via the configured strategy and creates a temporary
// RESERVED hold for the user.
//
// TODO(phase-4/5): acquire the lock, persist the hold with an expiry, and
// return the reservation so the caller can proceed to checkout.
func (s *ReservationService) HoldSeat(ctx context.Context, eventID, seatID, userID int64) error {
	_, err := s.locker.Acquire(ctx, seatID, userID)
	if err != nil {
		return err
	}
	// TODO: create reservation row with expires_at = now + HoldTTL.
	return nil
}

// Strategy reports which locking strategy is active.
func (s *ReservationService) Strategy() string { return s.locker.Name() }
