#!/bin/bash

# Setup Development Environment Script
# This script helps you quickly set up your local development environment

set -e  # Exit on error

echo "🚀 Coach Calendar - Development Environment Setup"
echo "=================================================="
echo ""

# Check if .env already exists
if [ -f ".env" ]; then
    echo "⚠️  .env file already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Setup cancelled. Existing .env file kept."
        exit 0
    fi
fi

# Copy the development example
if [ -f ".env.development.example" ]; then
    cp .env.development.example .env
    echo "✅ Created .env from .env.development.example"
else
    cp .env.example .env
    echo "✅ Created .env from .env.example"
fi

echo ""
echo "📝 Next steps:"
echo ""
echo "1. Edit .env file and add your Neon development database URL"
echo "   - Go to https://console.neon.tech"
echo "   - Create a 'Development' project (or development branch)"
echo "   - Copy the connection string"
echo "   - Paste it into DATABASE_URL in .env"
echo ""
echo "2. (Optional) Configure email settings in .env"
echo "   - For local testing, you can skip this"
echo "   - Or use a Gmail account with app password"
echo ""
echo "3. Install dependencies:"
echo "   go mod tidy"
echo ""
echo "4. Run the application:"
echo "   go run ."
echo ""
echo "5. Visit http://localhost:8080 in your browser"
echo ""
echo "📚 For more details, see:"
echo "   - NEON_ENVIRONMENTS.md - Managing dev/prod environments"
echo "   - NEON_MIGRATION.md - Database migration guide"
echo "   - README.md - Full application documentation"
echo ""
