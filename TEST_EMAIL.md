# Test Email Script

This script allows you to test email sending functionality independently from the booking system.

## Prerequisites

Set up your SMTP credentials as environment variables:

```bash
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_FROM="your-email@gmail.com"
export SMTP_PASSWORD="your-16-character-app-password"
export SMTP_USER_FRIENDLY_FROM="Христина Івасюк"  # Optional, defaults to "Христина Івасюк"
```

## Usage

**Important:** You must include `email.go` when running the script since it uses the EmailService code.

### Basic usage:
```bash
go run send_test_email.go email.go -to recipient@example.com
```

### With custom name:
```bash
go run send_test_email.go email.go -to recipient@example.com -name "John Doe"
```

### With Zoom link:
```bash
go run send_test_email.go email.go -to recipient@example.com -name "John Doe" -zoom "https://zoom.us/j/123456789"
```

### Show help:
```bash
go run send_test_email.go email.go -help
```

## What it does

The script will:
1. Initialize the email service using your SMTP configuration
2. Create a test booking for tomorrow at 2:00 PM in your local timezone
3. Send a booking confirmation email with:
   - Professional Ukrainian subject line
   - HTML formatted email body
   - Calendar invitation (.ics file) attachment
   - Google Calendar add link
   - Zoom meeting link (if provided)
   - Friendly "From" name: **Христина Івасюк**

## Testing the From Name

When you receive the test email, check that the sender displays as:
```
From: Христина Івасюк <your-email@gmail.com>
```

Instead of just:
```
From: your-email@gmail.com
```

## Troubleshooting

**Email not received:**
- Check your spam/junk folder
- Verify SMTP credentials are correct
- For Gmail, make sure you're using an App Password (not your regular password)
- Check the console output for error messages

**From name not showing:**
- Some email clients may take time to update the display name
- Try viewing in different email clients (web, mobile, desktop)
- The From name uses RFC 5322 format with base64 encoding for UTF-8 characters

## Example Output

```
========================================
Sending test confirmation email...
========================================
To: test@example.com
Name: John Doe
Slot Time: Thursday, January 3, 2026 at 2:00 PM EST
Zoom Link: https://zoom.us/j/123456789
========================================
Confirmation email sent successfully via SMTP to test@example.com
========================================
✅ Test email sent successfully!
========================================
Check your inbox at: test@example.com
Note: The email may take a few moments to arrive and might be in your spam folder.
```
