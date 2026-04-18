# Gasolina API

Optional cloud sync backend for the [Gasolina](https://github.com/amavis442/gasolina) Flutter app. The app is local-first — this API exists solely to push fuel entries to the cloud and pull them back. The primary use case is recovery after accidental local data loss.

---

## What it does

- Issues a JWT to a trusted device via a shared secret
- Exposes CRUD endpoints for fuel entries
- Supports delta sync (`?since=<unix_ms>`) and full recovery (`last_sync_at: 0`)
- Uses last-write-wins on `updated_at` for conflict resolution
- Soft-deletes entries so they propagate correctly during sync

---

## Tech stack

| Concern | Library |
|---|---|
| HTTP framework | [Fiber v2](https://github.com/gofiber/fiber) |
| PostgreSQL driver | [pgx v5](https://github.com/jackc/pgx) |
| JWT | [golang-jwt v5](https://github.com/golang-jwt/jwt) |

---

## Prerequisites

- Go 1.22+
- PostgreSQL 14+

---

## Setup

### 1. Clone and install dependencies

```bash
git clone https://github.com/amavis442/gasolina-api.git
cd gasolina-api
go mod download
```

### 2. Create the database

```sql
CREATE DATABASE gasolina;
```

Then run the migration:

```bash
psql -d gasolina -f migrations/001_initial.sql
```

### 3. Generate secrets

**`jwt_secret`** — signs and verifies JWTs. Use a long random string:

```bash
# Option A: openssl
openssl rand -hex 32

# Option B: Go one-liner
go run -e 'package main; import ("crypto/rand"; "encoding/hex"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b)) }'
```

**`device_secret`** — the password your Flutter app sends to get a token. Any strong random value works:

```bash
openssl rand -hex 24
```

> Keep both secrets out of version control. `config.json` is already in `.gitignore`.

### 4. Configure

Copy the example config and fill in your values:

```bash
cp config.json.example config.json
```

```json
{
  "database_url":  "postgres://user:pass@localhost:5432/gasolina?sslmode=disable",
  "jwt_secret":    "<output of openssl rand -hex 32>",
  "device_secret": "<output of openssl rand -hex 24>",
  "port":          "8080",
  "token_ttl":     "24h"
}
```

### 5. Run

```bash
go run ./cmd/api
```

---

## API

All `/v1/*` endpoints require `Authorization: Bearer <token>`. Every authenticated response returns a refreshed token in the `X-Refresh-Token` header — the client must replace its stored token with this value.

### Authentication

| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/token` | Exchange `device_secret` for a JWT |

**Request:**
```json
{ "device_secret": "your-device-secret" }
```

**Response:**
```json
{ "token": "<jwt>" }
```

### Fuel entries

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/entries` | All entries (add `?since=<unix_ms>` for delta) |
| `GET` | `/v1/entries/{id}` | Single entry |
| `POST` | `/v1/entries` | Create entry |
| `PUT` | `/v1/entries/{id}` | Update entry (last-write-wins on `updated_at`) |
| `DELETE` | `/v1/entries/{id}` | Soft-delete entry |
| `POST` | `/v1/entries/sync` | Bidirectional bulk sync |

**Sync payload:**
```json
{
  "last_sync_at": 0,
  "entries": [ ... ]
}
```
`last_sync_at: 0` triggers a full recovery — the server returns all entries.

---

## Tests

### Unit + integration (Go)

```bash
go test ./...
```

Covers domain validation, application service logic (in-memory mock), and all HTTP endpoints via `fiber.App.Test()` — no real server or database needed.

### E2E smoke tests (Deno)

Requires [Deno](https://deno.com) and a running API server pointed at a real PostgreSQL database.

```bash
# Run the full suite
cd e2e
deno task test

# Run a single file
deno task test:auth
deno task test:entries
deno task test:sync
```

Or from the repo root:

```bash
make e2e
```

The suite reads `config.json` directly — no environment variables needed. Each test file cleans up its own test rows (prefixed `e2e-*` / `sync-e2e-*`) at the start so runs are idempotent.

> Start the server with `make run` (or `go run ./cmd/api`) before running the E2E suite.

---

## Architecture

The project follows DDD + Clean Architecture. Dependencies point inward only:

```
interfaces → application → domain
infrastructure → domain
cmd → (everything, for wiring only)
```

See [.claude/CLAUDE.md](.claude/CLAUDE.md) for the full architectural rules.
