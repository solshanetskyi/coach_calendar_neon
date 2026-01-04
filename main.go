package main

import (
	"log"
	"net/http"
	"os"

	"coach-calendar-app/handlers"
)

// Wrapper function to convert types
func generateSlotsForHandlers() []handlers.AvailableSlot {
	slots := generateAvailableSlots()
	handlerSlots := make([]handlers.AvailableSlot, len(slots))
	for i, slot := range slots {
		handlerSlots[i] = handlers.AvailableSlot{
			SlotTime:  slot.SlotTime,
			Available: slot.Available,
		}
	}
	return handlerSlots
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize email service
	emailService := NewEmailService()

	// Initialize Zoom service
	zoomService := NewZoomService()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize API handlers
	apiHandlers := handlers.NewAPIHandlers(db, generateSlotsForHandlers, emailService, zoomService)

	// Register page routes
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/admin", handlers.AdminHandler)

	// Register API routes
	http.HandleFunc("/api/slots", apiHandlers.GetSlots)
	http.HandleFunc("/api/bookings", apiHandlers.CreateBooking)
	http.HandleFunc("/api/admin/slots", apiHandlers.GetAdminSlots)
	http.HandleFunc("/api/admin/block", apiHandlers.BlockSlot)
	http.HandleFunc("/api/admin/unblock", apiHandlers.UnblockSlot)
	http.HandleFunc("/api/admin/debug-blocked", apiHandlers.DebugBlockedSlots)
	http.HandleFunc("/api/admin/clear-all-blocked", apiHandlers.ClearAllBlockedSlots)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
