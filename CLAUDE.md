# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Bytecourses is a web application for managing course proposals. It's built with Go (backend) and uses server-side rendered HTML templates with PostgreSQL for persistence. The project includes both API endpoints (JSON) and server-rendered pages.

## Development Commands

### Building and Running

```bash
# Run the server (memory storage, useful for development)
go run cmd/server/main.go --storage=memory --seed-users=true --bcrypt-cost=5

# Run the server (SQL storage)
go run cmd/server/main.go --storage=sql

# Run database migrations
make migrate  # requires TEST_DATABASE_URL env var
```

### Testing

```bash
# Run all tests
make test

# Run only Go tests
make go-test

# Run only Python e2e tests
make py-test

# Run a single Python test
pytest test/e2e/auth_test.py::test_register -v

# Run a single Go test
go test ./internal/services -run TestAuthService_Register -v
```

### Linting

```bash
# Run all linters (Go and Python)
make lint

# Format and lint Go code
gofmt -w .
go vet ./...

# Format and lint Python code
ruff format .
ruff check .
```

### Setup

```bash
# Install all dev dependencies
make install

# Run full CI pipeline locally
make ci
```

## Architecture

### Layered Architecture

The codebase follows a clean layered architecture with clear separation of concerns:

```
cmd/server/          - Application entry point
internal/app/        - App initialization and router setup
internal/http/       - HTTP layer (handlers, middleware)
internal/services/   - Business logic layer
internal/store/      - Data access layer (repository pattern)
internal/domain/     - Domain models
internal/auth/       - Authentication utilities
internal/notify/     - Email notification system
```

### Layer Responsibilities

**Domain Layer** (`internal/domain/`):
- Core business entities: `User`, `Proposal`
- Domain logic lives on the entities (e.g., `IsViewableBy()`, `IsAmendable()`)
- No dependencies on other layers

**Store Layer** (`internal/store/`):
- Defines repository interfaces: `UserStore`, `ProposalStore`, `PasswordResetStore`
- Two implementations: `memstore` (in-memory) and `sqlstore` (PostgreSQL)
- `sqlstore.DB` implements all store interfaces (single struct, multiple interfaces)
- Pure data access, no business logic

**Services Layer** (`internal/services/`):
- Business logic and orchestration
- Services: `AuthService`, `ProposalService`
- Coordinates between stores, handles validations, executes workflows
- Returns domain errors (defined in `services/errors.go`)
- All services are collected in the `Services` struct created via `services.New()`

**HTTP Layer** (`internal/http/`):
- **Handlers**: Convert HTTP requests to service calls, handle HTTP concerns
  - `AuthHandler`, `ProposalHandler`, `PageHandlers`, `SystemHandlers`
  - API handlers return JSON, page handlers render HTML templates
- **Middleware**: `RequireUser`, `RequireLogin`, `RequireProposal`
  - Middleware injects authenticated users/proposals into request context
  - Access via `handlers.UserFromContext()`, `handlers.ProposalFromContext()`

**App Layer** (`internal/app/`):
- App initialization (`app.New()`)
- Dependency injection and wiring
- Router configuration
- Storage backend selection (memory vs SQL)

### Key Patterns

**Dependency Injection**:
- The `App` struct holds all dependencies
- Services are initialized with explicit `Dependencies` struct
- Handlers receive services via constructor injection

**Context-based Request State**:
- Middleware adds authenticated user/proposal to `context.Context`
- Handlers retrieve via helper functions: `UserFromContext()`, `ProposalFromContext()`

**Interface-based Repositories**:
- Store interfaces in `internal/store/store.go`
- Swap implementations (memory vs SQL) at app startup
- Makes testing straightforward (use memstore in tests)

**Service Error Handling**:
- Services return domain errors (e.g., `ErrInvalidCredentials`, `ErrNotFound`)
- Handlers map service errors to HTTP status codes via `handleServiceError()`

## Testing

### E2E Tests (Python)

E2E tests live in `test/e2e/` and use pytest + requests:
- Tests spin up a Go server per test function with a random port
- The `go_server` fixture handles server lifecycle
- Tests use memory storage (`--storage=memory`) and seed test users
- Lower bcrypt cost (`--bcrypt-cost=5`) speeds up tests

Example fixture usage:
```python
def test_login(go_server):
    r = requests.post(f"{go_server}/login", json={...})
    assert r.status_code == 200
```

### Go Unit Tests

Go tests typically test services and stores:
- Use `memstore` for fast in-memory testing
- Test files live alongside implementation (e.g., `services/auth_test.go`)

## Database

### Migrations

Database migrations use Goose and live in `migrations/`:
- Filenames: `001_description.sql`, `002_description.sql`, etc.
- Run with: `goose -dir migrations postgres "$DATABASE_URL" up`
- The `make migrate` command runs migrations against `TEST_DATABASE_URL`

### Storage Backends

Two storage backends available:
1. **Memory** (`--storage=memory`): Uses `memstore` package, useful for development/testing
2. **SQL** (`--storage=sql`): Uses `sqlstore` package with PostgreSQL via pgx driver

## Configuration

Configuration via environment variables:
- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string (required for `--storage=sql`)
- `TEST_DATABASE_URL`: Test database connection string (used by CI and `make migrate`)
- `ADMIN_EMAIL`, `ADMIN_PASSWORD`: Seeds admin user on startup
- `RESEND_API_KEY`, `RESEND_FROM_EMAIL`: Email configuration for password resets

Command-line flags:
- `--storage`: Choose backend (`memory` or `sql`)
- `--bcrypt-cost`: Set bcrypt cost (lower for tests)
- `--seed-users`: Seed test users (admin@local.bytecourses.org / user@local.bytecourses.org)

## Web Frontend

The app serves both API endpoints (`/api/*`) and server-rendered pages:
- Templates: `web/templates/`
- Static files: `web/static/`
- Page handlers use Go's `html/template` for rendering
- Markdown rendering via `yuin/goldmark` for proposal content
