# VS Code Database Connection Guide

This guide shows you how to connect to your Neon PostgreSQL database from Visual Studio Code.

## Quick Start - Using SQLTools (Recommended)

### Step 1: Install Extensions

In VS Code, install these two extensions:

1. **SQLTools** by Matheus Teixeira
   - Press ⌘+Shift+X (Extensions)
   - Search: "SQLTools"
   - Click Install

2. **SQLTools PostgreSQL/Cockroach Driver**
   - Search: "SQLTools PostgreSQL"
   - Click Install

### Step 2: Connect to Database

The connection configuration is already set up in `.vscode/sqltools_connections.json`!

1. Click the **SQLTools icon** in the left sidebar (database icon)
2. You'll see two connections:
   - **Neon Production - Bookings**
   - **Neon Development - Bookings**

3. Click on either connection
4. Enter password when prompted:
   - Production: `npg_Aho5nl0CPtHZ` (from `.env.production.example`)
   - Development: `npg_Aho5nl0CPtHZ` (from `.env`)

5. Click "Connect"

### Step 3: Start Using the Database

Once connected, you can:

- **View Tables**: Expand the connection → expand "public" schema → see tables
  - `bookings`
  - `blocked_slots`

- **Browse Data**: Right-click a table → "Show Table Records"

- **Run Queries**: Click "New SQL File" or press ⌘+E ⌘+E
  ```sql
  -- Example queries
  SELECT * FROM bookings ORDER BY slot_time DESC LIMIT 10;
  SELECT COUNT(*) FROM blocked_slots;
  SELECT name, email, slot_time FROM bookings WHERE slot_time > NOW();
  ```

- **Execute Query**: Place cursor on query → press ⌘+E ⌘+E

## Alternative: Using Terminal (psql)

### Option 1: Using the Helper Script

I've created a script that automatically connects:

```bash
./connect_neon.sh
```

This loads the DATABASE_URL from `.env.production.example` and connects via psql.

### Option 2: Manual psql Connection

Open VS Code's integrated terminal (⌘+`) and run:

```bash
# Load environment variable
export $(grep "^DATABASE_URL=" .env.production.example | xargs)

# Connect
psql "$DATABASE_URL"
```

Or directly:

```bash
psql "postgresql://neondb_owner:npg_Aho5nl0CPtHZ@ep-quiet-darkness-ag9g8x6n-pooler.c-2.eu-central-1.aws.neon.tech/bookings?sslmode=require"
```

### Common psql Commands

Once connected:

```sql
-- List all tables
\dt

-- Describe a table
\d bookings

-- View table data
SELECT * FROM bookings LIMIT 5;

-- Check database size
SELECT pg_size_pretty(pg_database_size('bookings'));

-- Quit
\q
```

## Connection Details

### Production Database
```
Host: ep-quiet-darkness-ag9g8x6n-pooler.c-2.eu-central-1.aws.neon.tech
Port: 5432
Database: bookings
Username: neondb_owner
Password: npg_Aho5nl0CPtHZ
SSL: Required
```

### Development Database
```
Host: ep-small-sound-agj17yfr-pooler.c-2.eu-central-1.aws.neon.tech
Port: 5432
Database: bookings
Username: neondb_owner
Password: npg_Aho5nl0CPtHZ
SSL: Required
```

## Useful Queries

### Check Migration Success

```sql
-- Count bookings
SELECT COUNT(*) as total_bookings FROM bookings;

-- Count blocked slots
SELECT COUNT(*) as total_blocked FROM blocked_slots;

-- Recent bookings
SELECT
    slot_time AT TIME ZONE 'Europe/Amsterdam' as slot_time_cet,
    name,
    email,
    created_at
FROM bookings
ORDER BY slot_time DESC
LIMIT 10;

-- Upcoming bookings
SELECT
    slot_time AT TIME ZONE 'Europe/Amsterdam' as slot_time_cet,
    name,
    email
FROM bookings
WHERE slot_time > NOW()
ORDER BY slot_time ASC;
```

### Database Administration

```sql
-- Check table sizes
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- View all indexes
SELECT
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public';

-- Check database connections
SELECT
    datname,
    usename,
    application_name,
    client_addr,
    state
FROM pg_stat_activity
WHERE datname = 'bookings';
```

### Data Verification

```sql
-- Check for duplicate slot times
SELECT slot_time, COUNT(*)
FROM bookings
GROUP BY slot_time
HAVING COUNT(*) > 1;

-- Check timezone storage (should be UTC)
SELECT
    slot_time,
    slot_time AT TIME ZONE 'UTC' as utc_time,
    slot_time AT TIME ZONE 'Europe/Amsterdam' as amsterdam_time
FROM bookings
LIMIT 5;
```

## Troubleshooting

### "Could not connect to database"

1. **Check internet connection**
2. **Verify credentials** in `.env.production.example`
3. **Check Neon status**: https://neon.tech/status
4. **Try without SSL** (temporarily for testing):
   - In connection settings, disable SSL
   - Note: Not recommended for production

### "SSL connection required"

Make sure SSL is enabled in your connection settings:
- SQLTools: Check "Use SSL" in connection settings
- psql: Include `?sslmode=require` in connection string

### "Database 'bookings' does not exist"

Create it in Neon Console:
1. Go to https://console.neon.tech
2. Select your project
3. Databases → Create Database
4. Name: `bookings`

### "Password authentication failed"

1. Double-check the password in `.env.production.example`
2. Copy-paste to avoid typos
3. Make sure there are no extra spaces or quotes

## Security Notes

⚠️ **Important Security Practices**:

1. **Never commit credentials** to git
   - `.env.production.example` is in `.gitignore`
   - SQLTools stores passwords in VS Code settings (not git)

2. **Use different databases** for dev/prod
   - Development: `ep-small-sound-agj17yfr-pooler`
   - Production: `ep-quiet-darkness-ag9g8x6n-pooler`

3. **Be careful with destructive queries**
   - Always use `WHERE` clauses with `UPDATE`/`DELETE`
   - Test on development first
   - Use transactions: `BEGIN; ... COMMIT;` or `ROLLBACK;`

## Files Created

- **`.vscode/sqltools_connections.json`** - SQLTools connection configuration
- **`connect_neon.sh`** - Helper script for psql connection
- **`VSCODE_DATABASE_SETUP.md`** - This guide

## Next Steps

1. Install SQLTools extensions
2. Connect to your database
3. Run some test queries
4. Verify migration data (if you've run the migration)
5. Explore the data and schema

Happy querying! 🎉
