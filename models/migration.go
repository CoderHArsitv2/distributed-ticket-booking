package models

import "gorm.io/gorm"

// AutoMigrate creates and updates every table to match the model definitions in
// this package. It is the project's only schema-management path: a schema
// change means changing the structs here, not writing a versioned SQL file.
//
// Models are listed parents-first so GORM can build foreign keys in order.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Event{},
		&Seat{},
		&Reservation{},
		&Booking{},
		&BookingSeat{},
	)
}
