# Makefile Commands Reference

This project includes a Makefile to simplify common tasks. Run `make help` to see all available commands.

## Quick Start

```bash
# See all available commands
make help

# Build and run the application
make build
make run

# Preview slots that would be blocked (safe to run)
make block-slots-dry

# Send a test email
make send-test-email
```

## Building & Running

### Build the Application
```bash
make build
```
Creates the `coach-calendar` executable in the current directory.

### Run the Application
```bash
make run
```
Builds and starts the server on port 8080.

### Clean Build Artifacts
```bash
make clean
```
Removes compiled binaries (`coach-calendar`, `send_test_email`).

### Run Tests
```bash
make test
```
Runs all Go tests with verbose output.

## Blocking Time Slots

### Dry Run (Preview Mode)
```bash
make block-slots-dry
```
Shows what slots would be blocked **without actually blocking them**. Safe to run anytime.

**Custom number of days:**
```bash
make block-slots-dry DAYS=60
make block-slots-dry DAYS=90
```

### Block Slots Locally
```bash
make block-slots
```
Blocks slots on your local development server (http://localhost:8080).

**Custom settings:**
```bash
# Block 90 days ahead
make block-slots DAYS=90

# Use different local URL
make block-slots API_URL=http://localhost:3000 DAYS=60
```

### Block Slots on Production
```bash
make block-slots-prod
```
Blocks slots on production server. **Includes a confirmation prompt** for safety.

**Custom production URL:**
```bash
make block-slots-prod PROD_URL=https://your-app.awsapprunner.com
make block-slots-prod PROD_URL=https://your-app.com DAYS=60
```

### What Gets Blocked

- **Days:** Monday, Tuesday, Wednesday, Thursday
- **Time:** 11:30 AM - 3:30 PM Amsterdam time (CET/CEST)
- **Interval:** Every 30 minutes
- **Total:** 9 slots per day × 4 days/week

## Email Testing

### Send Test Email
```bash
make send-test-email
```
Sends a test booking confirmation email to the default recipient (sergii.olshanetskyi@gmail.com).

**Custom recipient:**
```bash
make send-test-email TEST_EMAIL=someone@example.com
```

### Send Test Email with Zoom Link
```bash
make send-test-email-zoom
```
Sends a test email including a Zoom meeting link section.

**Custom recipient:**
```bash
make send-test-email-zoom TEST_EMAIL=someone@example.com
```

### Requirements

Before sending test emails, set these environment variables:
```bash
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_FROM="your-email@gmail.com"
export SMTP_PASSWORD="your-16-char-app-password"
export SMTP_USER_FRIENDLY_FROM="Христина Івасюк"  # Optional
```

## Docker Commands

### Build Docker Image
```bash
make docker-build
```
Creates a Docker image tagged as `coach-calendar:latest`.

### Run in Docker
```bash
make docker-run
```
Runs the application in a Docker container, automatically passing environment variables.

**Note:** Make sure to set environment variables before running:
```bash
export SMTP_HOST="smtp.gmail.com"
export SMTP_PORT="587"
export SMTP_FROM="your-email@gmail.com"
export SMTP_PASSWORD="your-app-password"
export ZOOM_ACCOUNT_ID="your-zoom-account-id"
export ZOOM_CLIENT_ID="your-zoom-client-id"
export ZOOM_CLIENT_SECRET="your-zoom-client-secret"

make docker-run
```

## Environment Variables

| Variable | Description | Used By |
|----------|-------------|---------|
| `API_URL` | Local API URL | Slot blocking (local) |
| `PROD_URL` | Production API URL | Slot blocking (prod) |
| `DAYS` | Days ahead to block slots | Slot blocking |
| `TEST_EMAIL` | Email recipient for testing | Email testing |
| `SMTP_HOST` | SMTP server hostname | Email, Docker |
| `SMTP_PORT` | SMTP server port | Email, Docker |
| `SMTP_FROM` | From email address | Email, Docker |
| `SMTP_PASSWORD` | SMTP password | Email, Docker |
| `SMTP_USER_FRIENDLY_FROM` | Display name for From field | Email, Docker |
| `ZOOM_ACCOUNT_ID` | Zoom account ID | Docker |
| `ZOOM_CLIENT_ID` | Zoom client ID | Docker |
| `ZOOM_CLIENT_SECRET` | Zoom client secret | Docker |

## Quick Shortcuts

The Makefile includes short aliases for common commands:

```bash
make slots-dry    # Same as: make block-slots-dry
make slots        # Same as: make block-slots
make slots-prod   # Same as: make block-slots-prod
make email        # Same as: make send-test-email
```

## Examples

### Daily Development Workflow
```bash
# Start fresh
make clean
make build
make run
```

### Test Email Configuration
```bash
# Preview what email looks like
make send-test-email TEST_EMAIL=your-email@example.com
```

### Weekly Slot Blocking
```bash
# Preview first
make block-slots-dry DAYS=30

# If looks good, block on production
make block-slots-prod PROD_URL=https://your-app.com DAYS=30
```

### Deploy New Version
```bash
# Build and test locally
make clean
make build
make test

# Build Docker image
make docker-build

# Test in Docker
make docker-run
```

## Tips

1. **Always dry-run first**: Use `make block-slots-dry` before blocking slots to preview what will happen.

2. **Production safety**: The `make block-slots-prod` command includes a confirmation prompt to prevent accidents.

3. **Environment variables**: Set commonly used variables in your shell profile:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PROD_URL="https://your-app-runner-url.amazonaws.com"
   export TEST_EMAIL="your-email@example.com"
   ```

4. **Quick help**: Run `make` or `make help` anytime to see available commands.

## Troubleshooting

### "make: command not found"
Install Make:
```bash
# macOS (usually pre-installed)
xcode-select --install

# Ubuntu/Debian
sudo apt-get install build-essential

# Windows
# Use WSL or install GNU Make for Windows
```

### "No rule to make target"
Check that you're in the project directory and the Makefile exists:
```bash
ls -la Makefile
```

### Permission denied when running binaries
Make the binary executable:
```bash
chmod +x coach-calendar
```
