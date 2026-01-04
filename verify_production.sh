#!/bin/bash
# Script to verify production deployment

echo "🔍 Verifying production blocked slots fix..."
echo ""

echo "1. Checking admin API for blocked slots:"
curl -s https://sweip8cyfh.eu-central-1.awsapprunner.com/api/admin/slots | \
  jq '.[] | select(.status == "blocked")' | head -10

echo ""
echo "2. Checking user API - the slot we just blocked should show as unavailable:"
curl -s https://sweip8cyfh.eu-central-1.awsapprunner.com/api/slots | \
  jq '.[] | select(.slot_time == "2026-01-06T11:30:00+01:00")'

echo ""
echo "✅ If you see blocked slots above, the fix is deployed successfully!"
echo "❌ If no blocked slots appear, the deployment may still be in progress."
