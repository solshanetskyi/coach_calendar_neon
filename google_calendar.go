package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"coach-calendar-app/handlers"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// GoogleCalendarService handles integration with Google Calendar
type GoogleCalendarService struct {
	service    *calendar.Service
	calendarID string
}

// NewGoogleCalendarService creates a new Google Calendar service using a service account
func NewGoogleCalendarService() (*GoogleCalendarService, error) {
	// Check if Google Calendar integration is enabled
	enabled := os.Getenv("GOOGLE_CALENDAR_ENABLED")
	if enabled != "true" && enabled != "yes" {
		log.Println("Google Calendar integration is disabled")
		return nil, nil
	}

	// Get the calendar ID (usually the email of the calendar owner)
	calendarID := os.Getenv("GOOGLE_CALENDAR_ID")
	if calendarID == "" {
		return nil, fmt.Errorf("GOOGLE_CALENDAR_ID environment variable is required")
	}

	// Get service account credentials
	// Can be either a file path or the JSON content directly
	credentialsPath := os.Getenv("GOOGLE_SERVICE_ACCOUNT_FILE")
	credentialsJSON := os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON")

	var opts []option.ClientOption

	if credentialsJSON != "" {
		// Use credentials JSON directly (useful for cloud deployments)
		opts = append(opts, option.WithCredentialsJSON([]byte(credentialsJSON)))
		log.Println("Using Google service account credentials from environment variable")
	} else if credentialsPath != "" {
		// Use credentials file
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
		log.Printf("Using Google service account credentials from file: %s", credentialsPath)
	} else {
		return nil, fmt.Errorf("either GOOGLE_SERVICE_ACCOUNT_FILE or GOOGLE_SERVICE_ACCOUNT_JSON is required")
	}

	// Add required scope
	opts = append(opts, option.WithScopes(calendar.CalendarReadonlyScope))

	ctx := context.Background()
	service, err := calendar.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Calendar service: %w", err)
	}

	log.Printf("Google Calendar service initialized for calendar: %s", calendarID)

	return &GoogleCalendarService{
		service:    service,
		calendarID: calendarID,
	}, nil
}

// GetBusySlots returns all busy time slots from Google Calendar for the given time range
func (g *GoogleCalendarService) GetBusySlots(startTime, endTime time.Time) ([]handlers.BusySlot, error) {
	if g == nil || g.service == nil {
		return nil, nil
	}

	ctx := context.Background()

	// Create a FreeBusy request
	freeBusyRequest := &calendar.FreeBusyRequest{
		TimeMin: startTime.Format(time.RFC3339),
		TimeMax: endTime.Format(time.RFC3339),
		Items: []*calendar.FreeBusyRequestItem{
			{Id: g.calendarID},
		},
	}

	freeBusyResponse, err := g.service.Freebusy.Query(freeBusyRequest).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query Google Calendar: %w", err)
	}

	var busySlots []handlers.BusySlot

	// Extract busy periods from the response
	calendarData, ok := freeBusyResponse.Calendars[g.calendarID]
	if !ok {
		log.Printf("No calendar data found for ID: %s", g.calendarID)
		return busySlots, nil
	}

	for _, busy := range calendarData.Busy {
		start, err := time.Parse(time.RFC3339, busy.Start)
		if err != nil {
			log.Printf("Error parsing busy start time: %v", err)
			continue
		}

		end, err := time.Parse(time.RFC3339, busy.End)
		if err != nil {
			log.Printf("Error parsing busy end time: %v", err)
			continue
		}

		busySlots = append(busySlots, handlers.BusySlot{
			Start: start,
			End:   end,
		})
	}

	log.Printf("Found %d busy slots from Google Calendar between %s and %s",
		len(busySlots), startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))

	return busySlots, nil
}

// IsSlotBusy checks if a specific time slot overlaps with any busy period
func (g *GoogleCalendarService) IsSlotBusy(slotTime time.Time, durationMinutes int, busySlots []handlers.BusySlot) bool {
	slotEnd := slotTime.Add(time.Duration(durationMinutes) * time.Minute)

	for _, busy := range busySlots {
		// Check if the slot overlaps with the busy period
		// Overlap occurs if: slot starts before busy ends AND slot ends after busy starts
		if slotTime.Before(busy.End) && slotEnd.After(busy.Start) {
			return true
		}
	}

	return false
}

// ValidateCredentials checks if the service account has access to the calendar
func (g *GoogleCalendarService) ValidateCredentials() error {
	if g == nil || g.service == nil {
		return fmt.Errorf("Google Calendar service not initialized")
	}

	ctx := context.Background()

	// Try to get calendar metadata to validate access
	_, err := g.service.CalendarList.Get(g.calendarID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to access calendar %s: %w (make sure you've shared the calendar with the service account email)", g.calendarID, err)
	}

	log.Printf("Successfully validated access to Google Calendar: %s", g.calendarID)
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
