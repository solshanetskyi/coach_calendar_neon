package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type AvailableSlot struct {
	SlotTime  string
	Available bool
}

// ensureDatabaseExists connects to the default database and creates the target database if it doesn't exist
func ensureDatabaseExists(databaseURL string) error {
	// Parse the database URL to extract the database name
	// Format: postgres://user:password@host:port/dbname?params
	dbName := extractDatabaseName(databaseURL)
	if dbName == "" || dbName == "postgres" {
		// If no database name or already using postgres, skip creation
		return nil
	}

	// Create a connection URL to the default 'postgres' database
	defaultURL := strings.Replace(databaseURL, "/"+dbName, "/postgres", 1)

	// Connect to the default database
	defaultDB, err := sql.Open("postgres", defaultURL)
	if err != nil {
		return fmt.Errorf("failed to connect to default database: %w", err)
	}
	defer defaultDB.Close()

	if err = defaultDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping default database: %w", err)
	}

	// Check if the target database exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)"
	err = defaultDB.QueryRow(query, dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		// Create the database
		createQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
		_, err = defaultDB.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database '%s' created successfully", dbName)
	} else {
		log.Printf("Database '%s' already exists", dbName)
	}

	return nil
}

// extractDatabaseName extracts the database name from a PostgreSQL connection URL
func extractDatabaseName(databaseURL string) string {
	// Find the last '/' which separates the host:port from the database name
	lastSlash := strings.LastIndex(databaseURL, "/")
	if lastSlash == -1 {
		return ""
	}

	// Extract everything after the last '/'
	remainder := databaseURL[lastSlash+1:]

	// Remove any query parameters (anything after '?')
	if questionMark := strings.Index(remainder, "?"); questionMark != -1 {
		remainder = remainder[:questionMark]
	}

	return remainder
}

func initDB() error {
	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Verify we're using the 'bookings' database
	dbName := extractDatabaseName(databaseURL)
	if dbName != "bookings" {
		log.Printf("Warning: Expected 'bookings' database but found '%s'. Please create a 'bookings' database in Neon Console.", dbName)
	}

	// Note: For Neon databases, we skip ensureDatabaseExists() because:
	// 1. Neon databases are pre-created through the Neon console
	// 2. Neon's pooler doesn't support connecting to the default 'postgres' database
	//    with SCRAM-SHA-256 authentication used in connection pooling
	// 3. The database specified in DATABASE_URL is guaranteed to exist

	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to Neon PostgreSQL database: %s", dbName)

	// Create tables with PostgreSQL syntax
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS bookings (
		id SERIAL PRIMARY KEY,
		slot_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		duration INTEGER NOT NULL DEFAULT 30,
		zoom_link TEXT
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
