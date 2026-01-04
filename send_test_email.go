// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	// Command line flags
	toEmail := flag.String("to", "", "Recipient email address (required)")
	name := flag.String("name", "Test User", "Recipient name")
	zoomLink := flag.String("zoom", "", "Zoom meeting link (optional)")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Test Email Sender for Coach Calendar")
		fmt.Println("\nUsage:")
		fmt.Println("  go run send_test_email.go email.go -to recipient@example.com [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  -to string")
		fmt.Println("        Recipient email address (required)")
		fmt.Println("  -name string")
		fmt.Println("        Recipient name (default \"Test User\")")
		fmt.Println("  -zoom string")
		fmt.Println("        Zoom meeting link (optional)")
		fmt.Println("  -help")
		fmt.Println("        Show this help message")
		fmt.Println("\nEnvironment Variables Required:")
		fmt.Println("  SMTP_HOST       - SMTP server hostname")
		fmt.Println("  SMTP_PORT       - SMTP server port")
		fmt.Println("  SMTP_FROM       - From email address")
		fmt.Println("  SMTP_PASSWORD   - SMTP password")
		fmt.Println("\nExample:")
		fmt.Println("  go run send_test_email.go email.go -to test@example.com -name \"John Doe\" -zoom \"https://zoom.us/j/123456789\"")
		return
	}

	if *toEmail == "" {
		log.Fatal("Error: -to flag is required. Use -help for usage information.")
	}

	// Check for required environment variables
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")

	// User-friendly "From" name with default
	fromName := os.Getenv("SMTP_USER_FRIENDLY_FROM")
	if fromName == "" {
		fromName = "Христина Івасюк"
	}

	if host == "" || port == "" || from == "" || password == "" {
		log.Fatal("Error: Email service is not enabled. Please set the required environment variables:\n" +
			"  SMTP_HOST, SMTP_PORT, SMTP_FROM, SMTP_PASSWORD")
	}

	// Initialize email service
	emailService := &EmailService{
		SMTPHost: host,
		SMTPPort: port,
		From:     from,
		FromName: fromName,
		Password: password,
		Enabled:  true,
	}

	// Create a test booking time (tomorrow at 2 PM)
	now := time.Now()
	testSlotTime := time.Date(now.Year(), now.Month(), now.Day()+1, 14, 0, 0, 0, now.Location())

	log.Println("========================================")
	log.Println("Sending test confirmation email...")
	log.Println("========================================")
	log.Printf("From: %s <%s>", fromName, from)
	log.Printf("To: %s", *toEmail)
	log.Printf("Name: %s", *name)
	log.Printf("Slot Time: %s", testSlotTime.Format("Monday, January 2, 2006 at 3:04 PM MST"))
	if *zoomLink != "" {
		log.Printf("Zoom Link: %s", *zoomLink)
	}
	log.Println("========================================")

	// Send the test email
	err := emailService.SendBookingConfirmation(*name, *toEmail, testSlotTime, *zoomLink)
	if err != nil {
		log.Fatalf("Failed to send test email: %v", err)
	}

	log.Println("========================================")
	log.Println("✅ Test email sent successfully!")
	log.Println("========================================")
	log.Printf("Check your inbox at: %s", *toEmail)
	log.Println("Note: The email may take a few moments to arrive and might be in your spam folder.")
}
