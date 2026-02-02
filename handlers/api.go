package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// EmailSender interface for sending emails
type EmailSender interface {
	SendBookingConfirmation(name, email string, slotTime time.Time, zoomLink string, duration int) error
	SendBookingNotificationToOwner(clientName, clientEmail, clientPhone string, slotTime time.Time, zoomLink string) error
}

// ZoomMeetingCreator interface for creating and deleting Zoom meetings
type ZoomMeetingCreator interface {
	CreateMeeting(name, email string, slotTime time.Time) (string, error)
	DeleteMeeting(joinURL string) error
}

// BusySlot represents a busy time period from Google Calendar
type BusySlot struct {
	Start time.Time
	End   time.Time
}

// GoogleCalendarChecker interface for checking Google Calendar availability and creating events
type GoogleCalendarChecker interface {
	GetBusySlots(startTime, endTime time.Time) ([]BusySlot, error)
	IsSlotBusy(slotTime time.Time, durationMinutes int, busySlots []BusySlot) bool
	CreateEvent(clientName, clientEmail, clientPhone string, slotTime time.Time, zoomLink string, duration int) error
}

// Re-export types from main package
type Booking struct {
	ID        int       `json:"id"`
	SlotTime  time.Time `json:"slot_time"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ZoomLink  string    `json:"zoom_link,omitempty"`
}

type BookingRequest struct {
	SlotTime string `json:"slot_time"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	Duration int    `json:"duration,omitempty"` // 30 or 60 minutes
}

type AvailableSlot struct {
	SlotTime  string `json:"slot_time"`
	Available bool   `json:"available"`
}

type AdminSlot struct {
	SlotTime string `json:"slot_time"`
	Status   string `json:"status"` // "available", "booked", "blocked"
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type GenerateSlotsFn func() []AvailableSlot

type Client struct {
	ID          int       `json:"id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	TelegramID  string    `json:"telegram_id,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ClientRequest struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	TelegramID  string `json:"telegram_id,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type APIHandlers struct {
	DB                     *sql.DB
	GenerateAvailableSlots GenerateSlotsFn
	EmailService           EmailSender
	ZoomService            ZoomMeetingCreator
	GoogleCalendar         GoogleCalendarChecker
}

func NewAPIHandlers(db *sql.DB, generateSlotsFn GenerateSlotsFn, emailService EmailSender, zoomService ZoomMeetingCreator, googleCalendar GoogleCalendarChecker) *APIHandlers {
	return &APIHandlers{
		DB:                     db,
		GenerateAvailableSlots: generateSlotsFn,
		EmailService:           emailService,
		ZoomService:            zoomService,
		GoogleCalendar:         googleCalendar,
	}
}

func (h *APIHandlers) GetSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get duration from query param (default 30 minutes)
	durationStr := r.URL.Query().Get("duration")
	duration := 30
	if durationStr == "60" {
		duration = 60
	}

	// Generate base 30-minute slots
	allSlots := h.GenerateAvailableSlots()

	// Filter slots based on duration
	var slots []AvailableSlot
	if duration == 60 {
		// For 60-minute slots, only keep slots on the hour (not :30)
		for _, slot := range allSlots {
			slotTime, err := time.Parse(time.RFC3339, slot.SlotTime)
			if err != nil {
				continue
			}
			// Only include slots that start on the hour
			if slotTime.Minute() == 0 {
				slots = append(slots, slot)
			}
		}
	} else {
		slots = allSlots
	}

	// Get booked slots from database
	rows, err := h.DB.Query("SELECT slot_time FROM bookings")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying bookings: %v", err)
		return
	}
	defer rows.Close()

	bookedSlots := make(map[int64]bool)
	for rows.Next() {
		var slotTime time.Time
		if err := rows.Scan(&slotTime); err != nil {
			continue
		}
		// Use Unix timestamp for timezone-independent comparison
		bookedSlots[slotTime.Unix()] = true
	}

	// Get blocked slots from database
	blockedRows, err := h.DB.Query("SELECT slot_time FROM blocked_slots")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying blocked slots: %v", err)
		return
	}
	defer blockedRows.Close()

	blockedSlots := make(map[int64]bool)
	for blockedRows.Next() {
		var slotTime time.Time
		if err := blockedRows.Scan(&slotTime); err != nil {
			continue
		}
		// Use Unix timestamp for timezone-independent comparison
		blockedSlots[slotTime.Unix()] = true
	}

	// Get busy slots from Google Calendar (if enabled)
	var googleBusySlots []BusySlot
	if h.GoogleCalendar != nil && len(slots) > 0 {
		// Calculate the time range for Google Calendar query
		var minTime, maxTime time.Time
		for _, slot := range slots {
			slotTime, err := time.Parse(time.RFC3339, slot.SlotTime)
			if err != nil {
				continue
			}
			if minTime.IsZero() || slotTime.Before(minTime) {
				minTime = slotTime
			}
			if maxTime.IsZero() || slotTime.After(maxTime) {
				maxTime = slotTime
			}
		}

		if !minTime.IsZero() && !maxTime.IsZero() {
			// Add some buffer to the end time
			maxTime = maxTime.Add(time.Hour)
			busySlots, err := h.GoogleCalendar.GetBusySlots(minTime, maxTime)
			if err != nil {
				log.Printf("Warning: Failed to get Google Calendar busy slots: %v", err)
			} else {
				googleBusySlots = busySlots
			}
		}
	}

	// Mark booked, blocked, and Google Calendar busy slots as unavailable
	for i := range slots {
		// Parse slot time to compare as Unix timestamp
		slotTime, err := time.Parse(time.RFC3339, slots[i].SlotTime)
		if err != nil {
			continue
		}
		unixTime := slotTime.Unix()

		// Check database bookings and blocks
		if bookedSlots[unixTime] || blockedSlots[unixTime] {
			slots[i].Available = false
			continue
		}

		// Check Google Calendar busy slots
		if h.GoogleCalendar != nil && len(googleBusySlots) > 0 {
			if h.GoogleCalendar.IsSlotBusy(slotTime, duration, googleBusySlots) {
				slots[i].Available = false
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}

func (h *APIHandlers) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.SlotTime == "" {
		http.Error(w, "Name, email, and slot_time are required", http.StatusBadRequest)
		return
	}

	slotTime, err := time.Parse(time.RFC3339, req.SlotTime)
	if err != nil {
		http.Error(w, "Invalid slot_time format", http.StatusBadRequest)
		return
	}

	// Check if slot is in the past
	if slotTime.Before(time.Now()) {
		http.Error(w, "Cannot book past slots", http.StatusBadRequest)
		return
	}

	// Convert to UTC for consistent storage
	slotTimeUTC := slotTime.UTC()

	// Check if slot is already booked BEFORE creating Zoom meeting
	var existingCount int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM bookings WHERE slot_time = $1", slotTimeUTC).Scan(&existingCount)
	if err != nil {
		http.Error(w, "Failed to check slot availability", http.StatusInternalServerError)
		log.Printf("Error checking slot availability: %v", err)
		return
	}
	if existingCount > 0 {
		http.Error(w, "Slot already booked", http.StatusConflict)
		return
	}

	// Check if slot is blocked
	var blockedCount int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM blocked_slots WHERE slot_time = $1", slotTimeUTC).Scan(&blockedCount)
	if err != nil {
		http.Error(w, "Failed to check slot availability", http.StatusInternalServerError)
		log.Printf("Error checking blocked slots: %v", err)
		return
	}
	if blockedCount > 0 {
		http.Error(w, "Slot is not available", http.StatusConflict)
		return
	}

	// Create Zoom meeting AFTER confirming slot is available
	var zoomLink string
	createZoomMeeting := strings.ToLower(os.Getenv("CREATE_ZOOM_MEETING"))
	if (createZoomMeeting == "yes" || createZoomMeeting == "true") && h.ZoomService != nil {
		zoomLink, err = h.ZoomService.CreateMeeting(req.Name, req.Email, slotTime)
		if err != nil {
			// Log the error but don't fail the booking
			log.Printf("Warning: Failed to create Zoom meeting: %v", err)
		}
	}

	// Insert booking into database with zoom_link and phone (store in UTC)
	_, err = h.DB.Exec(
		"INSERT INTO bookings (slot_time, name, email, zoom_link, phone) VALUES ($1, $2, $3, $4, $5)",
		slotTimeUTC, req.Name, req.Email, sql.NullString{String: zoomLink, Valid: zoomLink != ""}, sql.NullString{String: req.Phone, Valid: req.Phone != ""},
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
			// Race condition: another request booked the slot between our check and insert
			http.Error(w, "Slot already booked", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create booking", http.StatusInternalServerError)
			log.Printf("Error creating booking: %v", err)
		}
		return
	}

	// Default duration to 30 minutes if not specified
	duration := req.Duration
	if duration != 60 {
		duration = 30
	}

	// Send confirmation email to client if enabled
	sendConfirmationEmail := strings.ToLower(os.Getenv("SEND_CONFIRMATION_EMAIL"))
	if (sendConfirmationEmail == "yes" || sendConfirmationEmail == "true") && h.EmailService != nil {
		err = h.EmailService.SendBookingConfirmation(req.Name, req.Email, slotTime, zoomLink, duration)
		if err != nil {
			// Log the error but don't fail the booking
			log.Printf("Warning: Booking created but failed to send confirmation email: %v", err)
		}
	}

	// Send notification email to calendar owner
	if h.EmailService != nil {
		err = h.EmailService.SendBookingNotificationToOwner(req.Name, req.Email, req.Phone, slotTime, zoomLink)
		if err != nil {
			// Log the error but don't fail the booking
			log.Printf("Warning: Failed to send booking notification to owner: %v", err)
		}
	}

	// Add event to Google Calendar if enabled
	if h.GoogleCalendar != nil {
		err = h.GoogleCalendar.CreateEvent(req.Name, req.Email, req.Phone, slotTime, zoomLink, duration)
		if err != nil {
			// Log the error but don't fail the booking
			log.Printf("Warning: Failed to create Google Calendar event: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Booking created successfully",
	})
}

func (h *APIHandlers) GetAdminSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slots := h.GenerateAvailableSlots()
	adminSlots := make([]AdminSlot, 0, len(slots))

	// Track which slot times we've already processed
	processedSlots := make(map[int64]bool)

	// Calculate today's start (midnight) for filtering
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Get booked slots with booking info (only from today onwards)
	bookedMap := make(map[int64]Booking)
	bookingRows, err := h.DB.Query("SELECT slot_time, name, email, phone FROM bookings WHERE slot_time >= $1", todayStart)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying bookings: %v", err)
		return
	}
	defer bookingRows.Close()

	for bookingRows.Next() {
		var slotTime time.Time
		var name, email string
		var phone sql.NullString
		if err := bookingRows.Scan(&slotTime, &name, &email, &phone); err != nil {
			continue
		}
		// Use Unix timestamp for timezone-independent comparison
		booking := Booking{
			SlotTime: slotTime,
			Name:     name,
			Email:    email,
		}
		if phone.Valid {
			booking.Phone = phone.String
		}
		bookedMap[slotTime.Unix()] = booking
	}

	// Get blocked slots
	blockedMap := make(map[int64]bool)
	blockedRows, err := h.DB.Query("SELECT slot_time FROM blocked_slots")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying blocked slots: %v", err)
		return
	}
	defer blockedRows.Close()

	for blockedRows.Next() {
		var slotTime time.Time
		if err := blockedRows.Scan(&slotTime); err != nil {
			continue
		}
		// Use Unix timestamp for timezone-independent comparison
		blockedMap[slotTime.Unix()] = true
	}

	// First, add today's booked slots that won't be in the generated slots
	// (since generateAvailableSlots now excludes today for user booking)
	for unixTime, booking := range bookedMap {
		adminSlot := AdminSlot{
			SlotTime: booking.SlotTime.Format(time.RFC3339),
			Status:   "booked",
			Name:     booking.Name,
			Email:    booking.Email,
			Phone:    booking.Phone,
		}
		adminSlots = append(adminSlots, adminSlot)
		processedSlots[unixTime] = true
	}

	// Build admin slots response from generated slots
	for _, slot := range slots {
		// Parse slot time for timezone-independent comparison
		slotTime, err := time.Parse(time.RFC3339, slot.SlotTime)
		if err != nil {
			continue
		}

		unixTime := slotTime.Unix()

		// Skip if already processed (from bookings)
		if processedSlots[unixTime] {
			continue
		}

		adminSlot := AdminSlot{
			SlotTime: slot.SlotTime,
			Status:   "available",
		}

		if blockedMap[unixTime] {
			adminSlot.Status = "blocked"
		}

		adminSlots = append(adminSlots, adminSlot)
		processedSlots[unixTime] = true
	}

	// Sort admin slots by time
	sort.Slice(adminSlots, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, adminSlots[i].SlotTime)
		timeJ, _ := time.Parse(time.RFC3339, adminSlots[j].SlotTime)
		return timeI.Before(timeJ)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(adminSlots)
}

func (h *APIHandlers) BlockSlot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SlotTime string `json:"slot_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slotTime, err := time.Parse(time.RFC3339, req.SlotTime)
	if err != nil {
		http.Error(w, "Invalid slot_time format", http.StatusBadRequest)
		return
	}

	// Convert to UTC for consistent storage (SQLite driver doesn't handle timezones well)
	slotTimeUTC := slotTime.UTC()

	// Check if slot is already booked
	var count int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM bookings WHERE slot_time = $1", slotTimeUTC).Scan(&count)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error checking booking: %v", err)
		return
	}

	if count > 0 {
		http.Error(w, "Cannot block a slot that is already booked", http.StatusConflict)
		return
	}

	// Insert blocked slot (store in UTC)
	_, err = h.DB.Exec("INSERT INTO blocked_slots (slot_time) VALUES ($1)", slotTimeUTC)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "duplicate key value") || strings.Contains(errMsg, "unique constraint") {
			http.Error(w, "Slot already blocked", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Failed to block slot: %v", err), http.StatusInternalServerError)
			log.Printf("Error blocking slot: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Slot blocked successfully",
	})
}

func (h *APIHandlers) UnblockSlot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SlotTime string `json:"slot_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slotTime, err := time.Parse(time.RFC3339, req.SlotTime)
	if err != nil {
		http.Error(w, "Invalid slot_time format", http.StatusBadRequest)
		return
	}

	// Convert to UTC (database stores times in UTC)
	slotTimeUTC := slotTime.UTC()

	// Delete blocked slot
	result, err := h.DB.Exec("DELETE FROM blocked_slots WHERE slot_time = $1", slotTimeUTC)
	if err != nil {
		http.Error(w, "Failed to unblock slot", http.StatusInternalServerError)
		log.Printf("Error unblocking slot: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Slot not found in blocked list", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Slot unblocked successfully",
	})
}

func (h *APIHandlers) DebugBlockedSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type DebugInfo struct {
		SlotTimeRaw     string `json:"slot_time_raw"`
		SlotTimeUnix    int64  `json:"slot_time_unix"`
		SlotTimeRFC3339 string `json:"slot_time_rfc3339"`
		Location        string `json:"location"`
	}

	rows, err := h.DB.Query("SELECT slot_time FROM blocked_slots WHERE slot_time LIKE '2026-%' LIMIT 20")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var debugInfo []DebugInfo
	for rows.Next() {
		var slotTime time.Time
		if err := rows.Scan(&slotTime); err != nil {
			continue
		}
		debugInfo = append(debugInfo, DebugInfo{
			SlotTimeRaw:     slotTime.String(),
			SlotTimeUnix:    slotTime.Unix(),
			SlotTimeRFC3339: slotTime.Format(time.RFC3339),
			Location:        slotTime.Location().String(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debugInfo)
}

func (h *APIHandlers) ClearAllBlockedSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delete all blocked slots
	result, err := h.DB.Exec("DELETE FROM blocked_slots")
	if err != nil {
		http.Error(w, "Failed to clear blocked slots", http.StatusInternalServerError)
		log.Printf("Error clearing blocked slots: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "All blocked slots cleared",
		"rows_affected": rowsAffected,
	})
}

func (h *APIHandlers) CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SlotTime string `json:"slot_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slotTime, err := time.Parse(time.RFC3339, req.SlotTime)
	if err != nil {
		http.Error(w, "Invalid slot_time format", http.StatusBadRequest)
		return
	}

	// Convert to UTC for database query
	slotTimeUTC := slotTime.UTC()

	// Get the booking details (including zoom_link) before deleting
	var zoomLink sql.NullString
	var name, email string
	err = h.DB.QueryRow(
		"SELECT name, email, zoom_link FROM bookings WHERE slot_time = $1",
		slotTimeUTC,
	).Scan(&name, &email, &zoomLink)

	if err == sql.ErrNoRows {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying booking: %v", err)
		return
	}

	// Delete the Zoom meeting if one exists
	if zoomLink.Valid && zoomLink.String != "" && h.ZoomService != nil {
		if err := h.ZoomService.DeleteMeeting(zoomLink.String); err != nil {
			// Log the error but don't fail the booking cancellation
			log.Printf("Warning: Failed to delete Zoom meeting for booking %s: %v", req.SlotTime, err)
		}
	}

	// Delete the booking from database
	result, err := h.DB.Exec("DELETE FROM bookings WHERE slot_time = $1", slotTimeUTC)
	if err != nil {
		http.Error(w, "Failed to cancel booking", http.StatusInternalServerError)
		log.Printf("Error deleting booking: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	log.Printf("Booking cancelled successfully: %s - %s (%s)", req.SlotTime, name, email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Booking cancelled successfully",
	})
}

// GetClients returns all clients
func (h *APIHandlers) GetClients(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.DB.Query(`
		SELECT id, full_name, email, phone_number, telegram_id, notes, created_at, updated_at
		FROM clients
		ORDER BY full_name ASC
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying clients: %v", err)
		return
	}
	defer rows.Close()

	clients := []Client{}
	for rows.Next() {
		var client Client
		var email, phoneNumber, telegramID, notes sql.NullString
		err := rows.Scan(
			&client.ID,
			&client.FullName,
			&email,
			&phoneNumber,
			&telegramID,
			&notes,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning client: %v", err)
			continue
		}

		if email.Valid {
			client.Email = email.String
		}
		if phoneNumber.Valid {
			client.PhoneNumber = phoneNumber.String
		}
		if telegramID.Valid {
			client.TelegramID = telegramID.String
		}
		if notes.Valid {
			client.Notes = notes.String
		}

		clients = append(clients, client)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

// CreateClient creates a new client
func (h *APIHandlers) CreateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FullName == "" {
		http.Error(w, "full_name is required", http.StatusBadRequest)
		return
	}

	var id int
	err := h.DB.QueryRow(`
		INSERT INTO clients (full_name, email, phone_number, telegram_id, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`,
		req.FullName,
		sql.NullString{String: req.Email, Valid: req.Email != ""},
		sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		sql.NullString{String: req.TelegramID, Valid: req.TelegramID != ""},
		sql.NullString{String: req.Notes, Valid: req.Notes != ""},
	).Scan(&id)

	if err != nil {
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		log.Printf("Error creating client: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Client created successfully",
		"id":      id,
	})
}

// UpdateClient updates an existing client
func (h *APIHandlers) UpdateClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get client ID from query parameter
	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	var req ClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FullName == "" {
		http.Error(w, "full_name is required", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		UPDATE clients
		SET full_name = $1, email = $2, phone_number = $3, telegram_id = $4, notes = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`,
		req.FullName,
		sql.NullString{String: req.Email, Valid: req.Email != ""},
		sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		sql.NullString{String: req.TelegramID, Valid: req.TelegramID != ""},
		sql.NullString{String: req.Notes, Valid: req.Notes != ""},
		clientID,
	)

	if err != nil {
		http.Error(w, "Failed to update client", http.StatusInternalServerError)
		log.Printf("Error updating client: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Client updated successfully",
	})
}

// DeleteClient deletes a client
func (h *APIHandlers) DeleteClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get client ID from query parameter
	clientID := r.URL.Query().Get("id")
	if clientID == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("DELETE FROM clients WHERE id = $1", clientID)
	if err != nil {
		http.Error(w, "Failed to delete client", http.StatusInternalServerError)
		log.Printf("Error deleting client: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Client deleted successfully",
	})
}
