// go:build ignore
// +build ignore

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Booking struct {
	ID        int       `json:"id"`
	SlotTime  time.Time `json:"slot_time"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	Duration  int       `json:"duration"`
}

type BlockedSlot struct {
	ID        int       `json:"id"`
	SlotTime  time.Time `json:"slot_time"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Please set it with your Neon connection string.")
	}

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}

	fmt.Println("✅ Connected to Neon PostgreSQL database")

	// Read and import bookings
	if _, err := os.Stat("bookings_export.json"); err == nil {
		bookingsData, err := os.ReadFile("bookings_export.json")
		if err != nil {
			log.Fatalf("Failed to read bookings_export.json: %v", err)
		}

		var bookings []Booking
		if err := json.Unmarshal(bookingsData, &bookings); err != nil {
			log.Fatalf("Failed to parse bookings JSON: %v", err)
		}

		fmt.Printf("Importing %d bookings...\n", len(bookings))
		importedBookings := 0
		skippedBookings := 0

		for _, b := range bookings {
			_, err := db.Exec(
				"INSERT INTO bookings (slot_time, name, email, created_at, duration) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (slot_time) DO NOTHING",
				b.SlotTime, b.Name, b.Email, b.CreatedAt, b.Duration,
			)
			if err != nil {
				log.Printf("Warning: Failed to import booking for %s: %v", b.SlotTime, err)
				skippedBookings++
			} else {
				importedBookings++
			}
		}

		fmt.Printf("   ✅ Imported %d bookings (skipped %d duplicates)\n", importedBookings, skippedBookings)
	} else {
		fmt.Println("⚠️  No bookings_export.json found, skipping bookings import")
	}

	// Read and import blocked slots
	if _, err := os.Stat("blocked_slots_export.json"); err == nil {
		blockedData, err := os.ReadFile("blocked_slots_export.json")
		if err != nil {
			log.Fatalf("Failed to read blocked_slots_export.json: %v", err)
		}

		var blockedSlots []BlockedSlot
		if err := json.Unmarshal(blockedData, &blockedSlots); err != nil {
			log.Fatalf("Failed to parse blocked slots JSON: %v", err)
		}

		fmt.Printf("Importing %d blocked slots...\n", len(blockedSlots))
		importedBlocked := 0
		skippedBlocked := 0

		for _, bs := range blockedSlots {
			_, err := db.Exec(
				"INSERT INTO blocked_slots (slot_time, created_at) VALUES ($1, $2) ON CONFLICT (slot_time) DO NOTHING",
				bs.SlotTime, bs.CreatedAt,
			)
			if err != nil {
				log.Printf("Warning: Failed to import blocked slot for %s: %v", bs.SlotTime, err)
				skippedBlocked++
			} else {
				importedBlocked++
			}
		}

		fmt.Printf("   ✅ Imported %d blocked slots (skipped %d duplicates)\n", importedBlocked, skippedBlocked)
	} else {
		fmt.Println("⚠️  No blocked_slots_export.json found, skipping blocked slots import")
	}

	fmt.Println("\n✅ Import completed successfully!")
	fmt.Println("You can now start your application with the Neon database.")
}
