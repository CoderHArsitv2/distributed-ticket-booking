package models

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

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

// bookingStore is the GORM-backed BookingStore.
type bookingStore struct{ db *gorm.DB }

// NewBookingStore returns a GORM-backed BookingStore.
func NewBookingStore(db *gorm.DB) BookingStore { return &bookingStore{db: db} }

func (r *bookingStore) Create(ctx context.Context, b *Booking, seats []BookingSeat) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(b).Error; err != nil {
			return err
		}
		for i := range seats {
			seats[i].BookingID = b.ID
		}
		if len(seats) > 0 {
			if err := tx.Create(&seats).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *bookingStore) GetByReference(ctx context.Context, ref string) (*Booking, error) {
	var b Booking
	err := r.db.WithContext(ctx).Preload("Seats").Where("reference = ?", ref).First(&b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}
