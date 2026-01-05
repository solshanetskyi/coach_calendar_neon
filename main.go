package main

import (
	"log"
	"net/http"
	"os"

	"coach-calendar-app/handlers"

	"github.com/joho/godotenv"
)

// printConfigSummary prints all configuration environment variables at startup
func printConfigSummary() {
	log.Println("========================================")
	log.Println("Configuration Summary")
	log.Println("========================================")

	// Database configuration
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		// Mask password in URL for security
		maskedURL := maskPassword(dbURL)
		log.Printf("DATABASE_URL: %s", maskedURL)
	} else {
		log.Println("DATABASE_URL: <not set>")
	}

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080 (default)"
	}
	log.Printf("PORT: %s", port)

	// Email configuration
	log.Printf("SMTP_HOST: %s", getEnvOrDefault("SMTP_HOST", "<not set>"))
	log.Printf("SMTP_PORT: %s", getEnvOrDefault("SMTP_PORT", "<not set>"))
	log.Printf("SMTP_FROM: %s", getEnvOrDefault("SMTP_FROM", "<not set>"))
	log.Printf("SMTP_USER_FRIENDLY_FROM: %s", getEnvOrDefault("SMTP_USER_FRIENDLY_FROM", "<not set>"))
	log.Printf("SMTP_PASSWORD: %s", maskSecret(os.Getenv("SMTP_PASSWORD")))
	log.Printf("USE_AWS_SES: %s", getEnvOrDefault("USE_AWS_SES", "<not set>"))
	log.Printf("AWS_REGION: %s", getEnvOrDefault("AWS_REGION", "<not set>"))

	// Feature toggles
	log.Printf("SEND_CONFIRMATION_EMAIL: %s", getEnvOrDefault("SEND_CONFIRMATION_EMAIL", "<not set>"))
	log.Printf("CREATE_ZOOM_MEETING: %s", getEnvOrDefault("CREATE_ZOOM_MEETING", "<not set>"))

	// Zoom configuration
	log.Printf("ZOOM_ACCOUNT_ID: %s", maskSecret(os.Getenv("ZOOM_ACCOUNT_ID")))
	log.Printf("ZOOM_CLIENT_ID: %s", maskSecret(os.Getenv("ZOOM_CLIENT_ID")))
	log.Printf("ZOOM_CLIENT_SECRET: %s", maskSecret(os.Getenv("ZOOM_CLIENT_SECRET")))

	log.Println("========================================")
}

// getEnvOrDefault returns the environment variable value or a default string
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// maskSecret masks sensitive values for logging
func maskSecret(secret string) string {
	if secret == "" {
		return "<not set>"
	}
	if len(secret) <= 4 {
		return "****"
	}
	return secret[:4] + "****"
}

// maskPassword masks the password in a database URL
func maskPassword(dbURL string) string {
	// Format: postgres://user:password@host/db
	// Replace password with ****
	if len(dbURL) == 0 {
		return ""
	}

	// Find the password section (between : and @)
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
	// Load environment variables from .env file (if it exists)
	// Ignore error if file doesn't exist (e.g., in production)
	_ = godotenv.Load()

	// Print configuration summary
	printConfigSummary()

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
	http.HandleFunc("/api/admin/cancel", apiHandlers.CancelBooking)
	http.HandleFunc("/api/admin/debug-blocked", apiHandlers.DebugBlockedSlots)
	http.HandleFunc("/api/admin/clear-all-blocked", apiHandlers.ClearAllBlockedSlots)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
