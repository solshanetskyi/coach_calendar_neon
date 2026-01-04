package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type ZoomService struct {
	AccountID    string
	ClientID     string
	ClientSecret string
	AccessToken  string
	TokenExpiry  time.Time
	Enabled      bool
	mu           sync.Mutex
}

type zoomTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type zoomMeetingRequest struct {
	Topic     string              `json:"topic"`
	Type      int                 `json:"type"`
	StartTime string              `json:"start_time"`
	Duration  int                 `json:"duration"`
	Timezone  string              `json:"timezone"`
	Agenda    string              `json:"agenda"`
	Settings  zoomMeetingSettings `json:"settings"`
}

type zoomMeetingSettings struct {
	HostVideo        bool   `json:"host_video"`
	ParticipantVideo bool   `json:"participant_video"`
	JoinBeforeHost   bool   `json:"join_before_host"`
	MuteUponEntry    bool   `json:"mute_upon_entry"`
	WaitingRoom      bool   `json:"waiting_room"`
	AutoRecording    string `json:"auto_recording"`
}

type zoomMeetingResponse struct {
	ID       int64  `json:"id"`
	JoinURL  string `json:"join_url"`
	StartURL string `json:"start_url"`
}

func NewZoomService() *ZoomService {
	accountID := os.Getenv("ZOOM_ACCOUNT_ID")
	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_SECRET")

	enabled := accountID != "" && clientID != "" && clientSecret != ""

	if !enabled {
		log.Println("Zoom integration disabled - configuration not found")
		log.Println("To enable Zoom integration, set: ZOOM_ACCOUNT_ID, ZOOM_CLIENT_ID, ZOOM_CLIENT_SECRET")
		return &ZoomService{Enabled: false}
	}

	log.Println("Zoom integration enabled")
	return &ZoomService{
		AccountID:    accountID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Enabled:      enabled,
	}
}

func (z *ZoomService) getAccessToken() (string, error) {
	z.mu.Lock()
	defer z.mu.Unlock()

	// Return cached token if still valid
	if z.AccessToken != "" && time.Now().Before(z.TokenExpiry) {
		return z.AccessToken, nil
	}

	// Request new token using Server-to-Server OAuth
	tokenURL := "https://zoom.us/oauth/token"
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", z.AccountID)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(z.ClientID, z.ClientSecret)

	log.Printf("Requesting Zoom OAuth token for account: %s", z.AccountID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Zoom OAuth error - Status: %d, Response: %s", resp.StatusCode, string(body))
		log.Printf("Debug - Account ID length: %d, Client ID length: %d, Client Secret length: %d",
			len(z.AccountID), len(z.ClientID), len(z.ClientSecret))
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp zoomTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Cache the token
	z.AccessToken = tokenResp.AccessToken
	z.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second) // Refresh 5 min early

	log.Println("Successfully obtained Zoom access token")
	return z.AccessToken, nil
}

func (z *ZoomService) CreateMeeting(name, email string, slotTime time.Time) (string, error) {
	if !z.Enabled {
		log.Println("Zoom is disabled - skipping meeting creation")
		return "", nil
	}

	// Get access token
	token, err := z.getAccessToken()
	if err != nil {
		log.Printf("Failed to get Zoom access token: %v", err)
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	// Format the meeting topic
	topic := fmt.Sprintf("Онлайн-консультація - %s", name)

	// Prepare meeting request
	meetingReq := zoomMeetingRequest{
		Topic:     topic,
		Type:      2, // Scheduled meeting
		StartTime: slotTime.UTC().Format("2006-01-02T15:04:05Z"),
		Duration:  30,
		Timezone:  "UTC",
		Agenda:    fmt.Sprintf("Онлайн-консультація з %s (%s)", name, email),
		Settings: zoomMeetingSettings{
			HostVideo:        true,
			ParticipantVideo: true,
			JoinBeforeHost:   false,
			MuteUponEntry:    true,
			WaitingRoom:      true,
			AutoRecording:    "none",
		},
	}

	// Marshal request body
	body, err := json.Marshal(meetingReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal meeting request: %w", err)
	}

	// Create HTTP request
	apiURL := "https://api.zoom.us/v2/users/me/meetings"
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create meeting request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send meeting request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Zoom API error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
		return "", fmt.Errorf("zoom API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var meetingResp zoomMeetingResponse
	if err := json.NewDecoder(resp.Body).Decode(&meetingResp); err != nil {
		return "", fmt.Errorf("failed to decode meeting response: %w", err)
	}

	log.Printf("Zoom meeting created successfully: %s (ID: %d)", meetingResp.JoinURL, meetingResp.ID)
	return meetingResp.JoinURL, nil
}
