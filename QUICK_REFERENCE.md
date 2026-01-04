# Quick Reference - Neon Database

## Connection Credentials

### Development Database
```
postgresql://neondb_owner:npg_Aho5nl0CPtHZ@ep-small-sound-agj17yfr-pooler.c-2.eu-central-1.aws.neon.tech/bookings?sslmode=require
```

### Production Database
```
postgresql://neondb_owner:npg_Aho5nl0CPtHZ@ep-quiet-darkness-ag9g8x6n-pooler.c-2.eu-central-1.aws.neon.tech/bookings?sslmode=require
```

## Quick Commands

### Connect in VS Code
1. Install SQLTools extension
2. Click database icon in sidebar
3. Select "Neon Production" or "Neon Development"
4. Enter password when prompted

### Connect via Terminal
```bash
# Using helper script
./connect_neon.sh

# Or directly
psql "postgresql://neondb_owner:npg_Aho5nl0CPtHZ@ep-quiet-darkness-ag9g8x6n-pooler.c-2.eu-central-1.aws.neon.tech/bookings?sslmode=require"
```

### Run Migration
```bash
./run_migration.sh
```

### Start Local Server
```bash
go run .
```

## Common SQL Queries

```sql
-- View all bookings
SELECT * FROM bookings ORDER BY slot_time DESC;

-- View blocked slots
SELECT * FROM blocked_slots ORDER BY slot_time DESC;

-- Count records
SELECT COUNT(*) FROM bookings;
SELECT COUNT(*) FROM blocked_slots;

-- Upcoming bookings
SELECT slot_time, name, email
FROM bookings
WHERE slot_time > NOW()
ORDER BY slot_time ASC;
```

## Environment Variables

### Development (.env)
```bash
DATABASE_URL='postgresql://...bookings?sslmode=require'
PORT=8080
SEND_CONFIRMATION_EMAIL=NO
CREATE_ZOOM_MEETING=NO
```

### Production (.env.production.example)
```bash
DATABASE_URL=postgresql://...bookings?sslmode=require
PORT=8080
USE_AWS_SES=true
SMTP_FROM=verified@yourdomain.com
```

## Important Files

| File | Purpose |
|------|---------|
| `migrate_from_production.go` | Migration script |
| `run_migration.sh` | Run migration helper |
| `connect_neon.sh` | Connect to Neon via psql |
| `.env` | Local development config |
| `.env.production.example` | Production config (not in git) |
| `.vscode/sqltools_connections.json` | VS Code DB connections |

## Git Status

```bash
# Check what's staged
git status

# View changes
git diff

# Add new files
git add .

# Commit
git commit -m "Your message"
```

## Troubleshooting

### SCRAM-SHA-256 Error
❌ Remove `&channel_binding=require` from DATABASE_URL
✅ Use only `?sslmode=require`

### Database Not Found
Create `bookings` database in Neon Console:
https://console.neon.tech

### Connection Failed
1. Check internet connection
2. Verify credentials
3. Check Neon status: https://neon.tech/status

## Documentation Links

- [DATABASE_SETUP.md](DATABASE_SETUP.md) - Database setup guide
- [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md) - Full migration docs
- [README_MIGRATION.md](README_MIGRATION.md) - Quick migration start
- [VSCODE_DATABASE_SETUP.md](VSCODE_DATABASE_SETUP.md) - VS Code setup
