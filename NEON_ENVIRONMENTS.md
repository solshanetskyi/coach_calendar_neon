# Managing Development and Production Environments with Neon

This guide explains how to set up separate development and production Neon databases for your Coach Calendar application.

## Overview

You'll have:
- **Development Database** - For local testing and development
- **Production Database** - For your deployed AWS application

Both databases will have identical schemas (automatically created by the application).

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Development                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Your Laptop                     Neon Cloud                     │
│  ┌──────────────┐               ┌──────────────┐               │
│  │              │  DATABASE_URL │              │               │
│  │  go run .    │──────────────▶│  Development │               │
│  │  (port 8080) │               │   Database   │               │
│  │              │               │              │               │
│  └──────────────┘               └──────────────┘               │
│  .env file with                  - Smaller compute             │
│  dev DATABASE_URL                - Auto-suspend enabled        │
│                                   - Free/Pro tier              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                         Production                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  AWS Cloud                       Neon Cloud                     │
│  ┌──────────────┐               ┌──────────────┐               │
│  │              │  DATABASE_URL │              │               │
│  │  App Runner  │──────────────▶│  Production  │               │
│  │  / ECS       │               │   Database   │               │
│  │  / Lambda    │               │              │               │
│  └──────────────┘               └──────────────┘               │
│  Environment vars                - Larger compute              │
│  or Secrets Manager              - Always on                   │
│                                   - Auto-scaling enabled        │
└─────────────────────────────────────────────────────────────────┘
```

## Setup Instructions

### Step 1: Create Two Neon Databases

#### Option A: Two Separate Projects (Recommended)

1. Log in to https://console.neon.tech
2. Create a **Development** project:
   - Click "New Project"
   - Name: "Coach Calendar - Development"
   - Region: Choose closest to you for faster local development
   - Copy the connection string

3. Create a **Production** project:
   - Click "New Project"
   - Name: "Coach Calendar - Production"
   - Region: Choose closest to your AWS deployment region
   - Copy the connection string

**Benefits:**
- Complete isolation between environments
- Independent scaling and monitoring
- Different compute settings per environment
- No risk of accidentally affecting production from dev

#### Option B: Branches in Same Project (Alternative)

Neon supports database branching, which is perfect for dev/prod separation:

1. Create a single project: "Coach Calendar"
2. The default branch (`main`) will be your production database
3. Create a development branch:
   - Go to "Branches" in Neon dashboard
   - Click "Create Branch"
   - Name: "development"
   - Branch from: "main"
   - Copy the connection string for the development branch

**Benefits:**
- Lower cost (single project)
- Easy to sync dev from production
- Branch-specific compute settings

### Step 2: Configure Local Development Environment

Create a `.env` file for local development:

```bash
# Copy the example file
cp .env.example .env
```

Edit `.env` and add your **development** database URL:

```bash
# Database Configuration
DATABASE_URL=postgresql://user:password@ep-dev-xxx.region.aws.neon.tech/neondb?sslmode=require

# Server Configuration
PORT=8080

# Email Configuration (optional for local dev)
# You can leave email unconfigured or use a test SMTP service
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_FROM=your-test-email@gmail.com
# SMTP_PASSWORD=your-app-password
```

**Important:** Add `.env` to `.gitignore` (already done) to prevent committing secrets.

### Step 3: Test Local Development Setup

```bash
# Run the application locally
go run .
```

The application will:
1. Connect to your development Neon database
2. Automatically create the necessary tables
3. Start the server on port 8080

Verify it works:
```bash
# Check health endpoint
curl http://localhost:8080/health

# Open in browser
open http://localhost:8080
```

### Step 4: Configure AWS Production Environment

For AWS deployments, set the **production** database URL as an environment variable.

#### AWS App Runner

Set environment variable in the App Runner console or via `apprunner.yaml`:

```yaml
version: 1.0
runtime: go1
build:
  commands:
    build:
      - go build -o coach-calendar .
run:
  command: ./coach-calendar
  env:
    - name: DATABASE_URL
      value: postgresql://user:password@ep-prod-xxx.region.aws.neon.tech/neondb?sslmode=require
    - name: USE_AWS_SES
      value: "true"
    - name: SMTP_FROM
      value: verified@yourdomain.com
    - name: AWS_REGION
      value: us-east-1
```

**Better approach - Use AWS Secrets Manager:**

```yaml
run:
  env:
    - name: DATABASE_URL
      value-from: arn:aws:secretsmanager:region:account:secret:coach-calendar/database-url
```

#### AWS Elastic Beanstalk

Set environment variables in `.ebextensions/environment.config`:

```yaml
option_settings:
  aws:elasticbeanstalk:application:environment:
    DATABASE_URL: postgresql://user:password@ep-prod-xxx.region.aws.neon.tech/neondb?sslmode=require
    USE_AWS_SES: "true"
    SMTP_FROM: verified@yourdomain.com
    AWS_REGION: us-east-1
```

Or use the Elastic Beanstalk console to set environment variables.

#### AWS ECS/Fargate

Add to your task definition JSON:

```json
{
  "containerDefinitions": [
    {
      "environment": [
        {
          "name": "DATABASE_URL",
          "value": "postgresql://user:password@ep-prod-xxx.region.aws.neon.tech/neondb?sslmode=require"
        },
        {
          "name": "USE_AWS_SES",
          "value": "true"
        }
      ]
    }
  ]
}
```

**Better - Use secrets:**

```json
{
  "containerDefinitions": [
    {
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:coach-calendar/database-url"
        }
      ]
    }
  ]
}
```

#### AWS Lambda

Set environment variables in the Lambda console or via SAM/Terraform:

```yaml
# SAM template
Environment:
  Variables:
    DATABASE_URL: postgresql://user:password@ep-prod-xxx.region.aws.neon.tech/neondb?sslmode=require
```

### Step 5: Secure Your Connection Strings

**Never commit database URLs to git!** Use these approaches:

#### Option 1: AWS Secrets Manager (Recommended for Production)

1. Store the production database URL in AWS Secrets Manager:
```bash
aws secretsmanager create-secret \
  --name coach-calendar/database-url \
  --secret-string "postgresql://user:password@ep-prod-xxx.region.aws.neon.tech/neondb?sslmode=require" \
  --region us-east-1
```

2. Update your application to read from Secrets Manager (optional, or use AWS service integration):

```go
// Add to database.go if you want to fetch from Secrets Manager
import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/secretsmanager"
)

func getDatabaseURL() (string, error) {
    // Try environment variable first
    if url := os.Getenv("DATABASE_URL"); url != "" {
        return url, nil
    }

    // Fallback to Secrets Manager
    sess := session.Must(session.NewSession())
    svc := secretsmanager.New(sess)

    result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
        SecretId: aws.String("coach-calendar/database-url"),
    })
    if err != nil {
        return "", err
    }

    return *result.SecretString, nil
}
```

#### Option 2: Environment Variables (Simpler)

Keep using environment variables but ensure:
- `.env` is in `.gitignore` ✅ (already done)
- Production values are set in AWS service configuration
- Team members have their own `.env` files (not committed)

## Environment-Specific Configuration

### Development Environment

Characteristics:
- More verbose logging
- Relaxed email requirements (optional)
- Smaller Neon compute size (saves cost)
- Can use Neon's auto-suspend feature

```bash
# .env (local development)
DATABASE_URL=postgresql://...dev...
PORT=8080
# Email optional for testing
```

### Production Environment

Characteristics:
- Production-grade logging
- Email required (AWS SES)
- Larger Neon compute size (better performance)
- Disable auto-suspend for always-on availability

```bash
# AWS Environment Variables
DATABASE_URL=postgresql://...prod...
PORT=8080
USE_AWS_SES=true
SMTP_FROM=verified@yourdomain.com
AWS_REGION=us-east-1
```

## Schema Management

Both databases will have identical schemas because:
1. The application automatically creates tables on startup
2. Both use the same code from `database.go`

### Initial Setup

When you first run the application in each environment:

**Development:**
```bash
# Local machine
export DATABASE_URL="postgresql://...dev..."
go run .
# Tables created automatically
```

**Production:**
```bash
# Deploy to AWS
# App Runner/ECS/Lambda starts the application
# Tables created automatically on first run
```

### Schema Migrations

If you modify the database schema:

1. **Update `database.go`** with new table definitions
2. **Test locally** with development database
3. **Deploy to production** - new tables/columns will be created automatically

**Important:** The current setup only supports additive changes (new tables). For destructive changes (dropping columns), you'll need to:
- Manually run SQL in Neon console, or
- Implement a proper migration system (see below)

## Advanced: Proper Migration System

For production applications, consider using a migration tool:

### Option 1: golang-migrate

```bash
go get -u github.com/golang-migrate/migrate/v4
```

Create migrations:
```bash
migrate create -ext sql -dir migrations -seq create_bookings_table
```

This creates separate dev/prod workflows with version control.

### Option 2: Goose

```bash
go get -u github.com/pressly/goose/v3/cmd/goose
```

More Go-friendly migration approach.

## Syncing Development from Production

### Using Neon Branching (If using Option B)

```bash
# Create a fresh dev branch from production
# This copies all production data to development
```

In Neon console:
1. Delete old development branch
2. Create new development branch from main
3. Update your local `.env` with new development connection string

### Manual Data Sync

Export from production and import to development:

```bash
# 1. Set production DATABASE_URL temporarily
export DATABASE_URL="postgresql://...prod..."
go run import_to_postgres.go

# 2. Export to JSON
# (Modify export script to work with Postgres instead of SQLite)

# 3. Set development DATABASE_URL
export DATABASE_URL="postgresql://...dev..."
go run import_to_postgres.go
```

## Monitoring Both Environments

### Neon Dashboard

Monitor each database:
- **Metrics:** Query performance, connection count, storage usage
- **Queries:** See active and slow queries
- **Branches:** View all branches and their status
- **Compute:** Adjust compute resources per environment

### Application Logging

The application logs database initialization:
```
2026-01-04 12:00:00 Successfully connected to Neon PostgreSQL database
2026-01-04 12:00:00 Database tables initialized successfully
```

Check logs to verify correct database connection.

## Cost Optimization

### Development Database
- Use smaller compute size (0.25 vCPU)
- Enable auto-suspend after 5 minutes of inactivity
- Use Neon's free tier if within limits

### Production Database
- Right-size compute based on load (start with 0.5-1 vCPU)
- Disable auto-suspend for always-on availability
- Enable auto-scaling if traffic varies

### Neon Pricing Tiers
- **Free Tier:** Good for development, 0.25 vCPU, 10 projects
- **Pro:** Better for production, auto-scaling, more compute options
- **Enterprise:** High availability, dedicated support

## Troubleshooting

### Wrong Database Connected

Check which database you're connected to:
```bash
# Check environment variable
echo $DATABASE_URL

# Or check application logs
# Look for "Successfully connected to Neon PostgreSQL database"
```

### Tables Not Created

If tables don't exist:
1. Check application logs for errors
2. Verify DATABASE_URL is correct
3. Ensure Neon database is running (not suspended)
4. Check IAM permissions (if using Secrets Manager)

### Schema Differences

If dev and prod schemas differ:
1. Check `database.go` for recent changes
2. Manually sync schemas using SQL in Neon console
3. Consider implementing proper migrations

## Best Practices

✅ **Never commit `.env` files** - Already in `.gitignore`
✅ **Use separate projects/branches** - Complete environment isolation
✅ **Use AWS Secrets Manager in production** - Secure credential storage
✅ **Test in development first** - Always test changes locally
✅ **Monitor both environments** - Use Neon dashboard and application logs
✅ **Regular backups** - Neon provides automatic point-in-time recovery
✅ **Document your setup** - Keep team informed of environment configs

## Quick Reference

### Local Development
```bash
# .env file
DATABASE_URL=postgresql://...dev...

# Run locally
go run .
```

### Production Deployment
```bash
# AWS environment variable
DATABASE_URL=postgresql://...prod...

# Deploy
git push origin main
# (triggers AWS deployment)
```

### Check Current Environment
```bash
# See which database you're using
echo $DATABASE_URL | grep -o "ep-[^.]*"
# Shows: ep-dev-xxx or ep-prod-xxx
```

## Summary

This setup gives you:
- ✅ Isolated development and production databases
- ✅ Identical schemas in both environments
- ✅ Secure credential management
- ✅ Easy local development workflow
- ✅ Simple production deployment
- ✅ Cost-optimized infrastructure
