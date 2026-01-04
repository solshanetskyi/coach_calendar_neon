# Zoom Link Schema Migration

## Overview
Updated the bookings schema to store Zoom meeting links and creation timestamps.

## Changes Made

### 1. Database Schema ([database.go](database.go))
- Added `zoom_link TEXT` column to the `bookings` table (nullable/optional)
- The `created_at` timestamp already existed in the schema

### 2. Booking Struct Updates
Updated the `Booking` struct in three files:
- [handlers/api.go](handlers/api.go#L24-L31) - Added `ZoomLink` field
- [export_sqlite_data.go](export_sqlite_data.go#L17-L25) - Added `ZoomLink` field
- [import_to_postgres.go](import_to_postgres.go#L17-L25) - Added `ZoomLink` field

### 3. API Handler Updates ([handlers/api.go](handlers/api.go#L165-L188))
- Modified `CreateBooking` to create Zoom meeting before database insert
- Store the Zoom link in the database along with booking details
- Zoom link is stored as a SQL `NullString` (empty if Zoom service is unavailable)

### 4. Export/Import Scripts
- [export_sqlite_data.go](export_sqlite_data.go#L48) - Query now includes `zoom_link` field
- [import_to_postgres.go](import_to_postgres.go#L72) - Import now includes `zoom_link` field

### 5. Migration SQL ([add_zoom_link_column.sql](add_zoom_link_column.sql))
Created a migration script to add the column to existing databases:
```sql
ALTER TABLE bookings
ADD COLUMN IF NOT EXISTS zoom_link TEXT;
```

## Database Schema

### Bookings Table
```sql
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    slot_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    duration INTEGER NOT NULL DEFAULT 30,
    zoom_link TEXT
);
```

## Migration Instructions

### For Existing Databases

Run the migration SQL on your existing database:

```bash
# For Neon/PostgreSQL
psql $DATABASE_URL -f add_zoom_link_column.sql
```

Or apply directly in your database client:
```sql
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS zoom_link TEXT;
```

### For New Installations

No action needed - the schema will be created automatically when the application starts.

## Behavior

- **New bookings**: Will include a Zoom link if both the Zoom service is configured AND `CREATE_ZOOM_MEETING` is set to "yes"/"true"
- **Existing bookings**: Will have `NULL` for the `zoom_link` field (backward compatible)
- **JSON responses**: The `zoom_link` field uses `omitempty` tag, so it won't appear in JSON if empty
- **Failure handling**: If Zoom meeting creation fails, the booking still succeeds (link will be empty)
- **Confirmation emails**: Only sent if `SEND_CONFIRMATION_EMAIL` is set to "yes"/"true"

## Environment Variables

Two new optional environment variables control feature toggles:

```bash
# Enable/disable confirmation emails (default: disabled)
SEND_CONFIRMATION_EMAIL=yes

# Enable/disable Zoom meeting creation (default: disabled)
CREATE_ZOOM_MEETING=yes
```

Both variables accept "yes" or "true" (case-insensitive) to enable the feature. Any other value or omission will disable the feature.

## Verification

Build the project to verify all changes compile correctly:
```bash
go build
```

## Rollback

If you need to remove the column (not recommended):
```sql
ALTER TABLE bookings DROP COLUMN IF EXISTS zoom_link;
```
