#!/bin/bash

# Migration script to migrate data from production to Neon database
# This script loads the DATABASE_URL from .env.production.example and runs the migration

set -e

echo "=========================================="
echo "Production to Neon Migration Script"
echo "=========================================="

# Check if .env.production.example exists
if [ ! -f .env.production.example ]; then
    echo "Error: .env.production.example file not found"
    exit 1
fi

# Load DATABASE_URL from .env.production.example
export $(grep "^DATABASE_URL=" .env.production.example | xargs)

if [ -z "$DATABASE_URL" ]; then
    echo "Error: DATABASE_URL not found in .env.production.example"
    exit 1
fi

echo "Loaded DATABASE_URL from .env.production.example"
echo ""

# Run the migration
go run scripts/migrate_from_production.go
