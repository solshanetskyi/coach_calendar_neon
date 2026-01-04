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

	_ "modernc.org/sqlite"
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
	// Check if bookings.db exists
	if _, err := os.Stat("./bookings.db"); os.IsNotExist(err) {
		log.Fatal("bookings.db not found. Nothing to export.")
	}

	// Open SQLite database
	db, err := sql.Open("sqlite", "./bookings.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Export bookings
	bookings := []Booking{}
	rows, err := db.Query("SELECT id, slot_time, name, email, created_at, duration FROM bookings")
	if err != nil {
		log.Fatalf("Failed to query bookings: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.SlotTime, &b.Name, &b.Email, &b.CreatedAt, &b.Duration); err != nil {
			log.Printf("Warning: Failed to scan booking row: %v", err)
			continue
		}
		bookings = append(bookings, b)
	}

	// Export blocked slots
	blockedSlots := []BlockedSlot{}
	blockedRows, err := db.Query("SELECT id, slot_time, created_at FROM blocked_slots")
	if err != nil {
		log.Fatalf("Failed to query blocked slots: %v", err)
	}
	defer blockedRows.Close()

	for blockedRows.Next() {
		var bs BlockedSlot
		if err := blockedRows.Scan(&bs.ID, &bs.SlotTime, &bs.CreatedAt); err != nil {
			log.Printf("Warning: Failed to scan blocked slot row: %v", err)
			continue
		}
		blockedSlots = append(blockedSlots, bs)
	}

	// Write bookings to JSON
	bookingsJSON, err := json.MarshalIndent(bookings, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal bookings: %v", err)
	}

	if err := os.WriteFile("bookings_export.json", bookingsJSON, 0644); err != nil {
		log.Fatalf("Failed to write bookings export: %v", err)
	}

	// Write blocked slots to JSON
	blockedJSON, err := json.MarshalIndent(blockedSlots, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal blocked slots: %v", err)
	}

	if err := os.WriteFile("blocked_slots_export.json", blockedJSON, 0644); err != nil {
		log.Fatalf("Failed to write blocked slots export: %v", err)
	}

	fmt.Printf("✅ Export completed successfully!\n")
	fmt.Printf("   - Exported %d bookings to bookings_export.json\n", len(bookings))
	fmt.Printf("   - Exported %d blocked slots to blocked_slots_export.json\n", len(blockedSlots))
	fmt.Printf("\nYou can now import this data to PostgreSQL using import_to_postgres.go\n")
}
