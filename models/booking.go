package models

import "time"

// BookingStatus tracks the lifecycle of a finalized order.
type BookingStatus string

const (
	BookingPending   BookingStatus = "PENDING"
	BookingConfirmed BookingStatus = "CONFIRMED"
	BookingFailed    BookingStatus = "FAILED"
	BookingRefunded  BookingStatus = "REFUNDED"
)

// Booking is a finalized customer order with a unique reference code.
type Booking struct {
	ID          int64         `db:"id"           json:"id"`
	EventID     int64         `db:"event_id"     json:"event_id"`
	UserID      int64         `db:"user_id"      json:"user_id"`
	Reference   string        `db:"reference"    json:"reference"`
	TotalAmount int64         `db:"total_amount" json:"total_amount"` // minor units
	Status      BookingStatus `db:"status"       json:"status"`
	PaymentRef  string        `db:"payment_ref"  json:"payment_ref,omitempty"`
	CreatedAt   time.Time     `db:"created_at"   json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"   json:"updated_at"`
}

// BookingSeat is the junction row snapshotting purchase price per seat.
type BookingSeat struct {
	ID            int64 `db:"id"             json:"id"`
	BookingID     int64 `db:"booking_id"     json:"booking_id"`
	SeatID        int64 `db:"seat_id"        json:"seat_id"`
	PriceSnapshot int64 `db:"price_snapshot" json:"price_snapshot"` // minor units
}
