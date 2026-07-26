package models

import "time"

// ReservationStatus tracks the lifecycle of an ephemeral seat hold.
type ReservationStatus string

const (
	ReservationHeld      ReservationStatus = "HELD"
	ReservationConfirmed ReservationStatus = "CONFIRMED"
	ReservationExpired   ReservationStatus = "EXPIRED"
	ReservationReleased  ReservationStatus = "RELEASED"
)

// Reservation is a temporary hold on a seat, valid until ExpiresAt.
type Reservation struct {
	ID        int64             `db:"id"         json:"id"`
	SeatID    int64             `db:"seat_id"    json:"seat_id"`
	EventID   int64             `db:"event_id"   json:"event_id"`
	UserID    int64             `db:"user_id"    json:"user_id"`
	Status    ReservationStatus `db:"status"     json:"status"`
	ExpiresAt time.Time         `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time         `db:"created_at" json:"created_at"`
}
