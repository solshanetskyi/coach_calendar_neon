#!/bin/bash

# Export SQLite Schema Script
# This script extracts the schema from your SQLite database
# and shows you what tables and indexes exist

set -e

if [ ! -f "bookings.db" ]; then
    echo "❌ Error: bookings.db not found"
    echo "Make sure you're in the correct directory"
    exit 1
fi

echo "📊 SQLite Database Schema"
echo "========================="
echo ""

echo "Tables and their structure:"
echo "---"
sqlite3 bookings.db ".schema"
echo ""

echo "---"
echo "Table sizes:"
echo "---"
echo "Bookings:"
sqlite3 bookings.db "SELECT COUNT(*) FROM bookings;"

echo "Blocked Slots:"
sqlite3 bookings.db "SELECT COUNT(*) FROM blocked_slots;"
echo ""

echo "---"
echo "Sample data from bookings (first 3):"
echo "---"
sqlite3 bookings.db "SELECT id, slot_time, name, email FROM bookings LIMIT 3;"
echo ""

echo "---"
echo "Sample data from blocked_slots (first 3):"
echo "---"
sqlite3 bookings.db "SELECT id, slot_time FROM blocked_slots LIMIT 3;"
echo ""

echo "✅ Schema inspection complete!"
echo ""
echo "Note: The Go application automatically creates an equivalent"
echo "PostgreSQL schema in Neon when you run it."
