# Database Setup Guide

This application uses a dedicated `bookings` database in Neon PostgreSQL.

## Setting up the Bookings Database

### 1. Create the Database in Neon Console

1. Go to your Neon Console: https://console.neon.tech
2. Select your project
3. Click on "Databases" in the sidebar
4. Click "Create Database"
5. Name it: `bookings`
6. Click "Create"

### 2. Update Your Connection String

After creating the `bookings` database, get the connection string:

1. In Neon Console, go to your `bookings` database
2. Copy the connection string (with pooler)
3. Update your `.env` file with the new connection string

The connection string should look like:
```
DATABASE_URL='postgresql://neondb_owner:your-password@ep-xxx-pooler.region.aws.neon.tech/bookings?sslmode=require'
```

**Important**: Make sure the database name in the URL is `bookings`, not `neondb`.

### 3. Important Notes

- **Do NOT include** `channel_binding=require` in your connection string. This parameter causes SCRAM-SHA-256 authentication errors with the Go `lib/pq` driver.
- The correct format is: `?sslmode=require` (without channel_binding)
- The application will automatically create the necessary tables (`bookings` and `blocked_slots`) when it first connects

### 4. Troubleshooting

If you see a warning:
```
Warning: Expected 'bookings' database but found 'neondb'
```

This means your DATABASE_URL is still pointing to the 'neondb' database. Make sure to:
1. Update the `.env` file with the correct database name
2. Restart your terminal/IDE to clear cached environment variables
3. Or run: `unset DATABASE_URL` before running `go run .`

### 5. Environment Files

The following files have been updated to use the `bookings` database:
- `.env` - Your local development configuration
- `.env.development.example` - Template for development
- `.env.example` - General template
- `.env.production.example` - Template for production

## Database Schema

The application automatically creates these tables:

### `bookings` table
- `id` - Primary key (auto-increment)
- `slot_time` - Timestamp with timezone (unique)
- `name` - User's name
- `email` - User's email
- `created_at` - When the booking was created
- `duration` - Booking duration in minutes (default: 30)
- `zoom_link` - Optional Zoom meeting link

### `blocked_slots` table
- `id` - Primary key (auto-increment)
- `slot_time` - Timestamp with timezone (unique)
- `created_at` - When the slot was blocked
