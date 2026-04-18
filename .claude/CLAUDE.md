# Gasolina API — Project Context for Claude

## Project Overview

**Gasolina API** is a Go REST API that acts as an optional sync backend for the Gasolina Flutter app (`D:\Develop\Gasolina\gasolina`). The app is local-first; this API exists solely to push data to the cloud and pull it back — the primary use case is recovery after accidental local data loss.

- **Language**: Go
- **Database**: PostgreSQL
- **Auth**: Rolling JWT (device secret → token; every response returns a refreshed token)
- **Config**: `config.json` at repo root (no env var fallback)

---

## Tech Stack

| Concern | Library |
|---|---|
| HTTP framework | `github.com/gofiber/fiber/v2` |
| PostgreSQL driver | `github.com/jackc/pgx/v5` |
| JWT | `github.com/golang-jwt/jwt/v5` |

---

## Architecture: DDD + Clean Architecture

This project strictly follows **Domain-Driven Design** and **Clean Architecture**. These are non-negotiable. Every file must respect the layer it belongs to.

### Dependency Rule
Dependencies point **inward only**. Outer layers may import inner layers; inner layers must never import outer layers.

```
interfaces → application → domain
infrastructure → domain
cmd → (everything, for wiring only)
```

### Layer Responsibilities

| Layer | Package path | What belongs here |
|---|---|---|
| **Domain** | `internal/domain/` | Entities, value objects, repository interfaces, domain errors. **Zero external imports.** |
| **Application** | `internal/application/` | Use cases. Orchestrates domain objects. Input/output DTOs per use case. Imports `domain/` only. |
| **Infrastructure** | `internal/infrastructure/` | Repository implementations (PostgreSQL via pgx). Imports `domain/` only. |
| **Interfaces** | `internal/interfaces/` | HTTP handlers, request/response DTOs, JWT middleware. Imports `application/` and `domain/`. |
| **cmd** | `cmd/api/` | Entry point. The only place allowed to import across all layers for wiring. |

### Cross-cutting packages (outside `internal/`)
- `auth/` — JWT token generation and validation. Used by both `interfaces/http/middleware/` and `interfaces/http/handler/auth.go`.
- `config/` — Loads `config.json` into a typed `Config` struct.
- `server/` — Fiber app wiring and middleware stack.

---

## Folder Structure

```
gasolina-api/
├── .claude/CLAUDE.md
├── config.json                                # Gitignored — copy from config.json.example
├── config.json.example
├── cmd/api/main.go
├── internal/
│   ├── domain/fuelentry/
│   │   ├── entity.go          # FuelEntry struct + Validate()
│   │   └── repository.go      # Repository interface (port)
│   ├── application/fuelentry/
│   │   ├── service.go         # Use cases: Add, Update, Delete, GetByID, GetAll, Sync
│   │   └── dto.go             # Input/output DTOs
│   ├── infrastructure/persistence/postgres/
│   │   └── fuel_entry_repository.go
│   └── interfaces/http/
│       ├── handler/
│       │   ├── auth.go
│       │   └── entries.go
│       ├── middleware/jwt.go
│       └── response/response.go
├── auth/jwt.go
├── config/config.go
├── server/server.go
└── migrations/001_initial.sql
```

---

## Configuration (`config.json`)

```json
{
  "database_url":  "postgres://user:pass@localhost:5432/gasolina?sslmode=disable",
  "jwt_secret":    "change-me",
  "device_secret": "change-me",
  "port":          "8080",
  "token_ttl":     "24h"
}
```

`config.json` is the single source of truth. Do not add env var fallback.

---

## Rolling JWT

- `POST /auth/token` exchanges `device_secret` for a JWT
- Every authenticated response returns `X-Refresh-Token: <new_token>` header
- The client always replaces its stored token with the refreshed one
- Token claims: `sub = "gasolina-device"`, `exp`, `iat`

---

## API Surface

| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/token` | Get JWT from device secret |
| `GET` | `/v1/entries` | Pull all (or `?since=<ms>` for delta) |
| `GET` | `/v1/entries/{id}` | Single entry |
| `POST` | `/v1/entries` | Push new entry |
| `PUT` | `/v1/entries/{id}` | Push update (last-write-wins on `updated_at`) |
| `DELETE` | `/v1/entries/{id}` | Push deletion (soft-tracked for sync) |
| `POST` | `/v1/entries/sync` | Bidirectional bulk sync; `last_sync_at: 0` = full recovery |

---

## Hard Rules

### DDD + Clean Architecture (non-negotiable)
- `domain/` has **zero** imports from this project
- Handlers **never** call repositories directly — they go through application services
- Domain validation lives on the entity (`Validate()`) — not duplicated in handlers or services
- Infrastructure details (SQL, pgx) never leak into application or domain layers

### DRY
- HTTP response helpers (`writeJSON`, `writeError`) live in `internal/interfaces/http/response/` only
- Shared domain errors live in `internal/domain/` only

### Code style
- All files start with a 2-line `ABOUTME:` comment
- Evergreen naming — no `new_`, `old_`, `improved_`, `enhanced_` in identifiers
- No business logic in handlers — delegate to the application service

### Testing (TDD)
- Write failing test before implementation
- Domain and application: unit tests (no I/O)
- Repository: integration tests against real PostgreSQL
- Handlers: end-to-end HTTP tests
- `go test ./...` must always be green

---

## Related Project

Flutter app: `D:\Develop\Gasolina\gasolina`
The app's JSON field names (notably `total_cost` as snake_case) must be matched exactly in API request/response bodies.
