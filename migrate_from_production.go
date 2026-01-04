package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// AdminSlotResponse represents the response from the admin API
type AdminSlotResponse struct {
	SlotTime string `json:"slot_time"`
	Status   string `json:"status"` // "available", "booked", "blocked"
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

// MigrationStats tracks migration statistics
type MigrationStats struct {
	BookingsFetched   int
	BookingsInserted  int
	BookingsSkipped   int
	BlockedFetched    int
	BlockedInserted   int
	BlockedSkipped    int
	Errors            []string
}

func main() {
	// Configuration
	productionURL := "https://sweip8cyfh.eu-central-1.awsapprunner.com"

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Please set it or use: export DATABASE_URL='...'")
	}

	log.Println("========================================")
	log.Println("Production to Neon Migration Script")
	log.Println("========================================")
	log.Printf("Production URL: %s", productionURL)
	log.Printf("Target Database: %s", maskPassword(dbURL))
	log.Println("========================================")

	// Confirm with user
	fmt.Print("\nThis will migrate all bookings and blocked slots from production to Neon.\n")
	fmt.Print("Existing records with the same slot_time will be skipped.\n")
	fmt.Print("Do you want to continue? (yes/no): ")

	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" && confirm != "y" {
		log.Println("Migration cancelled by user")
		return
	}

	// Connect to Neon database
	log.Println("\nConnecting to Neon database...")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Connected to Neon database")

	// Fetch data from production
	log.Printf("\nFetching data from production API: %s/api/admin/slots", productionURL)
	slots, err := fetchAdminSlots(productionURL + "/api/admin/slots")
	if err != nil {
		log.Fatalf("Failed to fetch slots: %v", err)
	}
	log.Printf("✓ Fetched %d slots from production", len(slots))

	// Process and migrate data
	stats := MigrationStats{}

	log.Println("\nMigrating data...")
	for _, slot := range slots {
		if slot.Status == "booked" {
			stats.BookingsFetched++
			if err := insertBooking(db, slot); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Booking error (%s): %v", slot.SlotTime, err))
				stats.BookingsSkipped++
			} else {
				stats.BookingsInserted++
			}
		} else if slot.Status == "blocked" {
			stats.BlockedFetched++
			if err := insertBlockedSlot(db, slot); err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("Blocked slot error (%s): %v", slot.SlotTime, err))
				stats.BlockedSkipped++
			} else {
				stats.BlockedInserted++
			}
		}
	}

	// Print summary
	log.Println("\n========================================")
	log.Println("Migration Summary")
	log.Println("========================================")
	log.Printf("Bookings fetched:    %d", stats.BookingsFetched)
	log.Printf("Bookings inserted:   %d", stats.BookingsInserted)
	log.Printf("Bookings skipped:    %d", stats.BookingsSkipped)
	log.Println("----------------------------------------")
	log.Printf("Blocked slots fetched:  %d", stats.BlockedFetched)
	log.Printf("Blocked slots inserted: %d", stats.BlockedInserted)
	log.Printf("Blocked slots skipped:  %d", stats.BlockedSkipped)
	log.Println("========================================")

	if len(stats.Errors) > 0 {
		log.Println("\nErrors encountered:")
		for _, err := range stats.Errors {
			log.Printf("  - %s", err)
		}
	}

	log.Println("\n✓ Migration completed successfully!")
}

// fetchAdminSlots fetches all slots from the admin API
func fetchAdminSlots(url string) ([]AdminSlotResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var slots []AdminSlotResponse
	if err := json.NewDecoder(resp.Body).Decode(&slots); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return slots, nil
}

// insertBooking inserts a booking into the database
func insertBooking(db *sql.DB, slot AdminSlotResponse) error {
	// Parse the slot time
	slotTime, err := time.Parse(time.RFC3339, slot.SlotTime)
	if err != nil {
		return fmt.Errorf("invalid time format: %w", err)
	}

	// Convert to UTC for storage
	slotTimeUTC := slotTime.UTC()

	// Insert booking (ignore if already exists due to unique constraint)
	query := `
		INSERT INTO bookings (slot_time, name, email, created_at, duration)
		VALUES ($1, $2, $3, $4, 30)
		ON CONFLICT (slot_time) DO NOTHING
	`

	result, err := db.Exec(query, slotTimeUTC, slot.Name, slot.Email, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("  ⊙ Booking skipped (already exists): %s - %s", slot.SlotTime, slot.Name)
		return fmt.Errorf("already exists")
	}

	log.Printf("  ✓ Booking inserted: %s - %s (%s)", slot.SlotTime, slot.Name, slot.Email)
	return nil
}

// insertBlockedSlot inserts a blocked slot into the database
func insertBlockedSlot(db *sql.DB, slot AdminSlotResponse) error {
	// Parse the slot time
	slotTime, err := time.Parse(time.RFC3339, slot.SlotTime)
	if err != nil {
		return fmt.Errorf("invalid time format: %w", err)
	}

	// Convert to UTC for storage
	slotTimeUTC := slotTime.UTC()

	// Insert blocked slot (ignore if already exists due to unique constraint)
	query := `
		INSERT INTO blocked_slots (slot_time, created_at)
		VALUES ($1, $2)
		ON CONFLICT (slot_time) DO NOTHING
	`

	result, err := db.Exec(query, slotTimeUTC, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("  ⊙ Blocked slot skipped (already exists): %s", slot.SlotTime)
		return fmt.Errorf("already exists")
	}

	log.Printf("  ✓ Blocked slot inserted: %s", slot.SlotTime)
	return nil
}

// maskPassword masks the password in a database URL for logging
func maskPassword(dbURL string) string {
	if len(dbURL) == 0 {
		return ""
	}

	start := -1
	end := -1
	colonCount := 0

	for i, ch := range dbURL {
		if ch == ':' {
			colonCount++
			if colonCount == 2 {
				start = i + 1
			}
		}
		if ch == '@' && start != -1 {
			end = i
			break
		}
	}

	if start != -1 && end != -1 && start < end {
		return dbURL[:start] + "****" + dbURL[end:]
	}

	return dbURL
}
