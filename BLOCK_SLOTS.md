# Block Slots Script

This script automatically blocks time slots from Monday to Thursday, 11:30-15:30 Amsterdam time using the booking API.

## Purpose

Use this script to block out recurring unavailable time periods, such as:
- Lunch breaks
- Regular commitments
- Buffer time between appointments
- Any recurring blocked periods

## Usage

### Using Make (Recommended)

The easiest way to use the script is via the Makefile:

```bash
# Preview what would be blocked (dry run)
make block-slots-dry

# Block slots on local development server
make block-slots

# Block slots on production
make block-slots-prod

# Custom number of days
make block-slots-dry DAYS=60
make block-slots DAYS=90

# Production with custom URL and days
make block-slots-prod PROD_URL=https://your-app.com DAYS=60
```

### Using Go Directly

You can also run the script directly:

```bash
# Basic usage (local development)
go run block_slots.go

# Block slots on production
go run block_slots.go -api https://your-app-runner-url.amazonaws.com

# Dry run (preview what would be blocked)
go run block_slots.go -dry-run

# Block slots for 60 days ahead
go run block_slots.go -days 60

# Show help
go run block_slots.go -help
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `-api` | API base URL | `http://localhost:8080` |
| `-days` | Number of days ahead to block | `30` |
| `-dry-run` | Preview mode (doesn't actually block) | `false` |
| `-help` | Show help message | - |

## What it Does

The script will:
1. Calculate the next N days (default: 30)
2. Find all Monday-Thursday dates
3. Generate 30-minute slots from 11:30 to 15:30 Amsterdam time
4. Block each slot via the API endpoint `/api/admin/block`

### Time Slots Blocked Per Day:
- 11:30
- 12:00
- 12:30
- 13:00
- 13:30
- 14:00
- 14:30
- 15:00
- 15:30

**Total: 9 slots per day × 4 days per week**

## Examples

### Preview slots for the next 30 days:
```bash
go run block_slots.go -dry-run
```

Output:
```
========================================
Blocking slots for 30 days ahead
Time range: Monday-Thursday, 11:30-15:30 Amsterdam time
Total slots to block: 108
DRY RUN MODE - Not actually blocking slots
========================================
[1/108] Would block: Mon, Jan 6, 2026 11:30 CET (Monday)
[2/108] Would block: Mon, Jan 6, 2026 12:00 CET (Monday)
...
```

### Block slots on local dev server:
```bash
# Make sure your server is running on port 8080
go run block_slots.go
```

### Block slots on production:
```bash
go run block_slots.go -api https://abcd1234.us-east-1.awsapprunner.com -days 90
```

### Block slots for the entire quarter (90 days):
```bash
go run block_slots.go -api https://your-app.com -days 90
```

## Important Notes

1. **Timezone Handling**: All slots are in Amsterdam (Europe/Amsterdam) timezone, which automatically handles:
   - CET (Central European Time) in winter
   - CEST (Central European Summer Time) in summer

2. **API Endpoint Required**: Your application must have the `/api/admin/block` endpoint implemented.

3. **Idempotent**: Blocking the same slot multiple times is safe (won't cause errors).

4. **Rate Limiting**: The script adds a 100ms delay between requests to avoid overwhelming the API.

## Scheduling with Cron

To automatically block slots weekly, add to your crontab:

```bash
# Run every Sunday at 11 PM to block next week's slots
0 23 * * 0 cd /path/to/coach_calendar && go run block_slots.go -api https://your-app.com
```

Or create a systemd timer, AWS CloudWatch Event, or similar scheduling service.

## Troubleshooting

### "Failed to load Amsterdam timezone"
The system doesn't have timezone data. Install it:
```bash
# Ubuntu/Debian
sudo apt-get install tzdata

# macOS (usually pre-installed)
# If missing, reinstall Xcode Command Line Tools
```

### "API returned status 404"
The `/api/admin/block` endpoint doesn't exist. Verify your API is running and has this endpoint.

### "Connection refused"
The API server is not running or not accessible at the specified URL.

## Advanced Usage

### Customize Time Range

To block different hours, edit the script and modify:
```go
// Line ~55-57
startHour := 11
startMinute := 30
endHour := 15
endMinute := 30
```

### Customize Days of Week

To block different days, edit the script and modify:
```go
// Line ~52
if weekday >= time.Monday && weekday <= time.Thursday {
```

Change to, for example:
- All weekdays: `if weekday >= time.Monday && weekday <= time.Friday {`
- Only Wednesday: `if weekday == time.Wednesday {`
- Weekends: `if weekday == time.Saturday || weekday == time.Sunday {`
