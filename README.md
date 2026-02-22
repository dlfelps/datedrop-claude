# DateDrop

A dating app backend service built with Go and Gin. DateDrop matches users through a compatibility quiz, delivers time-limited "drops" (match notifications), and supports social features like "Shoot Your Shot" and Cupid nominations.

## Tech Stack

- **Language:** Go 1.21
- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **Storage:** In-memory repositories (swappable via repository interfaces)

## Project Structure

```
cmd/server/          - Application entrypoint
internal/
  api/
    handlers/        - HTTP request handlers
    middleware/       - Auth middleware
    routes.go        - Route definitions
  config/            - App configuration
  domain/entities/   - Domain models (User, Drop, Shot, Cupid, etc.)
  repository/        - Repository interfaces and in-memory implementations
  services/          - Business logic (matching, quiz, social, moderation)
pkg/utils/           - Shared utilities (validation, ID generation, compatibility)
```

## Getting Started

### Prerequisites

- Go 1.21+

### Build and Run

```bash
# Build
make build

# Run
make run
```

The server starts on `:8080` by default.

### Run Tests

```bash
make test
```

## API

All endpoints are under `/api/v1`. Public routes:

| Method | Path | Description |
|--------|------|-------------|
| POST | `/users` | Create a new user |
| POST | `/auth/login` | Log in |

Authenticated routes (require auth header):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/users/:id` | Get user profile |
| PATCH | `/users/:id` | Update user profile |
| GET | `/quiz/questions` | Get quiz questions |
| POST | `/quiz/responses` | Submit a quiz response |
| GET | `/quiz/status/:user_id` | Get quiz completion status |
| POST | `/matching/run` | Run the matching algorithm |
| GET | `/drops/current` | Get current drop |
| POST | `/drops/:id/accept` | Accept a drop |
| POST | `/drops/:id/decline` | Decline a drop |
| GET | `/drops/history` | Get drop history |
| GET | `/social/users` | Browse users |
| POST | `/social/shoot` | Shoot your shot |
| GET | `/social/shots/mutual` | Get mutual shots |
| POST | `/social/cupid` | Nominate a cupid |
| POST | `/social/cupid/:id/accept` | Accept cupid nomination |
| POST | `/social/cupid/:id/decline` | Decline cupid nomination |
| POST | `/moderation/block/:user_id` | Block a user |
| DELETE | `/moderation/block/:user_id` | Unblock a user |
| POST | `/moderation/report` | Report a user |

A health check is available at `GET /health`.

## Configuration

Default configuration (set in `internal/config/config.go`):

| Setting | Default |
|---------|---------|
| Server port | `:8080` |
| Quiz questions | 66 |
| Minimum age | 18 |
| Drop expiration | 72 hours |
| Matching lookback | 12 weeks |
