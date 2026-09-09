package models

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

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

// seatStore is the GORM-backed SeatStore.
type seatStore struct{ db *gorm.DB }

// NewSeatStore returns a GORM-backed SeatStore.
func NewSeatStore(db *gorm.DB) SeatStore { return &seatStore{db: db} }

func (r *seatStore) GetByID(ctx context.Context, id int64) (*Seat, error) {
	var s Seat
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *seatStore) ListByEvent(ctx context.Context, eventID int64) ([]Seat, error) {
	var seats []Seat
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("section asc, seat_number asc").
		Find(&seats).Error
	return seats, err
}

func (r *seatStore) ReserveOptimistic(ctx context.Context, seatID, expectedVersion int64) (bool, error) {
	res := r.db.WithContext(ctx).Model(&Seat{}).
		Where("id = ? AND version = ? AND status = ?", seatID, expectedVersion, SeatAvailable).
		Updates(map[string]any{
			"status":  SeatReserved,
			"version": gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *seatStore) ReleaseExpired(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&Seat{}).
		Where("status = ? AND reserved_until IS NOT NULL AND reserved_until < ?",
			SeatReserved, time.Now()).
		Updates(map[string]any{
			"status":         SeatAvailable,
			"reserved_until": nil,
			"version":        gorm.Expr("version + 1"),
		})
	return res.RowsAffected, res.Error
}
