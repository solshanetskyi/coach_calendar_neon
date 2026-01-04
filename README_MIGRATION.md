# Quick Start: Production to Neon Migration

## What You Need to Know

This migration will transfer **11 bookings** and **219 blocked slots** from your production AWS App Runner instance to your new Neon PostgreSQL database.

## Before You Start

1. **Create the `bookings` database in Neon Console**:
   - Go to https://console.neon.tech
   - Select your project
   - Create a new database named `bookings`

2. **Update `.env.production.example`** with your Neon production connection string:
   ```bash
   DATABASE_URL=postgresql://neondb_owner:your-password@ep-xxx.neon.tech/bookings?sslmode=require
   ```
   ⚠️ **Important**: Remove `channel_binding=require` if present in the URL

## Run the Migration

Simply execute:

```bash
./run_migration.sh
```

The script will:
1. Load the database URL from `.env.production.example`
2. Ask for confirmation
3. Fetch all data from production
4. Insert it into your Neon database
5. Show a summary of what was migrated

## What Gets Migrated

From: `https://sweip8cyfh.eu-central-1.awsapprunner.com/api/admin/slots`

- ✅ **Bookings** (11 records): customer name, email, slot time
- ✅ **Blocked Slots** (219 records): slot time
- ❌ **Available Slots**: Not migrated (they're generated dynamically)

## Safety Features

- **No duplicates**: Uses `ON CONFLICT DO NOTHING` - safe to run multiple times
- **No deletions**: Only inserts new data, never deletes existing records
- **Timezone safe**: Converts all times from CET (+01:00) to UTC for storage
- **Password masking**: Logs show masked database credentials

## Expected Output

```
==========================================
Production to Neon Migration Script
==========================================
Production URL: https://sweip8cyfh.eu-central-1.awsapprunner.com
Target Database: postgresql://neondb_owner:****@ep-xxx.neon.tech/bookings

This will migrate all bookings and blocked slots from production to Neon.
Existing records with the same slot_time will be skipped.
Do you want to continue? (yes/no): yes

Connecting to Neon database...
✓ Connected to Neon database

Fetching data from production API...
✓ Fetched 476 slots from production

Migrating data...
  ✓ Booking inserted: 2026-01-05T09:00:00+01:00 - Тортик (sergii.olshanetskyi@gmail.com)
  ✓ Booking inserted: 2026-01-05T10:00:00+01:00 - Tetiana (tanyakuklinska95@gmail.com)
  ...
  ✓ Blocked slot inserted: 2026-01-05T11:30:00+01:00
  ...

========================================
Migration Summary
========================================
Bookings fetched:    11
Bookings inserted:   11
Bookings skipped:    0
----------------------------------------
Blocked slots fetched:  219
Blocked slots inserted: 219
Blocked slots skipped:  0
========================================

✓ Migration completed successfully!
```

## After Migration

Verify the data:

```bash
# Connect to your Neon database
psql "$DATABASE_URL"

# Check counts
SELECT COUNT(*) FROM bookings;      -- Should show 11
SELECT COUNT(*) FROM blocked_slots; -- Should show 219

# View sample data
SELECT * FROM bookings ORDER BY slot_time LIMIT 5;
```

## Files Created

- **`scripts/migrate_from_production.go`** - The migration script
- **`run_migration.sh`** - Shell wrapper to run the migration
- **`MIGRATION_GUIDE.md`** - Detailed documentation
- **`README_MIGRATION.md`** - This quick start guide

## Troubleshooting

If you see "SCRAM-SHA-256 error":
```bash
# Remove channel_binding from your DATABASE_URL
# Wrong:   ?sslmode=require&channel_binding=require
# Correct: ?sslmode=require
```

For detailed troubleshooting, see [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md).

## Next Steps

After successful migration:
1. Test your application with the Neon database locally
2. Update AWS App Runner environment variables to use Neon
3. Deploy and verify production is working with Neon
4. Keep the old database as backup for a few days

## Need Help?

See the full documentation: [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md)
