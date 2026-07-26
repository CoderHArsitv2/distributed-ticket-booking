package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"distributed-ticket-booking/models"
)

// --- Event ---

type gormEventRepo struct{ db *gorm.DB }

// NewEventRepository returns a GORM-backed EventRepository.
func NewEventRepository(db *gorm.DB) EventRepository { return &gormEventRepo{db: db} }

func (r *gormEventRepo) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	var e models.Event
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *gormEventRepo) List(ctx context.Context) ([]models.Event, error) {
	var events []models.Event
	err := r.db.WithContext(ctx).Order("sale_starts_at asc").Find(&events).Error
	return events, err
}

func (r *gormEventRepo) Create(ctx context.Context, e *models.Event) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// --- Seat ---

type gormSeatRepo struct{ db *gorm.DB }

// NewSeatRepository returns a GORM-backed SeatRepository.
func NewSeatRepository(db *gorm.DB) SeatRepository { return &gormSeatRepo{db: db} }

func (r *gormSeatRepo) GetByID(ctx context.Context, id int64) (*models.Seat, error) {
	var s models.Seat
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *gormSeatRepo) ListByEvent(ctx context.Context, eventID int64) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("section asc, seat_number asc").
		Find(&seats).Error
	return seats, err
}

func (r *gormSeatRepo) ReserveOptimistic(ctx context.Context, seatID, expectedVersion int64) (bool, error) {
	res := r.db.WithContext(ctx).Model(&models.Seat{}).
		Where("id = ? AND version = ? AND status = ?", seatID, expectedVersion, models.SeatAvailable).
		Updates(map[string]any{
			"status":  models.SeatReserved,
			"version": gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *gormSeatRepo) ReleaseExpired(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&models.Seat{}).
		Where("status = ? AND reserved_until IS NOT NULL AND reserved_until < ?",
			models.SeatReserved, time.Now()).
		Updates(map[string]any{
			"status":         models.SeatAvailable,
			"reserved_until": nil,
			"version":        gorm.Expr("version + 1"),
		})
	return res.RowsAffected, res.Error
}

// --- Reservation ---

type gormReservationRepo struct{ db *gorm.DB }

// NewReservationRepository returns a GORM-backed ReservationRepository.
func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &gormReservationRepo{db: db}
}

func (r *gormReservationRepo) Create(ctx context.Context, res *models.Reservation) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *gormReservationRepo) GetActiveBySeat(ctx context.Context, seatID int64) (*models.Reservation, error) {
	var res models.Reservation
	err := r.db.WithContext(ctx).
		Where("seat_id = ? AND status = ? AND expires_at > ?", seatID, models.ReservationHeld, time.Now()).
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

func (r *gormReservationRepo) MarkExpired(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&models.Reservation{}).
		Where("status = ? AND expires_at < ?", models.ReservationHeld, time.Now()).
		Update("status", models.ReservationExpired)
	return res.RowsAffected, res.Error
}

// --- Booking ---

type gormBookingRepo struct{ db *gorm.DB }

// NewBookingRepository returns a GORM-backed BookingRepository.
func NewBookingRepository(db *gorm.DB) BookingRepository { return &gormBookingRepo{db: db} }

func (r *gormBookingRepo) Create(ctx context.Context, b *models.Booking, seats []models.BookingSeat) error {
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

func (r *gormBookingRepo) GetByReference(ctx context.Context, ref string) (*models.Booking, error) {
	var b models.Booking
	err := r.db.WithContext(ctx).Preload("Seats").Where("reference = ?", ref).First(&b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}
