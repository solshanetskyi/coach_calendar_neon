package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type AvailableSlot struct {
	SlotTime  string
	Available bool
}

func initDB() error {
	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to Neon PostgreSQL database")

	// Create tables with PostgreSQL syntax
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS bookings (
		id SERIAL PRIMARY KEY,
		slot_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		duration INTEGER NOT NULL DEFAULT 30
	);
	CREATE INDEX IF NOT EXISTS idx_slot_time ON bookings(slot_time);

	CREATE TABLE IF NOT EXISTS blocked_slots (
		id SERIAL PRIMARY KEY,
		slot_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_blocked_slot_time ON blocked_slots(slot_time);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database tables initialized successfully")
	return nil
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
