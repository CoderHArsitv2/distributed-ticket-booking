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
	ID          int64         `gorm:"primaryKey" json:"id"`
	EventID     int64         `gorm:"not null;index" json:"event_id"`
	UserID      int64         `gorm:"not null" json:"user_id"`
	Reference   string        `gorm:"type:varchar(64);not null;uniqueIndex" json:"reference"`
	TotalAmount int64         `gorm:"not null" json:"total_amount"` // minor units
	Status      BookingStatus `gorm:"type:varchar(16);not null;default:'PENDING'" json:"status"`
	PaymentRef  string        `gorm:"type:varchar(128)" json:"payment_ref,omitempty"`
	Seats       []BookingSeat `gorm:"foreignKey:BookingID" json:"seats,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// BookingSeat is the junction row snapshotting purchase price per seat.
type BookingSeat struct {
	ID            int64 `gorm:"primaryKey" json:"id"`
	BookingID     int64 `gorm:"not null;uniqueIndex:uq_booking_seat,priority:1" json:"booking_id"`
	SeatID        int64 `gorm:"not null;uniqueIndex:uq_booking_seat,priority:2" json:"seat_id"`
	PriceSnapshot int64 `gorm:"not null" json:"price_snapshot"` // minor units
}
