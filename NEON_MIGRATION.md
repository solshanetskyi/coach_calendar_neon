# Migration from SQLite to Neon (PostgreSQL)

This guide will help you migrate your Coach Calendar application from SQLite to Neon PostgreSQL database.

## Quick Overview

**Schema Migration:** ✅ Automatic! The application creates the schema for you.
**Data Migration:** 📦 Use the export/import scripts (optional, only if you have existing data).

📖 **For detailed schema migration information, see [SCHEMA_MIGRATION_GUIDE.md](SCHEMA_MIGRATION_GUIDE.md).**

## Prerequisites

1. A Neon account (sign up at https://neon.tech)
2. Go 1.21 or higher installed
3. Access to your existing `bookings.db` SQLite database (only if you want to preserve data)

## Migration Steps

### Step 1: Set up Neon Database

1. Log in to your Neon account at https://console.neon.tech
2. Create a new project (or use an existing one)
3. Copy your connection string from the Neon dashboard
   - It should look like: `postgresql://user:password@ep-xxx.region.aws.neon.tech/dbname?sslmode=require`

### Step 2: Configure Environment Variables

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your Neon connection string:
   ```
   DATABASE_URL=postgresql://user:password@ep-xxx.region.aws.neon.tech/dbname?sslmode=require
   ```

### Step 3: Export Existing Data (Optional)

If you have existing bookings or blocked slots in SQLite that you want to migrate:

1. Ensure your `bookings.db` file exists in the current directory

2. Run the export script:
   ```bash
   go run export_sqlite_data.go
   ```

   Note: The export script uses build tags to exclude it from normal builds, so it must be run explicitly with `go run`.

3. This will create two files:
   - `bookings_export.json` - Your existing bookings
   - `blocked_slots_export.json` - Your blocked time slots

### Step 4: Update Dependencies

Install the PostgreSQL driver:

```bash
go get github.com/lib/pq
go mod tidy
```

### Step 5: Test the Connection

Run the application to verify it connects to Neon:

```bash
go run .
```

The application will automatically:
- Connect to your Neon database
- Create the necessary tables (`bookings` and `blocked_slots`)
- Create indexes for performance

### Step 6: Import Existing Data (Optional)

If you exported data in Step 3, import it to Neon:

```bash
go run import_to_postgres.go
```

This will:
- Read the exported JSON files
- Insert all bookings into your Neon database
- Insert all blocked slots into your Neon database

### Step 7: Verify Migration

1. Start the application:
   ```bash
   go run .
   ```

2. Visit the admin panel:
   ```
   http://localhost:8080/admin
   ```

3. Verify that:
   - Existing bookings appear correctly
   - Blocked slots are shown
   - New bookings can be created
   - Time slots can be blocked/unblocked

### Step 8: Deploy to Production

When deploying to AWS App Runner or other platforms:

1. Set the `DATABASE_URL` environment variable in your deployment configuration
2. The application will automatically use the Neon database

## Key Changes Made

### Database Connection
- Changed from `modernc.org/sqlite` to `github.com/lib/pq` (PostgreSQL driver)
- Connection now uses `DATABASE_URL` environment variable
- Added SSL mode support for secure Neon connections

### SQL Syntax Updates
- Changed `INTEGER PRIMARY KEY AUTOINCREMENT` to `SERIAL PRIMARY KEY`
- Changed `DATETIME` to `TIMESTAMP WITH TIME ZONE`
- Changed `?` placeholders to `$1, $2, $3...` (PostgreSQL parameterized queries)
- Updated UNIQUE constraint error detection for PostgreSQL

### Schema Changes
- PostgreSQL uses `SERIAL` for auto-incrementing IDs
- Timestamps now use PostgreSQL's `TIMESTAMP WITH TIME ZONE` type
- Default timestamp uses `CURRENT_TIMESTAMP` (compatible with both)

## Rollback Plan

If you need to rollback to SQLite:

1. Keep your original `bookings.db` file as backup
2. Revert the code changes using git:
   ```bash
   git checkout HEAD -- database.go handlers/api.go go.mod go.sum
   ```
3. Reinstall SQLite dependencies:
   ```bash
   go mod tidy
   ```

## Benefits of Using Neon

- **Scalability**: PostgreSQL handles much larger datasets than SQLite
- **Concurrent Writes**: Multiple users can book simultaneously without locking issues
- **Cloud-native**: No need to manage database files
- **Auto-scaling**: Neon automatically scales based on your usage
- **Branching**: Neon supports database branching for testing
- **Backups**: Automatic point-in-time recovery
- **Better timezone handling**: PostgreSQL has superior timezone support

## Troubleshooting

### Connection Issues

If you see "connection refused" or timeout errors:
- Verify your `DATABASE_URL` is correct
- Check that your IP is allowed in Neon's firewall settings (if configured)
- Ensure SSL mode is set to `require` in the connection string

### Data Type Errors

If you see errors about data types:
- Make sure you're using the latest version of the code
- PostgreSQL is stricter about types than SQLite
- Check that timestamps are in RFC3339 format

### Performance Issues

If queries are slow:
- Verify that indexes were created (check application logs on startup)
- Consider adding additional indexes based on your query patterns
- Monitor query performance in the Neon dashboard

## Support

For Neon-specific issues:
- Documentation: https://neon.tech/docs
- Discord: https://discord.gg/92vNTzKDGp
- Support: support@neon.tech

For application issues:
- Check the application logs
- Verify environment variables are set correctly
- Test database connectivity independently
