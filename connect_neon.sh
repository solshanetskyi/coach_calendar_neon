#!/bin/bash

# Script to connect to Neon database via psql
# This loads the DATABASE_URL from .env.production.example

set -e

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

echo "Connecting to Neon database..."
echo ""

# Connect using psql
psql "$DATABASE_URL"
