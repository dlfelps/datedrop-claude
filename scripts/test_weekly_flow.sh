#!/bin/bash
# Test Weekly Flow: seed -> create users -> complete quizzes -> match -> accept
set -e

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

check() {
  local desc="$1"
  local expected_code="$2"
  local actual_code="$3"
  if [ "$actual_code" -eq "$expected_code" ]; then
    echo "  PASS: $desc (HTTP $actual_code)"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $desc (expected $expected_code, got $actual_code)"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Date Drop Weekly Flow Test ==="
echo ""

# 1. Health check
echo "1. Health check"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
check "GET /health" 200 "$CODE"

# 2. Seed data
echo "2. Seed data"
SEED_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/debug/seed")
CODE=$(echo "$SEED_RESPONSE" | tail -1)
BODY=$(echo "$SEED_RESPONSE" | head -1)
check "POST /debug/seed" 200 "$CODE"
echo "  Seed result: $BODY"

# 3. Create a new user
echo "3. Create user"
CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@stanford.edu",
    "name": "Test User",
    "date_of_birth": "2000-01-15",
    "gender": "female",
    "orientations": ["straight"]
  }')
CODE=$(echo "$CREATE_RESPONSE" | tail -1)
BODY=$(echo "$CREATE_RESPONSE" | head -1)
check "POST /api/v1/users" 201 "$CODE"
USER_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "")
echo "  User ID: $USER_ID"

# 4. Login
echo "4. Login"
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email": "testuser@stanford.edu"}')
CODE=$(echo "$LOGIN_RESPONSE" | tail -1)
BODY=$(echo "$LOGIN_RESPONSE" | head -1)
check "POST /api/v1/auth/login" 200 "$CODE"
TOKEN=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])" 2>/dev/null || echo "$USER_ID")
echo "  Token: $TOKEN"

# 5. Get user
echo "5. Get user"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/users/$TOKEN")
check "GET /api/v1/users/:id" 200 "$CODE"

# 6. Update user
echo "6. Update user"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X PATCH "$BASE_URL/api/v1/users/$TOKEN" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"bio": "Updated bio!"}')
check "PATCH /api/v1/users/:id" 200 "$CODE"

# 7. Get quiz questions
echo "7. Get quiz questions"
QUESTIONS_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/quiz/questions")
CODE=$(echo "$QUESTIONS_RESPONSE" | tail -1)
check "GET /api/v1/quiz/questions" 200 "$CODE"

# 8. Get quiz status
echo "8. Get quiz status (before)"
STATUS_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/quiz/status/$TOKEN")
CODE=$(echo "$STATUS_RESPONSE" | tail -1)
check "GET /api/v1/quiz/status" 200 "$CODE"

# 9. Run matching
echo "9. Run weekly matching"
MATCH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/matching/run" \
  -H "Authorization: Bearer $TOKEN")
CODE=$(echo "$MATCH_RESPONSE" | tail -1)
BODY=$(echo "$MATCH_RESPONSE" | head -1)
check "POST /api/v1/matching/run" 200 "$CODE"
echo "  Match result: $BODY"

# 10. Get current drop (use a seeded user token)
# Get first seeded user
SEEDED_USERS=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/api/v1/social/users")
FIRST_USER_ID=$(echo "$SEEDED_USERS" | python3 -c "import sys,json; users=json.load(sys.stdin)['users']; print(users[0]['id'] if users else '')" 2>/dev/null || echo "")

if [ -n "$FIRST_USER_ID" ]; then
  echo "10. Get current drop for seeded user"
  DROP_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $FIRST_USER_ID" \
    "$BASE_URL/api/v1/drops/current")
  CODE=$(echo "$DROP_RESPONSE" | tail -1)
  BODY=$(echo "$DROP_RESPONSE" | head -1)
  DROP_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "")

  if [ -n "$DROP_ID" ]; then
    check "GET /api/v1/drops/current" 200 "$CODE"
    echo "  Drop ID: $DROP_ID"

    # 11. Accept drop
    echo "11. Accept drop"
    CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
      -H "Authorization: Bearer $FIRST_USER_ID" \
      "$BASE_URL/api/v1/drops/$DROP_ID/accept")
    check "POST /api/v1/drops/:id/accept" 200 "$CODE"
  else
    echo "10. No drop found for seeded user (they may not have been matched)"
  fi
else
  echo "10. No seeded users found to test drops"
fi

# 12. Get drop history
echo "12. Get drop history"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" \
  "$BASE_URL/api/v1/drops/history")
check "GET /api/v1/drops/history" 200 "$CODE"

# Validate .edu email enforcement
echo "13. Reject non-.edu email"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"email":"bad@gmail.com","name":"Bad User","date_of_birth":"2000-01-01","gender":"male","orientations":["straight"]}')
check "POST /api/v1/users (non-.edu rejected)" 400 "$CODE"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
