package models

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

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

// eventStore is the GORM-backed EventStore.
type eventStore struct{ db *gorm.DB }

// NewEventStore returns a GORM-backed EventStore.
func NewEventStore(db *gorm.DB) EventStore { return &eventStore{db: db} }

func (r *eventStore) GetByID(ctx context.Context, id int64) (*Event, error) {
	var e Event
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *eventStore) List(ctx context.Context) ([]Event, error) {
	var events []Event
	err := r.db.WithContext(ctx).Order("sale_starts_at asc").Find(&events).Error
	return events, err
}

func (r *eventStore) Create(ctx context.Context, e *Event) error {
	return r.db.WithContext(ctx).Create(e).Error
}
