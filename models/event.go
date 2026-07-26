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
	ID             int64       `gorm:"primaryKey" json:"id"`
	Name           string      `gorm:"not null" json:"name"`
	Venue          string      `gorm:"not null" json:"venue"`
	TotalSeats     int         `gorm:"not null" json:"total_seats"`
	AvailableSeats int         `gorm:"not null" json:"available_seats"`
	SaleStartsAt   time.Time   `gorm:"not null" json:"sale_starts_at"`
	Status         EventStatus `gorm:"type:varchar(16);not null;default:'UPCOMING';index" json:"status"`
	Seats          []Seat      `gorm:"foreignKey:EventID" json:"seats,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}
