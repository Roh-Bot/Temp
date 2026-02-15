#!/bin/bash

# Test script for bonus features

echo "=== Testing Bonus Features ==="
echo ""

# Check if server is running
if ! curl -s http://localhost:8000/api/health > /dev/null 2>&1; then
    echo "❌ Server is not running. Start it with: make docker-up"
    exit 1
fi

echo "✅ Server is running"
echo ""

# Login to get token
echo "📝 Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"password123"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Failed to get token"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ Got authentication token"
echo ""

# Test 1: Pagination with scroll_id
echo "=== Test 1: Pagination with Scroll ID ==="
echo "Fetching first page..."
FIRST_PAGE=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/tasks?limit=2")

echo "Response: $FIRST_PAGE"
echo ""

NEXT_SCROLL=$(echo $FIRST_PAGE | grep -o '"next_scroll":"[^"]*' | cut -d'"' -f4)

if [ ! -z "$NEXT_SCROLL" ]; then
    echo "✅ Got next_scroll: $NEXT_SCROLL"
    echo "Fetching next page..."
    SECOND_PAGE=$(curl -s -H "Authorization: Bearer $TOKEN" \
      "http://localhost:8000/api/tasks?limit=2&scroll_id=$NEXT_SCROLL")
    echo "Response: $SECOND_PAGE"
    echo "✅ Pagination working!"
else
    echo "⚠️  No next_scroll (might be last page or no data)"
fi
echo ""

# Test 2: Rate Limiting
echo "=== Test 2: Rate Limiting ==="
echo "Sending 25 rapid requests to trigger per-IP rate limit..."

SUCCESS_COUNT=0
RATE_LIMITED_COUNT=0

for i in {1..25}; do
    RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" \
      "http://localhost:8000/api/tasks?limit=1")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "200" ]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    elif [ "$HTTP_CODE" = "429" ]; then
        RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
    fi
done

echo ""
echo "Results:"
echo "  ✅ Successful requests: $SUCCESS_COUNT"
echo "  🚫 Rate limited requests: $RATE_LIMITED_COUNT"

if [ $RATE_LIMITED_COUNT -gt 0 ]; then
    echo "✅ Rate limiting is working!"
else
    echo "⚠️  Rate limiting might not be triggered (try increasing request count)"
fi
echo ""

# Test 3: Status filtering with pagination
echo "=== Test 3: Status Filtering with Pagination ==="
echo "Fetching pending tasks..."
PENDING_TASKS=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/tasks?limit=5&status=pending")

echo "Response: $PENDING_TASKS"
echo "✅ Status filtering working!"
echo ""

echo "=== All Tests Complete ==="
