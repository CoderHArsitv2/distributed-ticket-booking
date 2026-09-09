package models

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

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

// reservationStore is the GORM-backed ReservationStore.
type reservationStore struct{ db *gorm.DB }

// NewReservationStore returns a GORM-backed ReservationStore.
func NewReservationStore(db *gorm.DB) ReservationStore { return &reservationStore{db: db} }

func (r *reservationStore) Create(ctx context.Context, res *Reservation) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *reservationStore) GetActiveBySeat(ctx context.Context, seatID int64) (*Reservation, error) {
	var res Reservation
	err := r.db.WithContext(ctx).
		Where("seat_id = ? AND status = ? AND expires_at > ?", seatID, ReservationHeld, time.Now()).
		Order("expires_at desc").
		First(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &res, nil
}

func (r *reservationStore) MarkExpired(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&Reservation{}).
		Where("status = ? AND expires_at < ?", ReservationHeld, time.Now()).
		Update("status", ReservationExpired)
	return res.RowsAffected, res.Error
}
