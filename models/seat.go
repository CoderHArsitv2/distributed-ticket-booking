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
	ID            int64      `gorm:"primaryKey" json:"id"`
	EventID       int64      `gorm:"not null;uniqueIndex:uq_seat_event_number,priority:1;index:idx_seat_event" json:"event_id"`
	SeatNumber    string     `gorm:"not null;uniqueIndex:uq_seat_event_number,priority:2" json:"seat_number"`
	Section       string     `gorm:"not null" json:"section"`
	Price         int64      `gorm:"not null" json:"price"` // minor units (e.g. cents)
	Status        SeatStatus `gorm:"type:varchar(16);not null;default:'AVAILABLE';index:idx_seat_status_expiry,priority:1" json:"status"`
	ReservedUntil *time.Time `gorm:"index:idx_seat_status_expiry,priority:2" json:"reserved_until,omitempty"`
	Version       int64      `gorm:"not null;default:0" json:"version"` // optimistic locking
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
