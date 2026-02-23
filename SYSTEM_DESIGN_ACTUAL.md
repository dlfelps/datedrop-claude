# DateDrop — System Design: Actual Implementation
**Version:** 1.0
**Date:** February 2026
**Scope:** Documents the backend as it is actually built and running.

---

## 1. Executive Summary

DateDrop is a backend dating application built in Go. It is a university-focused matchmaking service that delivers time-limited "drops" (match notifications) on a weekly cadence. Rather than unlimited swiping, users complete a 66-question compatibility quiz; an algorithm scores pairwise compatibility and produces one weekly match per user. The platform also exposes two supplemental social mechanics — "Shoot Your Shot" and "Cupid Mode" — that create out-of-band matches without consuming a user's weekly drop slot.

The implementation covers the full feature set described in the original design for Phase 1. It is structured as a single monolithic Go application (not a microservices deployment) with an in-memory data store and a layered, clean-architecture codebase that is designed to swap out the storage layer for a real database with minimal changes.

---

## 2. Architecture Overview

The application is a single binary that starts one HTTP server. All logic runs in-process. There is no message bus, no separate services, no database, and no external infrastructure dependencies beyond the Go runtime.

### 2.1 Architectural Style

**Monolith with layered (clean) architecture.**

The code is organized into four horizontal layers that enforce a strict dependency direction: HTTP → Services → Domain → Repository.

```
┌─────────────────────────────────────┐
│          HTTP Layer (Gin)           │  internal/api/handlers/
│  handlers · middleware · routes     │  internal/api/middleware/
└────────────────┬────────────────────┘
                 │ calls
┌────────────────▼────────────────────┐
│          Service Layer              │  internal/services/
│  business logic · workflows         │
└────────────────┬────────────────────┘
                 │ calls
┌────────────────▼────────────────────┐
│          Domain Layer               │  internal/domain/entities/
│  entities · state machines          │
└────────────────┬────────────────────┘
                 │ calls
┌────────────────▼────────────────────┐
│        Repository Layer             │  internal/repository/
│  interfaces + in-memory impls       │
└─────────────────────────────────────┘
```

### 2.2 Core Design Principles

- **Repository-pattern abstraction:** All data access goes through repository interfaces. The concrete implementation is in-memory but can be replaced with a SQL/NoSQL backend by swapping the struct that satisfies the interface.
- **Dependency injection:** All service and repository dependencies are wired in `cmd/server/main.go`. No global state.
- **Rich domain entities:** Business rules and state-machine logic live on the entity structs, not in the service layer.
- **Thread safety:** In-memory repositories use `sync.RWMutex` to be safe for concurrent requests.
- **Input validation at the boundary:** Email (.edu), age (18+), and field presence are validated at the handler and service layers before any domain mutation occurs.

### 2.3 Technology Stack

| Concern | Technology |
|---|---|
| Language | Go 1.21 |
| HTTP framework | Gin v1.9.1 |
| ID generation | github.com/google/uuid v1.5.0 |
| Data storage | In-memory maps (no external DB) |
| Authentication | Mock Bearer token (user_id as token) |
| Notifications | Console logging (no external push/email) |
| Build | GNU Make + `go build` |
| Tests | Go testing + integration tests + shell scripts |
| Deployment | Single binary (`bin/server`), port 8080 |

---

## 3. Service / Module Definitions

Because the system is a monolith the "services" are Go structs with methods, not network-addressable processes. Each service below maps to a file in `internal/services/`.

### 3.1 User Service (`user_service.go`)

Handles user lifecycle.

**Responsibilities:**
- Register a new user (validates .edu email, age ≥ 18, deduplicates email)
- Mock login (returns user record; no password check)
- Fetch user profile by ID
- Update mutable profile fields (name, bio, photos)
- Mark a user's quiz as complete

**Key data model (Go struct `User`):**

| Field | Type | Notes |
|---|---|---|
| ID | UUID string | Primary key |
| Email | string | .edu validated, unique |
| Name | string | Display name |
| DateOfBirth | time.Time | Age gate ≥ 18 |
| Gender | Enum | Male / Female / NonBinary |
| Orientation | Enum[] | Straight / Gay / Bisexual |
| Bio | string | |
| Photos | []string | URLs |
| QuizCompleted | bool | Gate for matching eligibility |
| IsActive | bool | Soft visibility flag |
| CreatedAt | time.Time | |

**Endpoints:**
```
POST   /api/v1/users            Register
POST   /api/v1/auth/login       Login (mock)
GET    /api/v1/users/:id        Get profile
PATCH  /api/v1/users/:id        Update profile
```

### 3.2 Quiz Service (`quiz_service.go`)

Manages the 66-question compatibility quiz.

**Responsibilities:**
- Seed and store all 66 questions on startup
- Return the full question set to authenticated users
- Accept individual question responses with an importance weight
- Track per-user quiz completion (answered / 66)
- Mark quiz complete when all questions answered

**Quiz structure:**

| Domain | Count | Config Weight |
|---|---|---|
| Lifestyle | 22 | 35% |
| Values | 22 | 40% |
| Politics | 22 | 25% |

Response types supported: `scale_5`, `scale_7`, `multiple_choice`, `boolean`.

Each response carries an `importance` field (1–5) that amplifies or dampens the pairwise score.

**Key data models:** `Question` (id, domain, text, response_type, options), `QuizResponse` (user_id, question_id, answer, importance).

**Endpoints:**
```
GET  /api/v1/quiz/questions          All questions
POST /api/v1/quiz/responses          Submit one response
GET  /api/v1/quiz/status/:user_id    Completion status
```

### 3.3 Matching Service (`matching_service.go`)

The compatibility scoring and weekly batch match engine.

**Responsibilities:**
- On demand (not scheduled): run the weekly matching job when `POST /api/v1/matching/run` is called
- Compute pairwise compatibility scores for all eligible user pairs
- Apply exclusion filters (orientation/gender, prior matches within 12-week lookback, blocks)
- Assign matches with a greedy bipartite algorithm (highest-score pair matched first)
- Create `Drop` records of type `Weekly` for each matched pair, auto-revealed

**Scoring algorithm (`pkg/utils/compatibility.go` + `matching_service.go`):**

1. **Per-question alignment:**
   - Scale (1–5 or 1–7): `1.0 - |a - b| / (maxScale - 1)` → normalized to [0,1]
   - Multiple choice / boolean: `1.0` if exact match, else `0.0`

2. **Importance weighting:** Each answer carries an importance weight (1–5).
   - `weight = importance / 5.0`
   - If raw score < 0.5: `weightedScore = rawScore * (1 - weight * 0.5)` — penalizes mismatch on important questions
   - If raw score ≥ 0.5: `weightedScore = rawScore * (0.5 + weight * 0.5)` — boosts agreement on important questions

3. **Domain aggregation:** `domainScore = mean(weightedScores for all questions in domain)`

4. **Final score:** `Σ(domainScore × domainWeight)`, clamped to [0.0, 1.0]

**Matching algorithm:** Greedy bipartite matching. All valid pairs scored and sorted descending; iterate top-to-bottom, match each user once to their highest-scoring available partner.

**Endpoints:**
```
POST /api/v1/matching/run    Trigger weekly batch (admin/protected)
```

### 3.4 Drop Service (`drop_service.go`)

Manages the lifecycle of every match record.

**Drop sources:** `Weekly` (from matching job), `Shot` (mutual Shoot Your Shot), `Cupid` (accepted Cupid nomination).

**State machine:**
```
Pending ──► Revealed ──► PendingMutual ──► Matched ──► Completed
                │               │
                ▼               ▼
            Declined        Declined
                │               │
                └───────────────┘
                        │
                    Expired (on 72-hr timeout)

Matched / Completed ──► Archived
```

**State transition rules (enforced by `Drop.CanTransitionTo()`):**
- `Pending → Revealed`: weekly drops auto-reveal on creation
- `Revealed → PendingMutual`: first user accepts
- `PendingMutual → Matched`: second user accepts
- Any non-terminal state → `Declined` (explicit decline) or `Expired` (72-hr expiry checked on access)
- `Matched/Completed → Archived`

**Key data model (`Drop`):**

| Field | Type |
|---|---|
| ID | UUID |
| UserAID / UserBID | UUID |
| CompatibilityScore | float64 |
| Type | Enum: Weekly / Shot / Cupid |
| Status | Enum (state machine above) |
| UserAAccepted / UserBAccepted | bool |
| WeekOf | time.Time |
| RevealedAt | time.Time |
| ExpiresAt | time.Time (72h after reveal) |
| CreatedAt | time.Time |

**Endpoints:**
```
GET  /api/v1/drops/current       Active drop for authenticated user
POST /api/v1/drops/:id/accept    Accept drop
POST /api/v1/drops/:id/decline   Decline drop
GET  /api/v1/drops/history       Paginated history
```

### 3.5 Social Service (`social_service.go`)

Implements Shoot Your Shot and Cupid Mode.

**Shoot Your Shot:**
- Any authenticated user can fire a shot at any other active, quiz-complete user (not self)
- Shots are stored silently; the target is never notified of a one-way shot
- When a reverse shot is detected (B fires at A and A already fired at B), both users are notified and a `Shot` drop is created immediately

**Cupid Mode:**
- A nominator picks two other users
- The matching engine's compatibility score is pre-computed for the nominee pair
- Both nominees are notified and must independently accept
- When both accept, a `Cupid` drop is created
- Status: `Pending → User1Accept / User2Accept → Accepted` (both accepted) or `Declined`

**Endpoints:**
```
GET  /api/v1/social/users            Browse users (paginated)
POST /api/v1/social/shoot            Fire a shot
GET  /api/v1/social/shots/mutual     List mutual shots
POST /api/v1/social/cupid            Nominate two users
POST /api/v1/social/cupid/:id/accept Accept nomination
POST /api/v1/social/cupid/:id/decline Decline nomination
```

### 3.6 Notification Service (`notification_service.go`)

Mock implementation. All notifications are written to `stdout` via `fmt.Println`. No external channels (no push, no email, no WebSocket).

**Notification events emitted:**
- New drop received
- Drop expiring soon
- Match confirmed
- Mutual shot detected
- Cupid nomination received
- Drop expired
- User blocked/reported

### 3.7 Moderation Service (`moderation_service.go`)

**Responsibilities:**
- Block/unblock a user (blocks are checked by the matching engine before pairing)
- Report a user with a category and optional description

**Report categories:** Harassment, Inappropriate, Spam, FakeProfile, Other
**Report statuses:** Open → Reviewed → Resolved

**Endpoints:**
```
POST   /api/v1/moderation/block/:user_id    Block
DELETE /api/v1/moderation/block/:user_id    Unblock
POST   /api/v1/moderation/report            Report
```

> **Note:** There is no admin dashboard or `GET /moderation/reports` endpoint in the implementation. Report data is stored but not exposed via API.

---

## 4. Data Architecture

### 4.1 Storage

All data is stored in-process in Go maps protected by `sync.RWMutex`. There is no external database, no cache layer, and no persistence; all data is lost on server restart.

| Repository | Backing structure |
|---|---|
| UserRepository | `map[string]*User` |
| QuestionRepository | `map[string]*Question` |
| ResponseRepository | `map[string]*QuizResponse` |
| DropRepository | `map[string]*Drop` |
| ShotRepository | `map[string]*Shot` |
| CupidRepository | `map[string]*CupidNomination` |
| ModerationRepository | `map[string]*Block`, `map[string]*Report` |

### 4.2 Repository Interfaces

Every repository is defined as a Go interface in `internal/repository/interfaces.go`. Concrete in-memory structs in `internal/repository/memory/` implement those interfaces. This allows the storage layer to be replaced (e.g., with PostgreSQL via `database/sql`) without changing any service code.

### 4.3 Data Sensitivity

There is no encryption at rest. Quiz responses and user data live in plain Go structs in memory. No audit trail exists beyond what is in the running process logs.

---

## 5. API Layer

### 5.1 Authentication

Mock Bearer token scheme. The client sends:
```
Authorization: Bearer <user-uuid>
```
The middleware (`internal/api/middleware/auth.go`) extracts the UUID and looks up the user. If the user exists and is active, the user is stored in the Gin context. There are no JWTs, no refresh tokens, and no password validation.

### 5.2 Public vs Protected Routes

| Route | Auth Required |
|---|---|
| `POST /api/v1/users` | No |
| `POST /api/v1/auth/login` | No |
| `GET /health` | No |
| All other `/api/v1/*` routes | Yes |
| `POST /debug/seed` | No (debug only) |

### 5.3 Error Mapping

Services return typed Go errors. Handlers map these to HTTP status codes:

| Scenario | Status |
|---|---|
| Invalid input / validation failure | 400 |
| Missing/invalid auth token | 401 |
| Acting on another user's resource | 403 |
| Resource not found | 404 |
| Duplicate email | 409 |
| Expired drop | 410 |
| Unexpected error | 500 |

### 5.4 Full Endpoint List

```
# Public
GET    /health
POST   /api/v1/users
POST   /api/v1/auth/login

# Users
GET    /api/v1/users/:id
PATCH  /api/v1/users/:id

# Quiz
GET    /api/v1/quiz/questions
POST   /api/v1/quiz/responses
GET    /api/v1/quiz/status/:user_id

# Matching
POST   /api/v1/matching/run

# Drops
GET    /api/v1/drops/current
POST   /api/v1/drops/:id/accept
POST   /api/v1/drops/:id/decline
GET    /api/v1/drops/history

# Social
GET    /api/v1/social/users
POST   /api/v1/social/shoot
GET    /api/v1/social/shots/mutual
POST   /api/v1/social/cupid
POST   /api/v1/social/cupid/:id/accept
POST   /api/v1/social/cupid/:id/decline

# Moderation
POST   /api/v1/moderation/block/:user_id
DELETE /api/v1/moderation/block/:user_id
POST   /api/v1/moderation/report

# Debug
POST   /debug/seed
```

---

## 6. Configuration

Defined in `internal/config/config.go`. No environment variable loading; all defaults are compile-time constants.

| Setting | Value |
|---|---|
| Server port | `:8080` |
| Read/Write timeout | 10s |
| Total quiz questions | 66 |
| Minimum age | 18 |
| Drop expiration | 72 hours |
| Lifestyle domain weight | 0.35 |
| Values domain weight | 0.40 |
| Politics domain weight | 0.25 |
| Matching lookback | 12 weeks |

---

## 7. Compatibility Logic: Gender / Orientation

User compatibility is checked before any match is created (weekly matching and Cupid pre-score).

```
A is compatible with B if:
  A is attracted to B's gender (per A's orientation(s))
  AND
  B is attracted to A's gender (per B's orientation(s))

Straight:    attracted to opposite gender; non-binary compatible
Gay:         attracted to same gender; non-binary compatible
Bisexual:    attracted to all genders
Non-binary:  compatible with all orientations
```

---

## 8. Testing

### 8.1 Integration Tests

`internal/api/integration_test.go` covers:
- Health check
- User creation (valid + invalid email/age)
- Login
- Profile retrieval and update
- Quiz question retrieval and response submission
- Quiz completion gating

### 8.2 Shell Test Scripts

| Script | Covers |
|---|---|
| `scripts/test_weekly_flow.sh` | Full register → quiz → match → accept cycle |
| `scripts/test_social_flow.sh` | Shoot Your Shot mutual-match flow |
| `scripts/test_cupid_flow.sh` | Cupid nomination and acceptance flow |

---

## 9. Build and Run

```bash
make build   # go build → bin/server
make run     # go run ./cmd/server
make test    # go test ./... -v
make clean   # rm bin/
```

---

## 10. Known Limitations of Current Implementation

1. **No persistence.** All data is in-memory and lost on restart.
2. **No real authentication.** Bearer token is the user's UUID — no password, no JWT, no expiry.
3. **No messaging.** Real-time chat is not implemented; the Messaging Service described in the design doc has no equivalent here.
4. **No notifications.** All notifications are `fmt.Println` console logs.
5. **No photo upload.** Photos are stored as strings (URLs) with no upload endpoint or object storage integration.
6. **No scheduling.** The weekly matching job is triggered manually via `POST /api/v1/matching/run`, not via a cron or scheduler.
7. **No admin API.** There are no admin endpoints to list/review moderation reports or take moderation actions.
8. **No rate limiting.** No per-user or per-IP throttling at any layer.
9. **No encryption.** Quiz responses and all user data are stored in plain memory; no AES-256 or any other at-rest encryption.
10. **Matching is O(n²).** No vectorized scoring or caching; acceptable for Phase 1 scale but not for 50,000+ users.
