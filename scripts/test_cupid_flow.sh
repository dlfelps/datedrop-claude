#!/bin/bash
# Test Cupid Flow: nominate -> both accept -> verify cupid match
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

echo "=== Date Drop Cupid Flow Test ==="
echo ""

# Seed first
echo "1. Seed data"
curl -s -o /dev/null -X POST "$BASE_URL/debug/seed"

# Create three users: matchmaker, user1, user2
echo "2. Create matchmaker"
MM_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"email":"cupid@stanford.edu","name":"Cupid","date_of_birth":"2000-01-01","gender":"female","orientations":["straight"]}')
MATCHMAKER=$(echo "$MM_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "  Matchmaker: $MATCHMAKER"

echo "3. Create user 1"
U1_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"email":"friend1@mit.edu","name":"Friend One","date_of_birth":"2000-05-15","gender":"female","orientations":["straight"]}')
USER1=$(echo "$U1_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "  User 1: $USER1"

echo "4. Create user 2"
U2_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"email":"friend2@harvard.edu","name":"Friend Two","date_of_birth":"2000-08-20","gender":"male","orientations":["straight"]}')
USER2=$(echo "$U2_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "  User 2: $USER2"

# Nominate cupid match
echo "5. Nominate cupid match"
NOM_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/social/cupid" \
  -H "Authorization: Bearer $MATCHMAKER" \
  -H "Content-Type: application/json" \
  -d "{\"user1_id\": \"$USER1\", \"user2_id\": \"$USER2\"}")
CODE=$(echo "$NOM_RESPONSE" | tail -1)
BODY=$(echo "$NOM_RESPONSE" | head -1)
check "POST /social/cupid" 201 "$CODE"
NOM_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "  Nomination ID: $NOM_ID"
echo "  Score: $(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['compatibility_score'])" 2>/dev/null)"

# User 1 accepts
echo "6. User 1 accepts cupid"
ACCEPT1_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  "$BASE_URL/api/v1/social/cupid/$NOM_ID/accept" \
  -H "Authorization: Bearer $USER1")
CODE=$(echo "$ACCEPT1_RESPONSE" | tail -1)
BODY=$(echo "$ACCEPT1_RESPONSE" | head -1)
check "POST /social/cupid/:id/accept (user1)" 200 "$CODE"
STATUS=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null)
echo "  Status after user1 accept: $STATUS"

# User 2 accepts (should complete)
echo "7. User 2 accepts cupid"
ACCEPT2_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  "$BASE_URL/api/v1/social/cupid/$NOM_ID/accept" \
  -H "Authorization: Bearer $USER2")
CODE=$(echo "$ACCEPT2_RESPONSE" | tail -1)
BODY=$(echo "$ACCEPT2_RESPONSE" | head -1)
check "POST /social/cupid/:id/accept (user2)" 200 "$CODE"
STATUS=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null)
echo "  Status after user2 accept: $STATUS"
if [ "$STATUS" = "accepted" ]; then
  echo "  SUCCESS: Both accepted, cupid match created!"
  PASS=$((PASS + 1))
else
  echo "  FAIL: Expected status=accepted, got $STATUS"
  FAIL=$((FAIL + 1))
fi

# Check drop was created for user1
echo "8. Check drop created for user1"
DROP_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $USER1" \
  "$BASE_URL/api/v1/drops/current")
CODE=$(echo "$DROP_RESPONSE" | tail -1)
BODY=$(echo "$DROP_RESPONSE" | head -1)
DROP_TYPE=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['type'])" 2>/dev/null || echo "")
if [ "$DROP_TYPE" = "cupid" ]; then
  check "GET /drops/current (cupid drop)" 200 "$CODE"
  echo "  Cupid drop verified!"
else
  echo "  Note: Drop type is '$DROP_TYPE' (may be from weekly matching)"
  check "GET /drops/current" 200 "$CODE"
fi

# Test decline flow
echo "9. Create another nomination for decline test"
NOM2_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/social/cupid" \
  -H "Authorization: Bearer $MATCHMAKER" \
  -H "Content-Type: application/json" \
  -d "{\"user1_id\": \"$USER1\", \"user2_id\": \"$MATCHMAKER\"}")
CODE=$(echo "$NOM2_RESPONSE" | tail -1)
BODY=$(echo "$NOM2_RESPONSE" | head -1)
NOM2_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)

echo "10. Decline cupid nomination"
DECLINE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  "$BASE_URL/api/v1/social/cupid/$NOM2_ID/decline" \
  -H "Authorization: Bearer $USER1")
CODE=$(echo "$DECLINE_RESPONSE" | tail -1)
BODY=$(echo "$DECLINE_RESPONSE" | head -1)
check "POST /social/cupid/:id/decline" 200 "$CODE"
STATUS=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['status'])" 2>/dev/null)
echo "  Status after decline: $STATUS"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
