#!/bin/bash

# Task Manager API Test Script

BASE_URL="http://localhost:8000/api"
TOKEN=""

echo "=== Task Manager API Test Script ==="
echo ""

# Test 1: Health Check
echo "1. Testing Health Check..."
curl -s "$BASE_URL/health" | jq .
echo ""

# Test 2: Register User
echo "2. Registering a new user..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "role": "user"
  }' | jq .
echo ""

# Test 3: Login
echo "3. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }')
echo "$LOGIN_RESPONSE" | jq .
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.Data.token')
echo "Token: $TOKEN"
echo ""

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "Failed to get token. Exiting."
  exit 1
fi

# Test 4: Create Task
echo "4. Creating a task..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Complete assignment",
    "description": "Finish the Go task manager project"
  }')
echo "$CREATE_RESPONSE" | jq .
echo ""

# Test 5: List Tasks
echo "5. Listing tasks..."
curl -s -X GET "$BASE_URL/tasks?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# Test 6: List Tasks with Filter
echo "6. Listing pending tasks..."
curl -s -X GET "$BASE_URL/tasks?page=1&limit=10&status=pending" \
  -H "Authorization: Bearer $TOKEN" | jq .
echo ""

# Test 7: Create another task
echo "7. Creating another task..."
curl -s -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Write documentation",
    "description": "Document the API endpoints"
  }' | jq .
echo ""

# Test 8: List all tasks again
echo "8. Listing all tasks..."
TASKS_RESPONSE=$(curl -s -X GET "$BASE_URL/tasks?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN")
echo "$TASKS_RESPONSE" | jq .
TASK_ID=$(echo "$TASKS_RESPONSE" | jq -r '.Data.tasks[0].id')
echo "First Task ID: $TASK_ID"
echo ""

# Test 9: Get specific task
if [ "$TASK_ID" != "null" ] && [ -n "$TASK_ID" ]; then
  echo "9. Getting task by ID..."
  curl -s -X GET "$BASE_URL/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN" | jq .
  echo ""
  
  # Test 10: Delete task
  echo "10. Deleting task..."
  curl -s -X DELETE "$BASE_URL/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN"
  echo ""
  echo "Task deleted"
  echo ""
fi

# Test 11: Register Admin
echo "11. Registering an admin user..."
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "role": "admin"
  }' | jq .
echo ""

# Test 12: Admin Login
echo "12. Admin logging in..."
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')
echo "$ADMIN_LOGIN_RESPONSE" | jq .
ADMIN_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | jq -r '.Data.token')
echo "Admin Token: $ADMIN_TOKEN"
echo ""

# Test 13: Admin List All Tasks
if [ "$ADMIN_TOKEN" != "null" ] && [ -n "$ADMIN_TOKEN" ]; then
  echo "13. Admin listing all tasks (across all users)..."
  curl -s -X GET "$BASE_URL/tasks?page=1&limit=10" \
    -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
  echo ""
fi

echo "=== Test Complete ==="
