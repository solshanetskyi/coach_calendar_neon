# Environment Setup - Quick Reference

## TL;DR - Getting Started

### Development (Local Machine)
```bash
# 1. Quick setup
./setup-dev-env.sh

# 2. Edit .env and add your Neon dev database URL
nano .env

# 3. Run
go run .
```

### Production (AWS)
Set `DATABASE_URL` environment variable in your AWS service (App Runner, ECS, Lambda, etc.) with your Neon production database URL.

---

## Complete Workflow

### Initial Setup

#### 1. Create Neon Databases

**Option A: Two Projects (Recommended)**
- Go to https://console.neon.tech
- Create "Coach Calendar - Development" project → Copy URL
- Create "Coach Calendar - Production" project → Copy URL

**Option B: Use Branches**
- Create one project "Coach Calendar"
- Main branch = Production
- Create "development" branch → Copy URL

#### 2. Local Development Setup

```bash
# Quick setup
./setup-dev-env.sh

# Or manual
cp .env.development.example .env

# Edit .env
nano .env
# Add: DATABASE_URL=postgresql://...dev-url...

# Install and run
go mod tidy
go run .

# Visit
open http://localhost:8080
```

#### 3. Production Setup

Choose your AWS service:

**App Runner:**
```yaml
# apprunner.yaml
run:
  env:
    - name: DATABASE_URL
      value: postgresql://...prod-url...
```

**ECS:**
```json
{
  "environment": [
    {
      "name": "DATABASE_URL",
      "value": "postgresql://...prod-url..."
    }
  ]
}
```

**Lambda:**
```yaml
Environment:
  Variables:
    DATABASE_URL: postgresql://...prod-url...
```

---

## File Structure

```
Your Project
│
├── .env                          # Local dev config (git ignored)
├── .env.development.example      # Dev template (committed)
├── .env.production.example       # Prod reference (committed)
│
└── Documentation
    ├── NEON_ENVIRONMENTS.md      # Full dev/prod guide
    ├── NEON_MIGRATION.md         # SQLite → Neon migration
    └── README.md                 # Main documentation
```

---

## Environment Variables Reference

### Required
| Variable | Dev | Prod | Description |
|----------|-----|------|-------------|
| `DATABASE_URL` | ✅ | ✅ | Neon PostgreSQL connection string |

### Optional
| Variable | Dev | Prod | Description |
|----------|-----|------|-------------|
| `PORT` | 8080 | 8080 | Server port |
| `USE_AWS_SES` | ❌ | ✅ | Use AWS SES for email |
| `SMTP_FROM` | 📧 | ✅ | Sender email address |
| `SMTP_HOST` | 📧 | ❌ | SMTP server (if not using SES) |
| `SMTP_PORT` | 📧 | ❌ | SMTP port (if not using SES) |
| `SMTP_PASSWORD` | 📧 | ❌ | SMTP password (if not using SES) |
| `AWS_REGION` | ❌ | ✅ | AWS region for SES |

Legend: ✅ Recommended | ❌ Not needed | 📧 Optional for testing

---

## Neon Database Configuration

### Development Database
- **Compute:** 0.25 vCPU (smallest)
- **Auto-suspend:** After 5 minutes (saves cost)
- **Tier:** Free tier is perfect
- **Region:** Closest to your location

### Production Database
- **Compute:** 0.5-1 vCPU (adjust based on load)
- **Auto-suspend:** Disabled (always on)
- **Tier:** Pro or higher
- **Region:** Same as your AWS deployment
- **Auto-scaling:** Enabled (optional, for variable load)

---

## Common Tasks

### Switch Environments Locally

```bash
# Use development
export DATABASE_URL="postgresql://...dev-url..."
go run .

# Use production (read-only testing)
export DATABASE_URL="postgresql://...prod-url..."
go run .
```

### Check Current Environment

```bash
# See which database you're connected to
echo $DATABASE_URL | grep -o "ep-[^.]*"
# Shows: ep-dev-xxx or ep-prod-xxx
```

### Sync Dev Database from Production

Using Neon branches (if using Option B):
1. Delete development branch in Neon console
2. Create new development branch from main
3. Update DATABASE_URL in .env

---

## Security Best Practices

### ✅ Do
- Keep `.env` in `.gitignore` (already done)
- Use AWS Secrets Manager for production
- Use separate databases for dev/prod
- Use environment-specific connection strings
- Rotate database passwords regularly

### ❌ Don't
- Commit `.env` files to git
- Use production database for development
- Hardcode credentials in code
- Share database URLs in chat/email
- Use the same database for all environments

---

## Troubleshooting

### "DATABASE_URL environment variable is not set"
**Solution:** Create `.env` file with `DATABASE_URL=...`

### "Failed to ping database"
**Solution:**
- Check DATABASE_URL is correct
- Verify Neon database is running (not suspended)
- Check your internet connection

### Tables not created
**Solution:**
- Check application logs for errors
- Ensure DATABASE_URL is correct
- Verify Neon database exists

### Wrong database connected
**Solution:**
- Check `echo $DATABASE_URL`
- Look for "ep-dev" or "ep-prod" in the URL
- Verify `.env` file has correct URL

---

## Cost Optimization

### Development
- ✅ Use Neon free tier (0.25 vCPU, 10 projects)
- ✅ Enable auto-suspend (saves compute time)
- ✅ Use smaller compute size
- ✅ Share dev project with team

### Production
- ✅ Right-size compute based on actual load
- ✅ Use auto-scaling if traffic varies
- ✅ Monitor usage in Neon dashboard
- ✅ Consider Neon Pro for better pricing

**Estimated Costs:**
- Development: $0/month (free tier)
- Production: $19/month (Pro plan) + compute usage

---

## Quick Commands

```bash
# Setup development environment
./setup-dev-env.sh

# Install dependencies
go mod tidy

# Run locally
go run .

# Build for production
go build -o coach-calendar

# Check environment
echo $DATABASE_URL

# Test database connection
go run . &
curl http://localhost:8080/health
```

---

## Getting Help

- **Full Environment Guide:** [NEON_ENVIRONMENTS.md](NEON_ENVIRONMENTS.md)
- **Migration Guide:** [NEON_MIGRATION.md](NEON_MIGRATION.md)
- **Application Docs:** [README.md](README.md)
- **Neon Docs:** https://neon.tech/docs
- **Neon Support:** https://discord.gg/92vNTzKDGp
