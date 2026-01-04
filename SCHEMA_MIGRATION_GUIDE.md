# SQLite to Neon Schema Migration Guide

## Understanding Schema Migration

**Good News:** The application automatically creates the Neon schema for you! The same code in `database.go` works for both SQLite and PostgreSQL.

## How It Works

When you run the application with a Neon `DATABASE_URL`, the `initDB()` function in [database.go](database.go) automatically:

1. Connects to Neon PostgreSQL
2. Creates the `bookings` table (if it doesn't exist)
3. Creates the `blocked_slots` table (if it doesn't exist)
4. Creates indexes for performance

**No manual schema creation needed!**

## Schema Comparison

### SQLite (Old)
```sql
CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slot_time DATETIME NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    duration INTEGER NOT NULL DEFAULT 30
);
CREATE INDEX IF NOT EXISTS idx_slot_time ON bookings(slot_time);

CREATE TABLE IF NOT EXISTS blocked_slots (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slot_time DATETIME NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_blocked_slot_time ON blocked_slots(slot_time);
```

### PostgreSQL/Neon (New)
```sql
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,                           -- Changed from AUTOINCREMENT
    slot_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE,  -- Changed from DATETIME
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,  -- Changed from DATETIME
    duration INTEGER NOT NULL DEFAULT 30
);
CREATE INDEX IF NOT EXISTS idx_slot_time ON bookings(slot_time);

CREATE TABLE IF NOT EXISTS blocked_slots (
    id SERIAL PRIMARY KEY,                           -- Changed from AUTOINCREMENT
    slot_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE,  -- Changed from DATETIME
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP  -- Changed from DATETIME
);
CREATE INDEX IF NOT EXISTS idx_blocked_slot_time ON blocked_slots(slot_time);
```

### Key Differences
| Feature | SQLite | PostgreSQL/Neon |
|---------|--------|-----------------|
| Auto-increment | `INTEGER PRIMARY KEY AUTOINCREMENT` | `SERIAL PRIMARY KEY` |
| Timestamps | `DATETIME` | `TIMESTAMP WITH TIME ZONE` |
| Timezone Support | Limited | Full support |
| Concurrent Writes | File-based locking | Row-level locking |
| Performance | Good for small scale | Excellent for any scale |

## Migration Methods

### Method 1: Fresh Start (Schema Only)

If you don't need to preserve existing data:

```bash
# 1. Set your Neon DATABASE_URL
export DATABASE_URL="postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require"

# 2. Run the application
go run .

# ✅ Schema is automatically created in Neon!
```

**When to use:**
- New deployment
- Testing/development environment
- No existing data to preserve

### Method 2: Schema + Data Migration (Recommended)

If you have existing SQLite data to preserve:

```bash
# 1. Export data from SQLite
go run export_sqlite_data.go

# Output:
# ✅ Exported X bookings to bookings_export.json
# ✅ Exported Y blocked slots to blocked_slots_export.json

# 2. Set your Neon DATABASE_URL
export DATABASE_URL="postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require"

# 3. Run application to create schema
go run . &
sleep 3  # Give it time to create tables
pkill -f "go run"

# 4. Import the data
go run import_to_postgres.go

# Output:
# ✅ Connected to Neon PostgreSQL database
# ✅ Imported X bookings (skipped Y duplicates)
# ✅ Imported Z blocked slots (skipped W duplicates)
```

**When to use:**
- Migrating from existing SQLite deployment
- Preserving production data
- Moving dev database to production

### Method 3: Inspect Then Create (Manual)

If you want to see and verify the schema first:

```bash
# 1. Inspect current SQLite schema
./export_sqlite_schema.sh

# Shows:
# - Table structures
# - Index definitions
# - Row counts
# - Sample data

# 2. Connect to Neon via psql to verify (optional)
psql "postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require"

# 3. Run application to create schema
export DATABASE_URL="postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require"
go run .

# 4. Verify schema in Neon
psql "postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require" \
  -c "\d bookings" \
  -c "\d blocked_slots"
```

**When to use:**
- Want to verify schema before running
- Learning/educational purposes
- Debugging schema issues

## Step-by-Step: Complete Migration

### Step 1: Backup Your SQLite Database

```bash
# Create a backup of your current database
cp bookings.db bookings.db.backup

# Verify backup exists
ls -lh bookings.db*
```

### Step 2: Inspect Current Data

```bash
# See what you have
./export_sqlite_schema.sh

# This shows:
# - Number of bookings
# - Number of blocked slots
# - Sample data from each table
```

### Step 3: Set Up Neon Database

```bash
# 1. Create Neon project at https://console.neon.tech
# 2. Copy your connection string
# 3. Set it as environment variable

export DATABASE_URL="postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require"

# Or add to .env file
echo 'DATABASE_URL=postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require' >> .env
```

### Step 4: Export SQLite Data

```bash
# Export to JSON files
go run export_sqlite_data.go

# Verify exports
ls -lh *_export.json
cat bookings_export.json | jq length  # Count bookings (if you have jq installed)
```

### Step 5: Create Schema in Neon

```bash
# Run application - it creates the schema automatically
go run . &

# Wait a few seconds for startup
sleep 5

# Check logs - you should see:
# "Successfully connected to Neon PostgreSQL database"
# "Database tables initialized successfully"

# Stop the app
pkill -f "go run"
```

### Step 6: Import Data to Neon

```bash
# Import the exported data
go run import_to_postgres.go

# You should see:
# ✅ Connected to Neon PostgreSQL database
# Importing X bookings...
# ✅ Imported X bookings (skipped 0 duplicates)
# Importing Y blocked slots...
# ✅ Imported Y blocked slots (skipped 0 duplicates)
# ✅ Import completed successfully!
```

### Step 7: Verify Migration

```bash
# Start the application
go run .

# In another terminal, test the endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/slots

# Or visit in browser
open http://localhost:8080
open http://localhost:8080/admin
```

### Step 8: Verify Data in Neon (Optional)

Using Neon console or psql:

```bash
# Connect to your Neon database
psql "postgresql://user:password@ep-xxx.region.aws.neon.tech/neondb?sslmode=require"

# Check tables exist
\dt

# Count rows
SELECT COUNT(*) FROM bookings;
SELECT COUNT(*) FROM blocked_slots;

# View sample data
SELECT * FROM bookings ORDER BY created_at DESC LIMIT 5;
SELECT * FROM blocked_slots ORDER BY created_at DESC LIMIT 5;

# Exit
\q
```

## Common Issues and Solutions

### Issue: "bookings.db not found"

**Solution:**
```bash
# Check current directory
pwd

# List database files
ls -la *.db

# If no database exists, skip export step
# The schema will still be created automatically
```

### Issue: "DATABASE_URL environment variable is not set"

**Solution:**
```bash
# Set the variable
export DATABASE_URL="your-neon-connection-string"

# Or add to .env file
echo 'DATABASE_URL=your-neon-connection-string' > .env

# Verify it's set
echo $DATABASE_URL
```

### Issue: "Failed to connect to PostgreSQL"

**Solution:**
```bash
# Check your connection string format
# It should be: postgresql://user:password@host/dbname?sslmode=require

# Verify Neon database is running (check Neon console)
# Check your internet connection
# Try connecting with psql to verify credentials
```

### Issue: "Duplicate key value violates unique constraint"

This happens if you run the import script multiple times.

**Solution:**
The import script automatically handles this with `ON CONFLICT DO NOTHING`. The duplicate entries are simply skipped and counted.

### Issue: Schema differences between environments

**Solution:**
Both environments use the same code in `database.go`, so schemas are always identical. If you manually modified one database, redeploy the application to recreate the schema.

## Verifying Schema Matches

To ensure dev and prod schemas are identical:

```bash
# Development
export DATABASE_URL="your-dev-database-url"
go run . &
sleep 3
psql "$DATABASE_URL" -c "\d bookings" > dev-schema.txt
pkill -f "go run"

# Production
export DATABASE_URL="your-prod-database-url"
go run . &
sleep 3
psql "$DATABASE_URL" -c "\d bookings" > prod-schema.txt
pkill -f "go run"

# Compare
diff dev-schema.txt prod-schema.txt
# (should be empty if identical)
```

## Rollback Plan

If something goes wrong:

```bash
# 1. You still have your SQLite backup
cp bookings.db.backup bookings.db

# 2. Revert code to use SQLite
git checkout HEAD -- database.go handlers/api.go go.mod go.sum
go mod tidy

# 3. Run with SQLite
go run .
```

Your SQLite database is unchanged and ready to use.

## Next Steps After Migration

1. ✅ Test all features in development
2. ✅ Verify data integrity
3. ✅ Set up production Neon database
4. ✅ Deploy to AWS with production DATABASE_URL
5. ✅ Monitor application logs and Neon dashboard
6. ✅ Keep SQLite backup for a while (just in case)

## Quick Reference Commands

```bash
# View SQLite schema
./export_sqlite_schema.sh

# Export SQLite data
go run export_sqlite_data.go

# Create Neon schema (automatic)
export DATABASE_URL="postgresql://..."
go run .

# Import data to Neon
go run import_to_postgres.go

# Verify in Neon
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM bookings;"

# Run application
go run .
```

## Summary

The schema migration is **automatic** - you don't need to manually create tables in Neon. The application does it for you using the same code that worked with SQLite.

The only thing you need to decide is:
- **Fresh start:** Just run the app with Neon DATABASE_URL
- **Preserve data:** Export from SQLite → Import to Neon

Both approaches result in the same schema in Neon!
