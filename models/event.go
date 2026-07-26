package models

import "time"

// EventStatus is the lifecycle state of an event.
type EventStatus string

const (
	EventUpcoming  EventStatus = "UPCOMING"
	EventOnSale    EventStatus = "ON_SALE"
	EventSoldOut   EventStatus = "SOLD_OUT"
	EventCancelled EventStatus = "CANCELLED"
)

// Event is the canonical record for a bookable event.
type Event struct {
	ID             int64       `db:"id"              json:"id"`
	Name           string      `db:"name"            json:"name"`
	Venue          string      `db:"venue"           json:"venue"`
	TotalSeats     int         `db:"total_seats"     json:"total_seats"`
	AvailableSeats int         `db:"available_seats" json:"available_seats"`
	SaleStartsAt   time.Time   `db:"sale_starts_at"  json:"sale_starts_at"`
	Status         EventStatus `db:"status"          json:"status"`
	CreatedAt      time.Time   `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at"      json:"updated_at"`
}
