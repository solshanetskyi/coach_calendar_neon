# Coach Calendar

A simple meeting booking application built with Go.

## Project Structure

```
.
├── main.go              # Application entry point and routing
├── database.go          # Database initialization and slot generation
├── email.go             # Email service for booking confirmations
├── handlers/
│   ├── api.go          # API handlers for bookings and admin operations
│   └── pages.go        # Page handlers (home, admin, health)
├── go.mod
└── bookings.db         # SQLite database (generated at runtime)
```

## Features

- **Public Booking Interface** (`/`)
  - Calendar view showing only January slots
  - Timezone-aware display (shows times in user's browser timezone)
  - 30-minute time slot selection (9 AM - 5 PM)
  - Booking form with name and email
  - Automatic email confirmation upon booking

- **Admin Panel** (`/admin`)
  - View all slots (available, booked, blocked)
  - Block/unblock time slots
  - Click on booked slots to view booking details in a modal
  - Filter slots by status
  - Timezone-aware display

## API Endpoints

### Public API
- `GET /api/slots` - Get available time slots
- `POST /api/bookings` - Create a new booking

### Admin API
- `GET /api/admin/slots` - Get all slots with status and booking info
- `POST /api/admin/block` - Block a time slot
- `POST /api/admin/unblock` - Unblock a time slot

## Running the Application

### Basic Setup

```bash
# Build
go build -o coach-calendar

# Run
./coach-calendar

# Or directly with go run
go run .
```

The server will start on port 8080 by default (configurable via `PORT` environment variable).

### Email Configuration (Optional)

The application supports two methods for sending emails:
1. **AWS SES** (Recommended for AWS-hosted applications)
2. **SMTP** (Gmail, Outlook, etc.)

#### Option 1: AWS SES (Recommended for AWS)

When hosting on AWS, use AWS SES for reliable, scalable email delivery:

```bash
export USE_AWS_SES="true"                   # Enable AWS SES
export SMTP_FROM="verified@yourdomain.com"  # Verified sender email in SES
export AWS_REGION="us-east-1"               # AWS region (optional, defaults to us-east-1)

# Then run the application
./coach-calendar
```

**AWS SES Setup Requirements:**
1. **Verify your sender email** in AWS SES Console
2. **IAM Permissions**: Ensure your EC2 instance/ECS task has an IAM role with SES permissions:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "ses:SendEmail",
           "ses:SendRawEmail"
         ],
         "Resource": "*"
       }
     ]
   }
   ```
3. **Move out of SES Sandbox** (if needed): By default, SES is in sandbox mode and can only send to verified emails. Request production access to send to any email.

**AWS Regions with SES:**
- `us-east-1` (N. Virginia)
- `us-west-2` (Oregon)
- `eu-west-1` (Ireland)
- `ap-southeast-1` (Singapore)
- And more...

#### Option 2: SMTP

For non-AWS hosting or testing locally:

```bash
export SMTP_HOST="smtp.gmail.com"           # Your SMTP server
export SMTP_PORT="587"                      # SMTP port (usually 587 for TLS)
export SMTP_FROM="your-email@gmail.com"     # Sender email address
export SMTP_PASSWORD="your-app-password"    # SMTP password or app-specific password

# Then run the application
./coach-calendar
```

**Note:** If email is not configured, the application will run normally but skip sending confirmation emails. A warning will be logged on startup.

#### Gmail Setup Example

If using Gmail, you'll need to:
1. Enable 2-factor authentication on your Google account
2. Generate an "App Password" (Google Account → Security → App Passwords)
3. Use the app password as `SMTP_PASSWORD`

```bash
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_FROM="yourname@gmail.com"
export SMTP_PASSWORD="your-16-char-app-password"
./coach-calendar
```

#### Other SMTP Providers

- **Outlook/Office365**: `smtp.office365.com:587`
- **Yahoo**: `smtp.mail.yahoo.com:587`
- **SendGrid**: `smtp.sendgrid.net:587`
- **Mailgun**: `smtp.mailgun.org:587`

### Deployment on AWS

When deploying to AWS (EC2, ECS, Lambda, etc.):

1. **Use AWS SES** for email (recommended):
   ```bash
   USE_AWS_SES=true
   SMTP_FROM=verified@yourdomain.com
   AWS_REGION=us-east-1
   ```

2. **Attach IAM Role** with SES permissions to your compute resource

3. **Environment Variables**: Set via:
   - EC2: User data script or `/etc/environment`
   - ECS: Task definition environment variables
   - Lambda: Function configuration
   - Elastic Beanstalk: Configuration in `.ebextensions` or console

4. **AWS Credentials**: No need to configure `AWS_ACCESS_KEY_ID` or `AWS_SECRET_ACCESS_KEY` when using IAM roles (recommended approach)

## Database

The application uses SQLite (pure Go implementation via `modernc.org/sqlite`) with two main tables:
- `bookings` - Stores booking information (slot_time, name, email, created_at, duration)
- `blocked_slots` - Stores administratively blocked time slots

**Note:** This version uses a pure Go SQLite driver that doesn't require CGO or C compiler, making it easy to deploy on any platform including AWS App Runner.

## Email Confirmations

When a booking is created, the application automatically sends a confirmation email to the customer with:
- Appointment date and time
- Duration (30 minutes)
- Customer name and email
- Professional formatting

The email service is fault-tolerant - if email sending fails, the booking is still created successfully, and an error is logged.

## Development

The codebase is organized into:
- **main.go** - Minimal entry point with routing
- **database.go** - Database logic and slot generation
- **handlers/** - HTTP request handlers split by concern:
  - API handlers for JSON endpoints
  - Page handlers for HTML pages
