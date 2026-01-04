.PHONY: help build run test clean block-slots block-slots-dry block-slots-prod send-test-email docker-build docker-run

# Default API URL for local development
API_URL ?= http://localhost:8080

# Production API URL (override with: make block-slots-prod PROD_URL=https://your-app.com)
PROD_URL ?= https://sweip8cyfh.eu-central-1.awsapprunner.com

# Number of days to block slots ahead
DAYS ?= 30

# Test email recipient
TEST_EMAIL ?= sergii.olshanetskyi@gmail.com

help:
	@echo "Coach Calendar - Available Make Commands"
	@echo ""
	@echo "Building & Running:"
	@echo "  make build              - Build the main application"
	@echo "  make run                - Run the application locally"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Clean build artifacts"
	@echo ""
	@echo "Slot Blocking:"
	@echo "  make block-slots-dry    - Preview slots that would be blocked (30 days)"
	@echo "  make block-slots        - Block slots on local server (30 days)"
	@echo "  make block-slots-prod   - Block slots on production server"
	@echo ""
	@echo "  Custom examples:"
	@echo "    make block-slots-dry DAYS=60"
	@echo "    make block-slots DAYS=90"
	@echo "    make block-slots-prod PROD_URL=https://your-app.com DAYS=60"
	@echo ""
	@echo "Email Testing:"
	@echo "  make send-test-email    - Send test email to default recipient"
	@echo ""
	@echo "  Custom examples:"
	@echo "    make send-test-email TEST_EMAIL=someone@example.com"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Run application in Docker"
	@echo ""
	@echo "Environment Variables:"
	@echo "  API_URL      - API URL for blocking slots (default: http://localhost:8080)"
	@echo "  PROD_URL     - Production API URL"
	@echo "  DAYS         - Days ahead to block (default: 30)"
	@echo "  TEST_EMAIL   - Email recipient for testing (default: sergii.olshanetskyi@gmail.com)"

# Build the main application
build:
	@echo "Building coach-calendar..."
	@go build -o coach-calendar
	@echo "✅ Build complete: ./coach-calendar"

# Build the test email sender
build-email-script:
	@echo "Building send_test_email..."
	@go build -o send_test_email send_test_email.go email.go
	@echo "✅ Build complete: ./send_test_email"

# Run the application
run: build
	@echo "Checking for existing coach-calendar process..."
	@pkill -f ./coach-calendar || true
	@sleep 1
	@echo "Starting coach-calendar on port 8080..."
	@./coach-calendar

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f coach-calendar send_test_email
	@echo "✅ Clean complete"

# Block slots - dry run (preview only)
block-slots-dry:
	@echo "🔍 DRY RUN: Previewing slots to block for $(DAYS) days..."
	@go run block_slots.go -dry-run -days $(DAYS)

# Block slots on local development server
block-slots:
	@echo "🔒 Blocking slots on local server for $(DAYS) days..."
	@echo "API URL: $(API_URL)"
	@go run block_slots.go -api $(API_URL) -days $(DAYS)

# Block slots on production server
block-slots-prod:
	@echo "🔒 Blocking slots on production for $(DAYS) days..."
	@echo "Production URL: $(PROD_URL)"
	@echo ""
	@read -p "Are you sure you want to block slots on PRODUCTION? [y/N] " confirm && [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]
	@go run block_slots.go -api $(PROD_URL) -days $(DAYS)

# Send test email
send-test-email:
	@echo "📧 Sending test email to $(TEST_EMAIL)..."
	@go run send_test_email.go email.go -to $(TEST_EMAIL) -name "Test User"

# Send test email with Zoom link
send-test-email-zoom:
	@echo "📧 Sending test email with Zoom link to $(TEST_EMAIL)..."
	@go run send_test_email.go email.go -to $(TEST_EMAIL) -name "Test User" -zoom "https://zoom.us/j/test123456789"

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t coach-calendar:latest .
	@echo "✅ Docker image built: coach-calendar:latest"

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 \
		-e SMTP_HOST="${SMTP_HOST}" \
		-e SMTP_PORT="${SMTP_PORT}" \
		-e SMTP_FROM="${SMTP_FROM}" \
		-e SMTP_PASSWORD="${SMTP_PASSWORD}" \
		-e SMTP_USER_FRIENDLY_FROM="${SMTP_USER_FRIENDLY_FROM}" \
		-e ZOOM_ACCOUNT_ID="${ZOOM_ACCOUNT_ID}" \
		-e ZOOM_CLIENT_ID="${ZOOM_CLIENT_ID}" \
		-e ZOOM_CLIENT_SECRET="${ZOOM_CLIENT_SECRET}" \
		coach-calendar:latest

# Quick shortcuts
slots-dry: block-slots-dry
slots: block-slots
slots-prod: block-slots-prod
email: send-test-email
