//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	AmsterdamTimezone = "Europe/Amsterdam"
	DefaultAPIURL     = "http://localhost:8080"
)

type BlockSlotRequest struct {
	SlotTime string `json:"slot_time"`
}

func main() {
	// Command line flags
	apiURL := flag.String("api", DefaultAPIURL, "API base URL")
	daysAhead := flag.Int("days", 30, "Number of days ahead to block slots")
	startFrom := flag.String("start", "", "Start date (YYYY-MM-DD), defaults to today")
	allDays := flag.Bool("all-days", false, "Block all days, not just Sundays")
	dryRun := flag.Bool("dry-run", false, "Show what would be blocked without actually blocking")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Block Slots Script for Coach Calendar")
		fmt.Println("\nBlocks all time slots on Sundays (9:00-20:00 Amsterdam time)")
		fmt.Println("\nUsage:")
		fmt.Println("  go run block_slots.go [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  -api string")
		fmt.Println("        API base URL (default \"http://localhost:8080\")")
		fmt.Println("  -days int")
		fmt.Println("        Number of days ahead to block slots (default 30)")
		fmt.Println("  -dry-run")
		fmt.Println("        Show what would be blocked without actually blocking")
		fmt.Println("  -help")
		fmt.Println("        Show this help message")
		fmt.Println("\nExample:")
		fmt.Println("  go run block_slots.go -api http://localhost:8080 -days 30")
		fmt.Println("  go run block_slots.go -api https://your-app.com -days 60 -dry-run")
		return
	}

	// Load Amsterdam timezone
	location, err := time.LoadLocation(AmsterdamTimezone)
	if err != nil {
		log.Fatalf("Failed to load Amsterdam timezone: %v", err)
	}

	// Determine start date
	now := time.Now().In(location)
	var startDate time.Time
	if *startFrom != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *startFrom, location)
		if err != nil {
			log.Fatalf("Invalid -start date %q, expected YYYY-MM-DD: %v", *startFrom, err)
		}
		startDate = parsed
	} else {
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	}

	var slotsToBlock []time.Time

	// Generate slots for the next N days
	for day := 0; day < *daysAhead; day++ {
		currentDate := startDate.AddDate(0, 0, day)
		weekday := currentDate.Weekday()

		// Only block Sundays (unless -all-days is set)
		if !*allDays && weekday != time.Sunday {
			continue
		}

		// Sunday: Block ALL slots (10:00 AM to 8:00 PM)
		startHour := 10
		startMinute := 0
		endHour := 20
		endMinute := 0

		// Generate slots for this day
		currentSlot := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(),
			startHour, startMinute, 0, 0, location)
		endTime := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(),
			endHour, endMinute, 0, 0, location)

		for !currentSlot.After(endTime) {
			slotsToBlock = append(slotsToBlock, currentSlot)
			currentSlot = currentSlot.Add(30 * time.Minute)
		}
	}

	if len(slotsToBlock) == 0 {
		log.Println("No slots to block")
		return
	}

	log.Println("========================================")
	log.Printf("Blocking slots for %d days ahead", *daysAhead)
	log.Println("Blocking schedule:")
	log.Println("  - Sunday: 9:00-20:00 (entire day)")
	log.Printf("Total slots to block: %d", len(slotsToBlock))
	if *dryRun {
		log.Println("DRY RUN MODE - Not actually blocking slots")
	}
	log.Println("========================================")

	successCount := 0
	failCount := 0
	skipCount := 0

	for i, slot := range slotsToBlock {
		slotStr := slot.Format(time.RFC3339)
		weekday := slot.Weekday()

		if *dryRun {
			fmt.Printf("[%d/%d] Would block: %s %s (%s)\n",
				i+1, len(slotsToBlock),
				slot.Format("Mon, Jan 2, 2006 15:04 MST"),
				weekday,
				slotStr)
			successCount++
			continue
		}

		// Make API request to block the slot
		err := blockSlot(*apiURL, slotStr)
		if err != nil {
			// Check if it's already blocked (UNIQUE constraint error)
			if isAlreadyBlockedError(err) {
				fmt.Printf("[%d/%d] ⏭️  Already blocked: %s %s\n",
					i+1, len(slotsToBlock),
					slot.Format("Mon, Jan 2, 2006 15:04 MST"),
					weekday)
				skipCount++
			} else {
				fmt.Printf("[%d/%d] ❌ Failed to block: %s - %v\n",
					i+1, len(slotsToBlock),
					slot.Format("Mon, Jan 2, 2006 15:04 MST"),
					err)
				failCount++
			}
		} else {
			fmt.Printf("[%d/%d] ✅ Blocked: %s %s\n",
				i+1, len(slotsToBlock),
				slot.Format("Mon, Jan 2, 2006 15:04 MST"),
				weekday)
			successCount++
		}

		// Small delay to avoid overwhelming the API
		if i < len(slotsToBlock)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Println("========================================")
	log.Printf("Summary:")
	log.Printf("  ✅ Success: %d", successCount)
	if failCount > 0 {
		log.Printf("  ❌ Failed: %d", failCount)
	}
	if skipCount > 0 {
		log.Printf("  ⏭️  Skipped: %d", skipCount)
	}
	log.Println("========================================")
}

func isAlreadyBlockedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "UNIQUE constraint failed") ||
		strings.Contains(errStr, "already blocked") ||
		strings.Contains(errStr, "duplicate")
}

func blockSlot(apiURL, slotTime string) error {
	// Create request payload
	payload := BlockSlotRequest{
		SlotTime: slotTime,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Make POST request to block slot
	url := fmt.Sprintf("%s/api/admin/block", apiURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := io.ReadAll(resp.Body)

	// Check response status - API returns 201 Created on success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		errMsg := fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))

		// Check if the error is due to duplicate/already blocked
		bodyStr := string(body)
		if strings.Contains(bodyStr, "UNIQUE constraint failed") ||
			strings.Contains(bodyStr, "already blocked") ||
			strings.Contains(bodyStr, "Slot already blocked") ||
			strings.Contains(bodyStr, "duplicate") {
			return fmt.Errorf("already blocked: %s", bodyStr)
		}

		return fmt.Errorf("%s", errMsg)
	}

	return nil
}
