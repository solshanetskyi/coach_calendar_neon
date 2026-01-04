package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// EmailSender interface for sending emails
type EmailSender interface {
	SendBookingConfirmation(name, email string, slotTime time.Time, zoomLink string) error
}

// ZoomMeetingCreator interface for creating Zoom meetings
type ZoomMeetingCreator interface {
	CreateMeeting(name, email string, slotTime time.Time) (string, error)
}

// Re-export types from main package
type Booking struct {
	ID        int       `json:"id"`
	SlotTime  time.Time `json:"slot_time"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	ZoomLink  string    `json:"zoom_link,omitempty"`
}

type BookingRequest struct {
	SlotTime string `json:"slot_time"`
	Name     string `json:"name"`
	Email    string `json:"email"`
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
}

type GenerateSlotsFn func() []AvailableSlot

type APIHandlers struct {
	DB                     *sql.DB
	GenerateAvailableSlots GenerateSlotsFn
	EmailService           EmailSender
	ZoomService            ZoomMeetingCreator
}

func NewAPIHandlers(db *sql.DB, generateSlotsFn GenerateSlotsFn, emailService EmailSender, zoomService ZoomMeetingCreator) *APIHandlers {
	return &APIHandlers{
		DB:                     db,
		GenerateAvailableSlots: generateSlotsFn,
		EmailService:           emailService,
		ZoomService:            zoomService,
	}
}

func (h *APIHandlers) GetSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only 30-minute slots are available
	slots := h.GenerateAvailableSlots()

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

	// Mark booked and blocked slots as unavailable
	for i := range slots {
		// Parse slot time to compare as Unix timestamp
		slotTime, err := time.Parse(time.RFC3339, slots[i].SlotTime)
		if err != nil {
			continue
		}
		unixTime := slotTime.Unix()
		if bookedSlots[unixTime] || blockedSlots[unixTime] {
			slots[i].Available = false
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

	// Create Zoom meeting first (before database insert) if enabled
	var zoomLink string
	createZoomMeeting := strings.ToLower(os.Getenv("CREATE_ZOOM_MEETING"))
	if (createZoomMeeting == "yes" || createZoomMeeting == "true") && h.ZoomService != nil {
		zoomLink, err = h.ZoomService.CreateMeeting(req.Name, req.Email, slotTime)
		if err != nil {
			// Log the error but don't fail the booking
			log.Printf("Warning: Failed to create Zoom meeting: %v", err)
		}
	}

	// Insert booking into database with zoom_link (store in UTC)
	_, err = h.DB.Exec(
		"INSERT INTO bookings (slot_time, name, email, zoom_link) VALUES ($1, $2, $3, $4)",
		slotTimeUTC, req.Name, req.Email, sql.NullString{String: zoomLink, Valid: zoomLink != ""},
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") || strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Slot already booked", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create booking", http.StatusInternalServerError)
			log.Printf("Error creating booking: %v", err)
		}
		return
	}

	// Send confirmation email if enabled
	sendConfirmationEmail := strings.ToLower(os.Getenv("SEND_CONFIRMATION_EMAIL"))
	if (sendConfirmationEmail == "yes" || sendConfirmationEmail == "true") && h.EmailService != nil {
		err = h.EmailService.SendBookingConfirmation(req.Name, req.Email, slotTime, zoomLink)
		if err != nil {
			// Log the error but don't fail the booking
			log.Printf("Warning: Booking created but failed to send confirmation email: %v", err)
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

	// Get booked slots with booking info
	bookedMap := make(map[int64]Booking)
	bookingRows, err := h.DB.Query("SELECT slot_time, name, email FROM bookings")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Error querying bookings: %v", err)
		return
	}
	defer bookingRows.Close()

	for bookingRows.Next() {
		var slotTime time.Time
		var name, email string
		if err := bookingRows.Scan(&slotTime, &name, &email); err != nil {
			continue
		}
		// Use Unix timestamp for timezone-independent comparison
		bookedMap[slotTime.Unix()] = Booking{
			SlotTime: slotTime,
			Name:     name,
			Email:    email,
		}
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

	// Build admin slots response
	for _, slot := range slots {
		adminSlot := AdminSlot{
			SlotTime: slot.SlotTime,
			Status:   "available",
		}

		// Parse slot time for timezone-independent comparison
		slotTime, err := time.Parse(time.RFC3339, slot.SlotTime)
		if err != nil {
			continue
		}

		unixTime := slotTime.Unix()
		if booking, ok := bookedMap[unixTime]; ok {
			adminSlot.Status = "booked"
			adminSlot.Name = booking.Name
			adminSlot.Email = booking.Email
		} else if blockedMap[unixTime] {
			adminSlot.Status = "blocked"
		}

		adminSlots = append(adminSlots, adminSlot)
	}

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
