package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type AvailableSlot struct {
	SlotTime  string
	Available bool
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", "./bookings.db")
	if err != nil {
		return err
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS bookings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slot_time DATETIME NOT NULL UNIQUE,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		duration INTEGER NOT NULL DEFAULT 30
	);
	CREATE INDEX IF NOT EXISTS idx_slot_time ON bookings(slot_time);

	CREATE TABLE IF NOT EXISTS blocked_slots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slot_time DATETIME NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_blocked_slot_time ON blocked_slots(slot_time);
	`

	_, err = db.Exec(createTableSQL)
	return err
}

func generateAvailableSlots() []AvailableSlot {
	slots := []AvailableSlot{}

	// Load Amsterdam timezone for consistency with booking system
	location, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		// Fallback to UTC if Amsterdam timezone not available
		location = time.UTC
	}

	now := time.Now().In(location)

	// Only generate slots for January
	// Find the next January (could be current year or next year)
	currentYear := now.Year()
	currentMonth := now.Month()

	var januaryYear int
	if currentMonth == time.January {
		// We're in January, use current year
		januaryYear = currentYear
	} else {
		// We're past January, use next year
		januaryYear = currentYear + 1
	}

	// Generate slots for all days in January
	for day := 1; day <= 31; day++ {
		// Generate 30-minute slots from 9 AM to 8 PM (Amsterdam time)
		for hour := 9; hour <= 20; hour++ {
			for minute := 0; minute < 60; minute += 30 {
				// Skip the 30-minute slot at 8:30 PM to keep end time at 8 PM
				if hour == 20 && minute == 30 {
					continue
				}

				slotTime := time.Date(januaryYear, time.January, day, hour, minute, 0, 0, location)

				// Only include future slots
				if slotTime.After(now) {
					slots = append(slots, AvailableSlot{
						SlotTime:  slotTime.Format(time.RFC3339),
						Available: true,
					})
				}
			}
		}
	}

	return slots
}
