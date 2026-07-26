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
	ID        int64             `gorm:"primaryKey" json:"id"`
	SeatID    int64             `gorm:"not null;index" json:"seat_id"`
	EventID   int64             `gorm:"not null;index" json:"event_id"`
	UserID    int64             `gorm:"not null" json:"user_id"`
	Status    ReservationStatus `gorm:"type:varchar(16);not null;default:'HELD';index:idx_res_status_expiry,priority:1" json:"status"`
	ExpiresAt time.Time         `gorm:"not null;index:idx_res_status_expiry,priority:2" json:"expires_at"`
	CreatedAt time.Time         `json:"created_at"`
}
