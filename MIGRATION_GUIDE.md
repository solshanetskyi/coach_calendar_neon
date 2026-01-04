# Migration Guide: Production to Neon Database

This guide explains how to migrate bookings and blocked slots from your production AWS App Runner deployment to your new Neon PostgreSQL database.

## Overview

The migration script (`scripts/migrate_from_production.go`) will:
1. Fetch all bookings and blocked slots from your production API
2. Insert them into your Neon database
3. Skip any records that already exist (based on `slot_time`)
4. Provide a detailed summary of the migration

## Prerequisites

1. **Production database name**: Make sure you've created a `bookings` database in your Neon console
2. **Connection string**: Update `.env.production.example` with your Neon production connection string
3. **Network access**: Ensure you can access both:
   - Production API: `https://sweip8cyfh.eu-central-1.awsapprunner.com`
   - Neon database (outbound connections allowed)

## Migration Steps

### Option 1: Using the Shell Script (Recommended)

The easiest way to run the migration:

```bash
./run_migration.sh
```

This script will:
- Load the `DATABASE_URL` from `.env.production.example`
- Run the migration
- Ask for confirmation before proceeding

### Option 2: Manual Execution

If you prefer to run it manually:

```bash
# Load the production database URL
export DATABASE_URL='postgresql://neondb_owner:password@ep-xxx.neon.tech/bookings?sslmode=require'

# Run the migration
go run scripts/migrate_from_production.go
```

## What the Migration Does

### 1. Fetches Data from Production
The script calls the production API endpoint:
```
GET https://sweip8cyfh.eu-central-1.awsapprunner.com/api/admin/slots
```

This returns all slots with their status:
- `"booked"` - Slots with customer bookings
- `"blocked"` - Manually blocked slots
- `"available"` - Available slots (ignored during migration)

### 2. Inserts into Neon Database

For **booked** slots:
- Inserts into the `bookings` table
- Includes: slot_time, name, email, created_at, duration
- Uses `ON CONFLICT (slot_time) DO NOTHING` to skip duplicates

For **blocked** slots:
- Inserts into the `blocked_slots` table
- Includes: slot_time, created_at
- Uses `ON CONFLICT (slot_time) DO NOTHING` to skip duplicates

### 3. Provides Summary

After completion, you'll see:
```
========================================
Migration Summary
========================================
Bookings fetched:    25
Bookings inserted:   25
Bookings skipped:    0
----------------------------------------
Blocked slots fetched:  10
Blocked slots inserted: 10
Blocked slots skipped:  0
========================================
```

## Important Notes

### Time Zone Handling
- The production API returns times in `+01:00` (CET/Amsterdam timezone)
- The script converts all times to UTC before storing in the database
- This ensures consistent timezone handling

### Duplicate Detection
- The script uses the `slot_time` field to detect duplicates
- If a booking or blocked slot already exists, it will be skipped
- This makes the script safe to run multiple times

### No Data Loss
- The script **does not delete** any existing data
- It only **inserts** new records
- Existing records are preserved

## Troubleshooting

### Error: "DATABASE_URL environment variable is not set"

**Solution**: Make sure `.env.production.example` has the correct `DATABASE_URL`:
```bash
DATABASE_URL=postgresql://neondb_owner:password@ep-xxx.neon.tech/bookings?sslmode=require
```

### Error: "SCRAM-SHA-256 error"

**Solution**: Remove `channel_binding=require` from your DATABASE_URL. The correct format is:
```
?sslmode=require
```
**NOT**:
```
?sslmode=require&channel_binding=require
```

### Error: "failed to ping database"

**Possible causes**:
1. Incorrect database credentials
2. Database doesn't exist (create the `bookings` database in Neon console)
3. Network connectivity issues
4. Firewall blocking outbound connections

### Error: "API returned status 500"

**Possible causes**:
1. Production service is down
2. Network connectivity issues

**Solution**: Check if the production URL is accessible:
```bash
curl https://sweip8cyfh.eu-central-1.awsapprunner.com/health
```

## Verification

After migration, verify the data was transferred correctly:

### Check Bookings Count
```sql
SELECT COUNT(*) FROM bookings;
```

### Check Blocked Slots Count
```sql
SELECT COUNT(*) FROM blocked_slots;
```

### View Recent Bookings
```sql
SELECT slot_time, name, email FROM bookings ORDER BY slot_time DESC LIMIT 10;
```

### View Blocked Slots
```sql
SELECT slot_time FROM blocked_slots ORDER BY slot_time DESC LIMIT 10;
```

## Post-Migration Steps

1. **Test the Neon database**: Run your application with the Neon database and verify everything works
2. **Update AWS App Runner**: Once verified, update your App Runner service to use the Neon database
3. **Monitor**: Keep an eye on the application logs for any issues

## Rollback

If you need to rollback:

### Clear Migrated Data
```sql
-- Clear all bookings
TRUNCATE TABLE bookings CASCADE;

-- Clear all blocked slots
TRUNCATE TABLE blocked_slots CASCADE;
```

### Revert to SQLite
Update your AWS App Runner environment variables to use the old SQLite database URL.

## Support

If you encounter issues:
1. Check the error messages in the migration output
2. Verify your database connection string
3. Ensure the `bookings` database exists in Neon
4. Check the troubleshooting section above
