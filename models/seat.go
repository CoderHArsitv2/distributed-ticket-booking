package models

import "time"

// SeatStatus is the inventory state of a single seat.
type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatReserved  SeatStatus = "RESERVED"
	SeatBooked    SeatStatus = "BOOKED"
	SeatBlocked   SeatStatus = "BLOCKED"
)

// Seat is a single inventory row. The Version column powers optimistic locking.
type Seat struct {
	ID            int64      `db:"id"            json:"id"`
	EventID       int64      `db:"event_id"      json:"event_id"`
	SeatNumber    string     `db:"seat_number"   json:"seat_number"`
	Section       string     `db:"section"       json:"section"`
	Price         int64      `db:"price"         json:"price"` // minor units (e.g. cents)
	Status        SeatStatus `db:"status"        json:"status"`
	ReservedUntil *time.Time `db:"reserved_until" json:"reserved_until,omitempty"`
	Version       int64      `db:"version"       json:"version"`
	CreatedAt     time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"    json:"updated_at"`
}
