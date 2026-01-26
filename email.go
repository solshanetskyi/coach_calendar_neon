package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type EmailService struct {
	SMTPHost string
	SMTPPort string
	From     string
	FromName string
	Password string
	Enabled  bool
}

func NewEmailService() *EmailService {
	// SMTP Configuration
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")

	// User-friendly "From" name with default
	fromName := os.Getenv("SMTP_USER_FRIENDLY_FROM")
	if fromName == "" {
		fromName = "Христина Івасюк"
	}

	enabled := host != "" && port != "" && from != "" && password != ""

	if !enabled {
		log.Println("Email service disabled - SMTP configuration not found")
		log.Println("To enable SMTP email confirmations, set: SMTP_HOST, SMTP_PORT, SMTP_FROM, SMTP_PASSWORD")
	} else {
		log.Printf("Email service enabled using SMTP (from: %s <%s>)", fromName, from)
	}

	return &EmailService{
		SMTPHost: host,
		SMTPPort: port,
		From:     from,
		FromName: fromName,
		Password: password,
		Enabled:  enabled,
	}
}

// getZoomSection returns the HTML for the Zoom meeting section, or empty string if no link
func getZoomSection(zoomLink string) string {
	if zoomLink == "" {
		return ""
	}

	return fmt.Sprintf(`
            <div class="calendar-section" style="background: #e3f2fd; border-left: 4px solid #2D8CFF;">
                <h3>🎥 Онлайн зустріч Zoom:</h3>
                <p>
                    <a href="%s" class="btn" target="_blank" style="display: inline-block; padding: 12px 24px; background: #2D8CFF; color: #ffffff !important; text-decoration: none; border-radius: 6px; margin: 10px;">Приєднатися до Zoom</a>
                </p>
                <p style="font-size: 14px; color: #666;">
                    Посилання на зустріч буде активне за 10 хвилин до початку
                </p>
            </div>`, zoomLink)
}

// generateGoogleCalendarURL creates a Google Calendar event URL
func generateGoogleCalendarURL(slotTime time.Time) string {
	endTime := slotTime.Add(30 * time.Minute)
	startUTC := slotTime.UTC().Format("20060102T150405Z")
	endUTC := endTime.UTC().Format("20060102T150405Z")
	url := fmt.Sprintf("https://calendar.google.com/calendar/render?action=TEMPLATE&text=%s&dates=%s/%s",
		"Coaching+Session",
		startUTC,
		endUTC,
	)
	return url
}

// generateICalendar creates an iCalendar (ICS) format string for the appointment
func generateICalendar(name, email string, slotTime time.Time) string {
	// Calculate end time (30 minutes after start)
	endTime := slotTime.Add(30 * time.Minute)

	// Format times in iCalendar format (YYYYMMDDTHHMMSSZ in UTC)
	startUTC := slotTime.UTC().Format("20060102T150405Z")
	endUTC := endTime.UTC().Format("20060102T150405Z")
	now := time.Now().UTC().Format("20060102T150405Z")

	// Generate a unique ID for the event
	eventID := fmt.Sprintf("%d@coach-calendar.com", time.Now().UnixNano())

	// Create iCalendar content
	ical := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Coach Calendar//Booking System//EN
CALSCALE:GREGORIAN
METHOD:REQUEST
BEGIN:VEVENT
UID:%s
DTSTAMP:%s
DTSTART:%s
DTEND:%s
SUMMARY:Онлайн консультація з %s
DESCRIPTION:Your coaching appointment has been confirmed.\n\nClient: %s\nEmail: %s
LOCATION:Online/TBD
STATUS:CONFIRMED
SEQUENCE:0
BEGIN:VALARM
TRIGGER:-PT15M
ACTION:DISPLAY
DESCRIPTION:Reminder: Онлайн консультація з %s починається через 15 хвилин
END:VALARM
END:VEVENT
END:VCALENDAR`, eventID, now, startUTC, endUTC, name, name, email)

	return ical
}

func (e *EmailService) SendBookingConfirmation(name, email string, slotTime time.Time, zoomLink string) error {
	if !e.Enabled {
		log.Printf("Email service disabled - skipping confirmation email to %s", email)
		return nil
	}

	// Format the booking time
	formattedTime := slotTime.Format("Monday, January 2, 2006 at 3:04 PM MST")

	// Generate Google Calendar URL
	googleCalURL := generateGoogleCalendarURL(slotTime)

	// Create email subject and body
	subject := "Підтвердження онлайн-запису - безкоштовна консультація з Христиною Івасюк"

	// HTML body
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #800020 0%%, #5c0011 100%%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .details { background: white; padding: 20px; border-left: 4px solid #800020; margin: 20px 0; }
        .detail-row { margin: 10px 0; }
        .calendar-section { background: white; padding: 20px; margin: 20px 0; text-align: center; border-radius: 8px; }
        .btn { display: inline-block; padding: 12px 24px; background: #800020; color: white !important; text-decoration: none; border-radius: 6px; margin: 10px; }
        .btn:hover { background: #5c0011; color: white !important; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Підтвердження онлайн-запису</h1>
            <p>Безкоштовна консультація з Христиною Івасюк</p>
        </div>
        <div class="content">
            <p>Вітаємо, <strong>%s</strong>!</p>
            <p>Дякуємо за бронювання зустрічі!</p>

            <div class="details">
                <h3>Деталі зустрічі:</h3>
                <div class="detail-row">📅 <strong>Дата і час:</strong> %s</div>
                <div class="detail-row">⏱️ <strong>Тривалість:</strong> 30 хвилин</div>
                <div class="detail-row">👤 <strong>Ім'я:</strong> %s</div>
                <div class="detail-row">📧 <strong>Email:</strong> %s</div>
            </div>

            %s

            <div class="calendar-section">
                <h3>📅 Додати до календаря:</h3>
                <p>
                    <a href="%s" class="btn" target="_blank" style="display: inline-block; padding: 12px 24px; background: #800020; color: #ffffff !important; text-decoration: none; border-radius: 6px; margin: 10px;">Додати в Google Calendar</a>
                </p>
                <p style="font-size: 14px; color: #666;">
                    Або відкрийте прикріплений файл invite.ics для інших календарів
                    <br>(Outlook, Apple Calendar тощо)
                </p>
            </div>

            <p>Будь ласка, приходьте вчасно на вашу зустріч.</p>
            <p>Якщо вам потрібно скасувати або перенести зустріч, будь ласка, зв'яжіться зі мною якнайшвидше.</p>

            <p style="margin-top: 30px;">
                З повагою,<br>
                <strong>Христина Івасюк</strong>
            </p>
        </div>
        <div class="footer">
            Це автоматичне повідомлення. Будь ласка, не відповідайте на цей email.
        </div>
    </div>
</body>
</html>`, name, formattedTime, name, email, getZoomSection(zoomLink), googleCalURL)

	// Plain text fallback
	zoomText := ""
	if zoomLink != "" {
		zoomText = fmt.Sprintf(`
🎥 Онлайн зустріч Zoom:
%s

`, zoomLink)
	}

	textBody := fmt.Sprintf(`Вітаємо, %s!

Дякую за бронювання зустрічі!

Деталі зустрічі:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📅 Дата і час: %s
⏱️ Тривалість: 30 хвилин
👤 Ім'я: %s
📧 Email: %s
%s
📅 Додати до календаря:
%s

Або відкрийте прикріплений файл invite.ics для інших календарів.

Будь ласка, приходьте вчасно на вашу зустріч.

Якщо вам потрібно скасувати або перенести зустріч, будь ласка, зв'яжіться зі мною якнайшвидше.

З повагою,
Христина Івасюк

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Це автоматичне повідомлення. Будь ласка, не відповідайте на цей email.
`, name, formattedTime, name, email, zoomText, googleCalURL)

	// Generate iCalendar attachment
	icalContent := generateICalendar(name, email, slotTime)

	// Send via SMTP
	return e.sendViaSMTP(email, subject, htmlBody, textBody, icalContent)
}

func (e *EmailService) sendViaSMTP(toEmail, subject, htmlBody, textBody, icalContent string) error {
	// Create boundaries for multipart message
	mixedBoundary := fmt.Sprintf("mixed_boundary_%d", rand.Int63())
	altBoundary := fmt.Sprintf("alt_boundary_%d", rand.Int63())

	// Build multipart email with HTML, text fallback, and calendar attachment
	var message strings.Builder
	// Use RFC 5322 format with base64-encoded UTF-8 display name
	encodedFromName := base64.StdEncoding.EncodeToString([]byte(e.FromName))
	message.WriteString(fmt.Sprintf("From: =?UTF-8?B?%s?= <%s>\r\n", encodedFromName, e.From))
	message.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", mixedBoundary))
	message.WriteString("\r\n")

	// Start multipart/alternative section for HTML and text
	message.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
	message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", altBoundary))
	message.WriteString("\r\n")

	// Plain text version
	message.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	message.WriteString("\r\n")
	message.WriteString(textBody)
	message.WriteString("\r\n\r\n")

	// HTML version
	message.WriteString(fmt.Sprintf("--%s\r\n", altBoundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n\r\n")

	// End multipart/alternative section
	message.WriteString(fmt.Sprintf("--%s--\r\n", altBoundary))
	message.WriteString("\r\n")

	// Calendar attachment part
	message.WriteString(fmt.Sprintf("--%s\r\n", mixedBoundary))
	message.WriteString("Content-Type: text/calendar; charset=UTF-8; method=REQUEST; name=\"invite.ics\"\r\n")
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString("Content-Disposition: attachment; filename=\"invite.ics\"\r\n")
	message.WriteString("\r\n")
	message.WriteString(base64.StdEncoding.EncodeToString([]byte(icalContent)))
	message.WriteString("\r\n")
	message.WriteString(fmt.Sprintf("--%s--\r\n", mixedBoundary))

	// Set up authentication
	auth := smtp.PlainAuth("", e.From, e.Password, e.SMTPHost)

	// Send the email
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	err := smtp.SendMail(addr, auth, e.From, []string{toEmail}, []byte(message.String()))
	if err != nil {
		log.Printf("Failed to send confirmation email via SMTP to %s: %v", toEmail, err)
		return fmt.Errorf("failed to send confirmation email via SMTP: %w", err)
	}

	log.Printf("Confirmation email sent successfully via SMTP to %s", toEmail)
	return nil
}

// SendBookingNotificationToOwner sends a notification email to the calendar owner about a new booking
func (e *EmailService) SendBookingNotificationToOwner(clientName, clientEmail, clientPhone string, slotTime time.Time, zoomLink string) error {
	if !e.Enabled {
		log.Printf("Email service disabled - skipping owner notification")
		return nil
	}

	// Send to the calendar owner (SMTP_FROM)
	ownerEmail := e.From

	// Convert to CET timezone for display
	cetLocation, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		log.Printf("Warning: Could not load CET timezone, using UTC: %v", err)
		cetLocation = time.UTC
	}
	slotTimeCET := slotTime.In(cetLocation)

	// Format the booking time in CET
	formattedTime := slotTimeCET.Format("Monday, January 2, 2006 at 15:04 CET")

	// Create email subject
	subject := fmt.Sprintf("Нове бронювання: %s - %s", clientName, slotTimeCET.Format("02.01.2006 15:04"))

	// Phone info
	phoneInfo := "Не вказано"
	if clientPhone != "" {
		phoneInfo = clientPhone
	}

	// Zoom link info
	zoomInfo := ""
	if zoomLink != "" {
		zoomInfo = fmt.Sprintf(`
            <div class="detail-row">🎥 <strong>Zoom:</strong> <a href="%s">%s</a></div>`, zoomLink, zoomLink)
	}

	// HTML body
	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #2e7d32 0%%, #1b5e20 100%%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .details { background: white; padding: 20px; border-left: 4px solid #2e7d32; margin: 20px 0; }
        .detail-row { margin: 10px 0; }
        .footer { text-align: center; color: #666; font-size: 12px; margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Нове бронювання!</h1>
        </div>
        <div class="content">
            <p>Отримано нове бронювання на консультацію.</p>

            <div class="details">
                <h3>Деталі бронювання:</h3>
                <div class="detail-row">📅 <strong>Дата і час:</strong> %s</div>
                <div class="detail-row">⏱️ <strong>Тривалість:</strong> 30 хвилин</div>
                <div class="detail-row">👤 <strong>Клієнт:</strong> %s</div>
                <div class="detail-row">📧 <strong>Email:</strong> <a href="mailto:%s">%s</a></div>
                <div class="detail-row">📱 <strong>Телефон:</strong> %s</div>%s
            </div>

            <p style="margin-top: 20px; color: #666;">
                Клієнту надіслано підтвердження на email.
            </p>
        </div>
        <div class="footer">
            Coach Calendar - Автоматичне сповіщення
        </div>
    </div>
</body>
</html>`, formattedTime, clientName, clientEmail, clientEmail, phoneInfo, zoomInfo)

	// Plain text version
	zoomText := ""
	if zoomLink != "" {
		zoomText = fmt.Sprintf("🎥 Zoom: %s\n", zoomLink)
	}

	textBody := fmt.Sprintf(`🎉 Нове бронювання!

Отримано нове бронювання на консультацію.

Деталі бронювання:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📅 Дата і час: %s
⏱️ Тривалість: 30 хвилин
👤 Клієнт: %s
📧 Email: %s
📱 Телефон: %s
%s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Клієнту надіслано підтвердження на email.

--
Coach Calendar - Автоматичне сповіщення
`, formattedTime, clientName, clientEmail, phoneInfo, zoomText)

	// Send via SMTP (no calendar attachment for owner notification)
	return e.sendSimpleEmail(ownerEmail, subject, htmlBody, textBody)
}

// sendSimpleEmail sends a simple email without calendar attachment
func (e *EmailService) sendSimpleEmail(toEmail, subject, htmlBody, textBody string) error {
	// Create boundary for multipart message
	boundary := fmt.Sprintf("boundary_%d", rand.Int63())

	// Build multipart email with HTML and text
	var message strings.Builder
	encodedFromName := base64.StdEncoding.EncodeToString([]byte(e.FromName))
	message.WriteString(fmt.Sprintf("From: =?UTF-8?B?%s?= <%s>\r\n", encodedFromName, e.From))
	message.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	message.WriteString(fmt.Sprintf("Subject: =?UTF-8?B?%s?=\r\n", base64.StdEncoding.EncodeToString([]byte(subject))))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	message.WriteString("\r\n")

	// Plain text version
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString("\r\n")
	message.WriteString(base64.StdEncoding.EncodeToString([]byte(textBody)))
	message.WriteString("\r\n\r\n")

	// HTML version
	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: base64\r\n")
	message.WriteString("\r\n")
	message.WriteString(base64.StdEncoding.EncodeToString([]byte(htmlBody)))
	message.WriteString("\r\n\r\n")

	// End multipart
	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Set up authentication
	auth := smtp.PlainAuth("", e.From, e.Password, e.SMTPHost)

	// Send the email
	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	err := smtp.SendMail(addr, auth, e.From, []string{toEmail}, []byte(message.String()))
	if err != nil {
		log.Printf("Failed to send email via SMTP to %s: %v", toEmail, err)
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	log.Printf("Email sent successfully via SMTP to %s", toEmail)
	return nil
}
