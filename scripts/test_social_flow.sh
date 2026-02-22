#!/bin/bash
# Test Social Flow: browse users -> shoot your shot -> mutual match
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

echo "=== Date Drop Social Flow Test ==="
echo ""

# Seed first
echo "1. Seed data"
curl -s -o /dev/null -X POST "$BASE_URL/debug/seed"

# Create two test users
echo "2. Create user A"
A_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"email":"usera@stanford.edu","name":"User A","date_of_birth":"2000-06-15","gender":"female","orientations":["straight"]}')
USER_A=$(echo "$A_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "  User A: $USER_A"

echo "3. Create user B"
B_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"email":"userb@mit.edu","name":"User B","date_of_birth":"2000-03-20","gender":"male","orientations":["straight"]}')
USER_B=$(echo "$B_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "  User B: $USER_B"

# Browse users
echo "4. Browse users"
BROWSE_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $USER_A" \
  "$BASE_URL/api/v1/social/users")
CODE=$(echo "$BROWSE_RESPONSE" | tail -1)
check "GET /api/v1/social/users" 200 "$CODE"

# User A shoots at User B
echo "5. User A shoots at User B"
SHOT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/social/shoot" \
  -H "Authorization: Bearer $USER_A" \
  -H "Content-Type: application/json" \
  -d "{\"target_id\": \"$USER_B\"}")
CODE=$(echo "$SHOT_RESPONSE" | tail -1)
BODY=$(echo "$SHOT_RESPONSE" | head -1)
check "POST /social/shoot (A->B)" 201 "$CODE"
MUTUAL=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['mutual'])" 2>/dev/null)
echo "  Mutual: $MUTUAL"

# User B shoots at User A (should be mutual!)
echo "6. User B shoots at User A (should be mutual)"
SHOT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/social/shoot" \
  -H "Authorization: Bearer $USER_B" \
  -H "Content-Type: application/json" \
  -d "{\"target_id\": \"$USER_A\"}")
CODE=$(echo "$SHOT_RESPONSE" | tail -1)
BODY=$(echo "$SHOT_RESPONSE" | head -1)
check "POST /social/shoot (B->A)" 201 "$CODE"
MUTUAL=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['mutual'])" 2>/dev/null)
echo "  Mutual: $MUTUAL"
if [ "$MUTUAL" = "True" ]; then
  echo "  SUCCESS: Mutual shot detected!"
  PASS=$((PASS + 1))
else
  echo "  FAIL: Expected mutual=True"
  FAIL=$((FAIL + 1))
fi

# Check mutual shots
echo "7. Get mutual shots for User A"
MUTUAL_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $USER_A" \
  "$BASE_URL/api/v1/social/shots/mutual")
CODE=$(echo "$MUTUAL_RESPONSE" | tail -1)
check "GET /social/shots/mutual" 200 "$CODE"

# Try duplicate shot (should conflict)
echo "8. Duplicate shot (should fail)"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/social/shoot" \
  -H "Authorization: Bearer $USER_A" \
  -H "Content-Type: application/json" \
  -d "{\"target_id\": \"$USER_B\"}")
check "POST /social/shoot (duplicate rejected)" 409 "$CODE"

# Self-shot (should fail)
echo "9. Self-shot (should fail)"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/api/v1/social/shoot" \
  -H "Authorization: Bearer $USER_A" \
  -H "Content-Type: application/json" \
  -d "{\"target_id\": \"$USER_A\"}")
check "POST /social/shoot (self rejected)" 400 "$CODE"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
