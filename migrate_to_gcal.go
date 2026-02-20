//go:build ignore
// +build ignore

// Migration script to copy existing blocked slots and bookings from database to Google Calendar
// Run with: go run migrate_to_gcal.go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Booking struct {
	SlotTime time.Time
	Name     string
	Email    string
	Phone    sql.NullString
	ZoomLink sql.NullString
}

type BlockedSlot struct {
	SlotTime time.Time
}

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would be created without actually creating events")
	bookingsOnly := flag.Bool("bookings-only", false, "Only migrate bookings, skip blocked slots")
	blockedOnly := flag.Bool("blocked-only", false, "Only migrate blocked slots, skip bookings")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Migration Script - Copy DB slots to Google Calendar")
		fmt.Println("\nCopies existing blocked slots and bookings from database to Google Calendar")
		fmt.Println("\nUsage:")
		fmt.Println("  go run migrate_to_gcal.go [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  -dry-run")
		fmt.Println("        Show what would be created without actually creating events")
		fmt.Println("  -bookings-only")
		fmt.Println("        Only migrate bookings, skip blocked slots")
		fmt.Println("  -blocked-only")
		fmt.Println("        Only migrate blocked slots, skip bookings")
		fmt.Println("  -help")
		fmt.Println("        Show this help message")
		fmt.Println("\nRequired environment variables:")
		fmt.Println("  DATABASE_URL                   - PostgreSQL connection string")
		fmt.Println("  GOOGLE_CALENDAR_ID             - Google Calendar ID")
		fmt.Println("  GOOGLE_SERVICE_ACCOUNT_FILE    - Path to service account JSON file")
		fmt.Println("  or GOOGLE_SERVICE_ACCOUNT_JSON - Service account JSON content")
		return
	}

	// Get database URL
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Get Google Calendar config
	calendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if calendarID == "" {
		log.Fatal("GOOGLE_CALENDAR_ID environment variable is required")
	}

	credentialsJSON := os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON")
	credentialsFile := os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE")
	if credentialsJSON == "" && credentialsFile == "" {
		log.Fatal("Either GOOGLE_SERVICE_ACCOUNT_JSON or GOOGLE_SERVICE_ACCOUNT_FILE is required")
	}

	// Connect to database
	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Initialize Google Calendar service
	log.Println("Initializing Google Calendar service...")
	ctx := context.Background()
	var opts []option.ClientOption

	if credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
	} else {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	opts = append(opts, option.WithScopes(calendar.CalendarScope))

	calService, err := calendar.NewService(ctx, opts...)
	if err != nil {
		log.Fatalf("Failed to create Google Calendar service: %v", err)
	}
	log.Printf("Google Calendar service initialized for calendar: %s", calendarID)

	// Validate calendar access
	_, err = calService.Calendars.Get(calendarID).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Failed to access calendar: %v", err)
	}
	log.Println("Calendar access validated")

	log.Println("========================================")
	if *dryRun {
		log.Println("DRY RUN MODE - No events will be created")
	}
	log.Println("========================================")

	// Migrate blocked slots
	if !*bookingsOnly {
		migrateBlockedSlots(ctx, db, calService, calendarID, *dryRun)
	}

	// Migrate bookings
	if !*blockedOnly {
		migrateBookings(ctx, db, calService, calendarID, *dryRun)
	}

	log.Println("========================================")
	log.Println("Migration complete!")
}

func migrateBlockedSlots(ctx context.Context, db *sql.DB, calService *calendar.Service, calendarID string, dryRun bool) {
	log.Println("\n--- Migrating Blocked Slots ---")

	// Query blocked slots (only future ones)
	rows, err := db.Query("SELECT slot_time FROM blocked_slots WHERE slot_time > NOW() ORDER BY slot_time")
	if err != nil {
		log.Printf("Error querying blocked slots: %v", err)
		return
	}
	defer rows.Close()

	var blockedSlots []BlockedSlot
	for rows.Next() {
		var slot BlockedSlot
		if err := rows.Scan(&slot.SlotTime); err != nil {
			log.Printf("Error scanning blocked slot: %v", err)
			continue
		}
		blockedSlots = append(blockedSlots, slot)
	}

	log.Printf("Found %d future blocked slots", len(blockedSlots))

	if len(blockedSlots) == 0 {
		log.Println("No blocked slots to migrate")
		return
	}

	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, slot := range blockedSlots {
		// Check if event already exists for this time
		exists, err := eventExistsAtTime(ctx, calService, calendarID, slot.SlotTime, "Заблоковано")
		if err != nil {
			log.Printf("Error checking existing event: %v", err)
		}
		if exists {
			skipCount++
			continue
		}

		if dryRun {
			log.Printf("Would create blocked event at: %s", slot.SlotTime.Format("2006-01-02 15:04 MST"))
			successCount++
			continue
		}

		err = createBlockedEvent(ctx, calService, calendarID, slot.SlotTime)
		if err != nil {
			log.Printf("Error creating blocked event for %s: %v", slot.SlotTime.Format("2006-01-02 15:04"), err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("Blocked slots: %d created, %d skipped (already exist), %d errors", successCount, skipCount, errorCount)
}

func migrateBookings(ctx context.Context, db *sql.DB, calService *calendar.Service, calendarID string, dryRun bool) {
	log.Println("\n--- Migrating Bookings ---")

	// Query bookings (only future ones)
	rows, err := db.Query(`
		SELECT slot_time, name, email, phone, zoom_link
		FROM bookings
		WHERE slot_time > NOW()
		ORDER BY slot_time
	`)
	if err != nil {
		log.Printf("Error querying bookings: %v", err)
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.SlotTime, &b.Name, &b.Email, &b.Phone, &b.ZoomLink); err != nil {
			log.Printf("Error scanning booking: %v", err)
			continue
		}
		bookings = append(bookings, b)
	}

	log.Printf("Found %d future bookings", len(bookings))

	if len(bookings) == 0 {
		log.Println("No bookings to migrate")
		return
	}

	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, booking := range bookings {
		// Check if event already exists for this time with this client
		searchTerm := fmt.Sprintf("Консультація: %s", booking.Name)
		exists, err := eventExistsAtTime(ctx, calService, calendarID, booking.SlotTime, searchTerm)
		if err != nil {
			log.Printf("Error checking existing event: %v", err)
		}
		if exists {
			skipCount++
			continue
		}

		if dryRun {
			log.Printf("Would create booking event: %s at %s", booking.Name, booking.SlotTime.Format("2006-01-02 15:04 MST"))
			successCount++
			continue
		}

		phone := ""
		if booking.Phone.Valid {
			phone = booking.Phone.String
		}
		zoomLink := ""
		if booking.ZoomLink.Valid {
			zoomLink = booking.ZoomLink.String
		}

		err = createBookingEvent(ctx, calService, calendarID, booking.Name, booking.Email, phone, booking.SlotTime, zoomLink)
		if err != nil {
			log.Printf("Error creating booking event for %s at %s: %v", booking.Name, booking.SlotTime.Format("2006-01-02 15:04"), err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("Bookings: %d created, %d skipped (already exist), %d errors", successCount, skipCount, errorCount)
}

func eventExistsAtTime(ctx context.Context, calService *calendar.Service, calendarID string, slotTime time.Time, summaryContains string) (bool, error) {
	// Search for events at this exact time
	timeMin := slotTime.Add(-1 * time.Minute).Format(time.RFC3339)
	timeMax := slotTime.Add(1 * time.Minute).Format(time.RFC3339)

	events, err := calService.Events.List(calendarID).
		TimeMin(timeMin).
		TimeMax(timeMax).
		SingleEvents(true).
		Context(ctx).
		Do()
	if err != nil {
		return false, err
	}

	for _, event := range events.Items {
		if event.Summary == summaryContains || (len(summaryContains) > 0 && len(event.Summary) > 0 && event.Summary == summaryContains) {
			return true, nil
		}
	}

	return false, nil
}

func createBlockedEvent(ctx context.Context, calService *calendar.Service, calendarID string, slotTime time.Time) error {
	// Create a 30-minute blocked event
	endTime := slotTime.Add(30 * time.Minute)

	event := &calendar.Event{
		Summary:     "Заблоковано",
		Description: "Час заблоковано (мігровано з бази даних)",
		Start: &calendar.EventDateTime{
			DateTime: slotTime.Format(time.RFC3339),
			TimeZone: "Europe/Amsterdam",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "Europe/Amsterdam",
		},
		ColorId: "4", // Red/pink color for blocked
	}

	createdEvent, err := calService.Events.Insert(calendarID, event).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create blocked event: %w", err)
	}

	log.Printf("Created blocked event: %s (ID: %s)", slotTime.Format("2006-01-02 15:04"), createdEvent.Id)
	return nil
}

func createBookingEvent(ctx context.Context, calService *calendar.Service, calendarID string, clientName, clientEmail, clientPhone string, slotTime time.Time, zoomLink string) error {
	// Default to 30 minutes duration
	endTime := slotTime.Add(30 * time.Minute)

	// Build description
	description := fmt.Sprintf("Клієнт: %s\nEmail: %s", clientName, clientEmail)
	if clientPhone != "" {
		description += fmt.Sprintf("\nТелефон: %s", clientPhone)
	}
	if zoomLink != "" {
		description += fmt.Sprintf("\n\nZoom: %s", zoomLink)
	}
	description += "\n\n(Мігровано з бази даних)"

	event := &calendar.Event{
		Summary:     fmt.Sprintf("Консультація: %s", clientName),
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: slotTime.Format(time.RFC3339),
			TimeZone: "Europe/Amsterdam",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "Europe/Amsterdam",
		},
	}

	if zoomLink != "" {
		event.Location = zoomLink
	}

	createdEvent, err := calService.Events.Insert(calendarID, event).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to create booking event: %w", err)
	}

	log.Printf("Created booking event: %s at %s (ID: %s)", clientName, slotTime.Format("2006-01-02 15:04"), createdEvent.Id)
	return nil
}

// Helper function to pretty print JSON for debugging
func prettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}
