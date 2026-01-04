# Neon Migration Summary

## What Changed

Your Coach Calendar application has been migrated from SQLite to Neon PostgreSQL database.

## Quick Start

### 1. Set up Neon Database
- Create a project at https://console.neon.tech
- Copy your connection string

### 2. Configure Environment
```bash
cp .env.example .env
# Edit .env and add your DATABASE_URL
```

### 3. (Optional) Export SQLite Data
```bash
go run export_sqlite_data.go
```

### 4. Install Dependencies & Run
```bash
go mod tidy
go run .
```

### 5. (Optional) Import Old Data
```bash
go run import_to_postgres.go
```

## Files Changed

### Modified Files
- [database.go](database.go) - PostgreSQL connection and table creation
- [handlers/api.go](handlers/api.go) - PostgreSQL query syntax ($1, $2 instead of ?)
- [go.mod](go.mod) - Changed from SQLite to PostgreSQL driver
- [.env.example](.env.example) - Added DATABASE_URL configuration
- [.gitignore](.gitignore) - Added export JSON files

### New Files
- [NEON_MIGRATION.md](NEON_MIGRATION.md) - Complete migration guide
- [export_sqlite_data.go](export_sqlite_data.go) - Export SQLite data to JSON
- [import_to_postgres.go](import_to_postgres.go) - Import JSON data to PostgreSQL

## Key Technical Changes

### Database Driver
```go
// Before
import _ "modernc.org/sqlite"
db, err = sql.Open("sqlite", "./bookings.db")

// After
import _ "github.com/lib/pq"
db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
```

### SQL Syntax
```sql
-- Before (SQLite)
CREATE TABLE bookings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  slot_time DATETIME NOT NULL UNIQUE
);

-- After (PostgreSQL)
CREATE TABLE bookings (
  id SERIAL PRIMARY KEY,
  slot_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE
);
```

### Query Parameters
```go
// Before (SQLite)
db.Exec("INSERT INTO bookings VALUES (?, ?, ?)", a, b, c)

// After (PostgreSQL)
db.Exec("INSERT INTO bookings VALUES ($1, $2, $3)", a, b, c)
```

### Error Detection
```go
// Before (SQLite)
if err.Error() == "UNIQUE constraint failed: bookings.slot_time" {

// After (PostgreSQL)
if strings.Contains(err.Error(), "duplicate key value") ||
   strings.Contains(err.Error(), "unique constraint") {
```

## Environment Variables

Required:
- `DATABASE_URL` - Neon PostgreSQL connection string

Optional:
- `PORT` - Server port (default: 8080)
- `SMTP_*` or `USE_AWS_SES` - Email configuration

## Benefits of Neon

✅ **Scalability** - Handle more concurrent users
✅ **Cloud-native** - No local database files to manage
✅ **Auto-scaling** - Scales with your usage
✅ **Backups** - Automatic point-in-time recovery
✅ **Branching** - Database branching for testing
✅ **Better Timezones** - Superior timezone handling

## Rollback to SQLite

If needed, rollback using git:
```bash
git checkout HEAD -- database.go handlers/api.go go.mod go.sum
go mod tidy
```

## Next Steps

1. Test the application locally with Neon
2. Verify all features work (booking, blocking, admin panel)
3. Update deployment configuration with DATABASE_URL
4. Deploy to production

## Need Help?

- See [NEON_MIGRATION.md](NEON_MIGRATION.md) for detailed instructions
- Neon Docs: https://neon.tech/docs
- Neon Discord: https://discord.gg/92vNTzKDGp
