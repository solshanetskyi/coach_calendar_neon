#!/bin/bash
# Clear all blocked slots from production

PROD_URL="https://sweip8cyfh.eu-central-1.awsapprunner.com"

echo "🗑️  Fetching all blocked slots from production..."
SLOTS=$(curl -s "$PROD_URL/api/admin/debug-blocked" | jq -r '.[].slot_time_rfc3339')

COUNT=$(echo "$SLOTS" | wc -l | tr -d ' ')
echo "Found $COUNT slots to unblock"
echo ""

i=1
echo "$SLOTS" | while read slot; do
  if [ -z "$slot" ]; then
    continue
  fi

  echo "[$i/$COUNT] Unblocking: $slot"
  curl -s -X POST "$PROD_URL/api/admin/unblock" \
    -H "Content-Type: application/json" \
    -d "{\"slot_time\":\"$slot\"}" | jq -r '.message'

  i=$((i+1))
  sleep 0.1  # Small delay to avoid overwhelming the API
done

echo ""
echo "✅ All blocked slots cleared from production!"
echo "Now run: make block-slots-prod DAYS=90"
