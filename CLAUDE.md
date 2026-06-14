# monsee — Monitoring Platform

Production-grade uptime monitoring system with a public status page and admin dashboard.

---

## Stack

| Layer | Technology |
|---|---|
| Backend | Go + Fiber v3 + Clean Architecture |
| Frontend | Next.js 15 + shadcn/ui + React Query + Zustand (`frontend/` folder in this repo) |
| Database | PostgreSQL 16 + sqlc + pgx/v5 |
| Queue | Redis 7 + Asynq |
| Observability | Zap (structured logs) + Prometheus + OpenTelemetry |
| Security | JWT (httpOnly cookie) + API keys + AES-256-GCM + rate limiting + audit log |
| Infrastructure | Docker Compose — `compose.yml` (dev) + `compose.prod.yml` (prod), both currently empty |

Go module: `github.com/majcek210/monsee`

---

## Architecture — Clean Layers (Strict Dependency Rule)

Outer layers depend on inner layers. Never the reverse. If you import `fiber` inside a service, that is a layer violation.

```
HTTP Request
     |
Handler  (internal/handler/)       -- knows: fiber, service interfaces
     |
Service  (internal/service/)       -- knows: domain, repository interfaces
     |
Repository  (internal/repository/) -- knows: domain, sqlc/pgx
     |
PostgreSQL
```

### Layer Responsibilities

| Layer | Path | Rule |
|---|---|---|
| Domain | `internal/domain/` | Pure Go structs + interfaces. Zero framework imports. |
| Repository | `internal/repository/postgres/` | Implements domain interfaces. Uses sqlc. Swappable with mocks. |
| Service | `internal/service/` | Business logic only. No HTTP, no raw SQL. |
| Handler | `internal/handler/` | Fiber handlers. Translates HTTP <-> domain. No business logic. |

---

## Full Project Structure

```
status-monitor/
├── compose.yml                   -- dev (volume mounts, hot reload, all services)
├── compose.prod.yml              -- prod (resource limits, restart policies, log rotation)
├── .env                          -- never commit
├── .env.example
├── Makefile
│
├── backend/
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   ├── cmd/
│   │   ├── server/main.go        -- HTTP server entrypoint (runs migrations then starts Fiber)
│   │   ├── worker/main.go        -- Asynq worker entrypoint
│   │   ├── migrate/main.go       -- golang-migrate CLI helper
│   │   └── create-admin/main.go  -- seed first admin user
│   │
│   ├── internal/
│   │   ├── domain/               -- Layer 1: pure business types + interfaces
│   │   │   ├── monitor.go
│   │   │   ├── service.go
│   │   │   ├── incident.go
│   │   │   ├── apikey.go
│   │   │   ├── user.go
│   │   │   └── errors.go         -- sentinel errors (ErrNotFound, ErrUnauthorized...)
│   │   │
│   │   ├── repository/           -- Layer 2: DB access
│   │   │   ├── postgres/
│   │   │   │   ├── monitor.go
│   │   │   │   ├── service.go
│   │   │   │   ├── incident.go
│   │   │   │   ├── check_result.go
│   │   │   │   ├── apikey.go
│   │   │   │   └── user.go
│   │   │   └── redis/
│   │   │       └── ratelimit.go
│   │   │
│   │   ├── service/              -- Layer 3: business logic
│   │   │   ├── monitor.go
│   │   │   ├── incident.go
│   │   │   ├── checker.go        -- RunCheck, handleOutage, handleRecovery
│   │   │   ├── apikey.go
│   │   │   └── user.go
│   │   │
│   │   ├── handler/              -- Layer 4: HTTP (Fiber)
│   │   │   ├── router.go         -- wire routes + error handler
│   │   │   ├── monitor.go
│   │   │   ├── service.go
│   │   │   ├── incident.go
│   │   │   ├── apikey.go
│   │   │   ├── user.go
│   │   │   ├── notification.go
│   │   │   ├── webhook.go        -- CRUD + delivery logs
│   │   │   └── v1/               -- public REST API (API key auth)
│   │   │       ├── status.go
│   │   │       └── incidents.go
│   │   │
│   │   ├── checks/               -- check implementations (pure functions, no side effects)
│   │   │   ├── http.go           -- DONE
│   │   │   ├── tcp.go
│   │   │   └── runner.go         -- dispatches to http/tcp based on monitor.Type
│   │   │
│   │   ├── worker/               -- Asynq task definitions + handlers
│   │   │   ├── tasks.go          -- task type constants
│   │   │   ├── check_monitor.go  -- handler for monitor:check task
│   │   │   └── scheduler.go      -- polls ListDue every 15s, enqueues tasks
│   │   │
│   │   ├── middleware/
│   │   │   ├── auth.go           -- JWT session middleware (httpOnly cookie)
│   │   │   ├── apikey.go         -- API key middleware (Bearer token)
│   │   │   ├── ratelimit.go      -- Redis sliding window
│   │   │   └── audit.go          -- writes to audit_log table on writes
│   │   │
│   │   ├── notifications/        -- alert channels (monitor down/recovered)
│   │   │   ├── dispatcher.go     -- routes to discord / email
│   │   │   ├── discord.go        -- POST to Discord webhook URL
│   │   │   └── email.go          -- SMTP email alert
│   │   │
│   │   ├── webhooks/             -- generic outgoing HTTP callbacks
│   │   │   ├── dispatcher.go     -- fires on any platform event
│   │   │   └── delivery.go       -- HTTP POST + retry + log to webhook_logs
│   │   │
│   │   ├── config/
│   │   │   └── config.go         -- typed config loaded from env
│   │   │
│   │   └── telemetry/
│   │       ├── logger.go         -- Zap setup (JSON prod, console dev)
│   │       ├── metrics.go        -- Prometheus metrics registry
│   │       └── tracing.go        -- OpenTelemetry tracer setup
│   │
│   ├── db/
│   │   ├── migrations/           -- numbered .up.sql / .down.sql files
│   │   ├── queries/              -- .sql files read by sqlc
│   │   └── sqlc/                 -- generated Go code (NEVER edit by hand)
│   │
│   ├── pkg/
│   │   ├── encrypt/              -- AES-256-GCM helpers
│   │   └── hash/                 -- bcrypt + SHA-256
│   │
│   └── lib/                      -- existing util helpers (StrPtr, ParseUUID, etc.)
│
└── frontend/                     -- Next.js 15 app (dark, zinc/violet theme)
    ├── app/
    │   ├── layout.tsx            -- always dark: <html class="dark">
    │   ├── page.tsx              -- public status page
    │   ├── login/
    │   └── admin/
    ├── components/
    │   ├── ui/                   -- shadcn components
    │   └── admin/                -- page-specific components
    ├── lib/
    │   ├── api/                  -- typed fetch wrappers per resource
    │   ├── hooks/                -- React Query hooks
    │   └── store/                -- Zustand stores
    └── types/                    -- shared TypeScript types
```

---

## Database Schema

### Existing Migrations (000001–000004)

**000001 — users**
```sql
CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  role          TEXT NOT NULL DEFAULT 'viewer',  -- 'admin' | 'viewer'
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at   TIMESTAMPTZ
);
```

**000002 — services**
```sql
CREATE TABLE services (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  description TEXT,
  status      TEXT NOT NULL DEFAULT 'operational',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at TIMESTAMPTZ
);
```

**000003 — monitors**
```sql
CREATE TABLE monitors (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service_id            UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  name                  TEXT NOT NULL,
  type                  TEXT NOT NULL,              -- 'http' | 'tcp'
  url                   TEXT,
  host                  TEXT,
  port                  INT,
  interval_seconds      INT NOT NULL DEFAULT 60,
  timeout_ms            INT NOT NULL DEFAULT 5000,
  retry_count           INT NOT NULL DEFAULT 2,
  consecutive_failures  INT NOT NULL DEFAULT 0,
  degraded_threshold_ms INT,
  http_method           TEXT DEFAULT 'GET',
  http_expected_status  INT DEFAULT 200,
  enabled               BOOL NOT NULL DEFAULT true,
  next_check_at         TIMESTAMPTZ DEFAULT now(),
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at           TIMESTAMPTZ
);
```

**000004 — check_results**
```sql
CREATE TABLE check_results (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  monitor_id       UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
  status           TEXT NOT NULL,   -- 'up' | 'down' | 'degraded'
  response_time_ms INT,
  error            TEXT,
  checked_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON check_results(monitor_id, checked_at DESC);
```

### Still Needed (future migrations)

**000005 — incidents**
```sql
CREATE TABLE incidents (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service_id   UUID NOT NULL REFERENCES services(id),
  monitor_id   UUID REFERENCES monitors(id),
  title        TEXT NOT NULL,
  severity     TEXT NOT NULL DEFAULT 'high',   -- 'low' | 'medium' | 'high'
  status       TEXT NOT NULL DEFAULT 'open',   -- 'open' | 'resolved'
  resolved_at  TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON incidents(service_id);
CREATE INDEX ON incidents(monitor_id);
CREATE INDEX ON incidents(status);
```

**000006 — api_keys**
```sql
CREATE TABLE api_keys (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id),
  name        TEXT NOT NULL,
  key_hash    TEXT NOT NULL UNIQUE,   -- SHA-256 of sk_... key
  prefix      TEXT NOT NULL,          -- first 8 chars for display
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_used   TIMESTAMPTZ,
  archived_at TIMESTAMPTZ
);
```

**000007 — notification_channels**
```sql
-- Alert channels: Discord + email. Fire when monitor goes down or recovers.
CREATE TABLE notification_channels (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  type        TEXT NOT NULL,          -- 'discord' | 'email'
  config      TEXT NOT NULL,          -- AES-256-GCM encrypted JSON
  enabled     BOOL NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at TIMESTAMPTZ
);
```

**000008 — webhooks**
```sql
-- Generic outgoing HTTP callbacks. Users configure a URL + optional secret.
-- Fires on any platform event (monitor.down, monitor.recovered, incident.created, etc.)
CREATE TABLE webhooks (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  url         TEXT NOT NULL,          -- AES-256-GCM encrypted
  secret      TEXT,                   -- AES-256-GCM encrypted, used for HMAC-SHA256 signature header
  events      TEXT[] NOT NULL DEFAULT '{}',  -- e.g. {'monitor.down','incident.created'}
  enabled     BOOL NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  archived_at TIMESTAMPTZ
);

CREATE TABLE webhook_logs (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  webhook_id   UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
  event        TEXT NOT NULL,
  status_code  INT,
  error        TEXT,
  duration_ms  INT,
  delivered_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON webhook_logs(webhook_id, delivered_at DESC);
```

**000009 — audit_log**
```sql
-- Every admin write is logged here. Never log decrypted credential values.
CREATE TABLE audit_log (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID REFERENCES users(id),
  action      TEXT NOT NULL,      -- 'create' | 'update' | 'archive' | 'delete'
  resource    TEXT NOT NULL,      -- 'monitor' | 'service' | 'incident' ...
  resource_id TEXT,
  ip          TEXT,
  user_agent  TEXT,
  diff        JSONB,              -- field names only, NEVER decrypted credential values
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON audit_log(user_id);
CREATE INDEX ON audit_log(resource, resource_id);
CREATE INDEX ON audit_log(created_at DESC);
```

---

## Migrations — golang-migrate

Migrations run automatically on server startup via `cmd/server/main.go` using `golang-migrate/migrate` with the `pgx/v5` driver.

```go
// Runs before Fiber starts
m, err := migrate.New("file://db/migrations", cfg.DatabaseURL)
m.Up() // applies all pending .up.sql files in order
```

- Files: `NNNNNN_name.up.sql` / `NNNNNN_name.down.sql`
- Never edit a migration that has already been applied — add a new one
- `make migrate-new` creates a timestamped pair

---

## Domain Layer — Key Types

### Errors (internal/domain/errors.go)
```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")
    ErrConflict     = errors.New("conflict")
    ErrValidation   = errors.New("validation error")
    ErrArchivedOnly = errors.New("resource must be archived before deletion")
)

type AppError struct {
    Sentinel error
    Message  string
    Field    string
}
func (e *AppError) Unwrap() error { return e.Sentinel }

// Constructors: domain.NotFound("monitor not found"), domain.ValidationErr("url", "invalid URL")
```

### Error → HTTP mapping (internal/handler/router.go)
```go
map[error]int{
    domain.ErrNotFound:     404,
    domain.ErrUnauthorized: 401,
    domain.ErrForbidden:    403,
    domain.ErrConflict:     409,
    domain.ErrValidation:   422,
}
// Unknown errors → 500, log full details with Zap, return generic message
```

---

## sqlc Configuration (backend/sqlc.yaml)

```yaml
version: '2'
sql:
  - engine: postgresql
    queries: db/queries/
    schema: db/migrations/
    gen:
      go:
        package: sqlcdb
        out: db/sqlc
        emit_json_tags: true
        emit_interface: true
        emit_pointers_for_null_types: true
        emit_empty_slices: true
```

Run `cd backend && sqlc generate` after adding or changing any `.sql` query file.
Generated files in `db/sqlc/` are **never edited by hand**.

---

## Asynq Worker Queue

### Task Constants (internal/worker/tasks.go)
```go
const (
    TaskRunMonitorCheck   = "monitor:check"
    TaskCleanupOldResults = "cleanup:results"
)

type MonitorCheckPayload struct {
    MonitorID string `json:"monitor_id"`
}
```

### Scheduler Pattern
- Ticks every **15 seconds**
- Calls `monitorRepo.ListDue()` — monitors where `enabled=true AND next_check_at <= now()`
- Enqueues each with `asynq.TaskID("check:" + m.ID)` — deduplication key prevents double-enqueue
- `asynq.ErrTaskIDConflict` is expected and ignored
- MaxRetry: 3, Timeout: 30s per task

### Worker Config
- Concurrency: 10
- Queues: `monitors` (priority 8), `cleanup` (priority 1)

### Asynqmon
Sidecar on port **8081** — queue depth, failed jobs, retry history.

---

## Core Check Logic (internal/service/checker.go)

`RunCheck(ctx, monitorID)` — single entry point for all check logic:

1. Fetch monitor from repo
2. Start OTel span + Prometheus timer
3. Run `checks.Run(ctx, monitor)` — dispatches to http.go or tcp.go
4. Insert result into `check_results`
5. If **down**: increment `consecutive_failures`; if `>= retry_count` → `handleOutage`
6. If **up**: reset `consecutive_failures` → `handleRecovery`
7. `SetNextCheckAt` — always runs, even on failure

### handleOutage
- Idempotent: `GetOpenForMonitor` first — skip if incident already open
- Create incident, fire both notifications (discord/email) AND matching webhooks

### handleRecovery
- Find open incident → set `status = resolved`, `resolved_at = now()`
- Fire both notifications AND matching webhooks

---

## Notifications vs Webhooks

**Notifications** (`internal/notifications/`) — platform-managed alert channels:
- Types: `discord` | `email`
- Triggered by: `monitor.down`, `monitor.recovered`
- Config stored encrypted in `notification_channels` table
- Discord: POST embed to webhook URL
- Email: SMTP with configurable from/to

**Webhooks** (`internal/webhooks/`) — user-configured generic HTTP callbacks:
- User provides a URL + optional HMAC secret + event filter list
- Fires on any event: `monitor.down`, `monitor.recovered`, `incident.created`, `incident.resolved`, etc.
- Delivery logged to `webhook_logs` (status code, duration, error)
- Signature header: `X-Webhook-Signature: sha256=<hmac>`
- Retry on failure (up to 3 attempts)

---

## Security

### AES-256-GCM (pkg/encrypt/encrypt.go)
Used for notification channel config + webhook url/secret stored in DB.
Format: `base64(nonce || ciphertext || tag)`
Key: `ENCRYPTION_KEY` env var (exactly 32 bytes).

### API Keys
- Format: `sk_` prefix + random bytes
- Stored as SHA-256 hash in DB, prefix shown for display
- Full key shown once on creation only
- Bearer token in `Authorization` header for `/api/v1/*`

### JWT Auth
- httpOnly cookie for admin session
- Payload: `user_id` + `role`
- `role = admin` → full write access
- `role = viewer` → GET only, 403 on writes

### Rate Limiting
Redis sliding window on IP. Key: `rl:<ip>`.

### Audit Log
Every admin write (POST/PATCH/DELETE) logs: who, what, when, IP.
**Never log decrypted credential values** — log field names only.

---

## Observability

### Zap Logger
- Production: JSON to stdout (Docker log driver collects it)
- Development: human-readable console
- Always include `trace_id` from OTel span

### Prometheus Metrics (port 2112, /metrics)
| Metric | Type | Labels |
|---|---|---|
| `monitor_check_duration_seconds` | histogram | `type` (http/tcp) |
| `monitor_status_total` | counter | `status` (up/down/degraded) |
| `active_incidents_total` | gauge | — |
| `queue_depth` | gauge | — |
| `http_request_duration_seconds` | histogram | `route`, `method`, `status` |

### OpenTelemetry
- Span on `RunCheck`, handler entry points, DB queries
- Pattern: `ctx, span := tracer.Start(ctx, "OperationName"); defer span.End()`
- `var tracer = otel.Tracer("status-monitor")`

---

## Frontend Patterns

### React Query (lib/hooks/)
```ts
useQuery({
  queryKey: ['monitors', serviceId],
  queryFn: () => api.monitors.list(serviceId),
  staleTime: 30_000,
  refetchInterval: 60_000,
})

useMutation({
  mutationFn: (id: string) => api.monitors.archive(id),
  onSuccess: () => qc.invalidateQueries({ queryKey: ['monitors'] }),
})
```

### Zustand (lib/store/ui.ts)
Client state only: modal open/closed, selected item ID. Nothing server-related.

### Component Pattern
```tsx
if (isLoading) return <Skeleton />
if (isError)   return <ErrorCard message={error.message} />
if (!data?.length) return <EmptyState resource="monitors" />
return <DataList data={data} />
```

### Theme
- Always dark: `<html class="dark">` hardcoded in layout.tsx
- Background: `#09090b` (zinc-950), primary: `oklch(0.72 0.17 265)` (violet)

---

## Docker Compose Services

| Service | Port | Notes |
|---|---|---|
| postgres | internal | healthcheck: `pg_isready`, memory limit 512M |
| redis | internal | `--appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru` |
| backend | 8080, 2112 | HTTP API + Prometheus metrics |
| worker | — | same binary, different CMD (`/worker`), scales independently |
| asynqmon | 8081 | queue visibility UI (`hibiken/asynqmon` image) |
| frontend | 3000 | Next.js |

`compose.yml` — dev: volume mounts, hot reload, no resource limits
`compose.prod.yml` — prod: no mounts, memory limits, `restart: unless-stopped`, `json-file` log driver with rotation

### Multi-stage Dockerfile (Go)
- Stage `builder`: `go build` → `/bin/server` + `/bin/worker`
- Stage `runner`: copies `/bin/server` + `db/migrations/`, exposes 8080 + 2112
- Stage `worker`: copies `/bin/worker` only

---

## Makefile Commands

```makefile
up            # docker compose up
down          # docker compose down
build         # docker compose up --build
logs svc=X    # docker compose logs -f X

shell-backend # docker compose exec backend sh
shell-db      # psql into postgres
shell-redis   # redis-cli

migrate       # run migrations via cmd/migrate
migrate-new   # create timestamped migration file pair
sqlc          # cd backend && sqlc generate

lint          # golangci-lint run ./...
test          # go test ./... -race -cover
test-int      # go test ./... -race -tags integration

create-admin  # seed first admin user
metrics       # open http://localhost:2112/metrics
asynqmon      # open http://localhost:8081
```

---

## Build Phases

| Phase | What Gets Built |
|---|---|
| 1 | Foundation: Fiber + pgx + migrations + sqlc + health. Dockerfile, compose files, Makefile. |
| 2 | Domain types + errors. Repository layer. Migrations 000005–000009. |
| 3 | Service + handler layers. JWT auth. Admin CRUD. create-admin. |
| 4 | TCP check + runner. Asynq worker + scheduler. checker service. Incidents auto-create/resolve. |
| 5 | Rate limiting. Audit log. AES-256-GCM. API keys. /api/v1 public REST API. |
| 6 | Notifications (Discord + email). Webhooks system (outgoing + delivery logs). |
| 7 | Prometheus. Zap logging throughout. OTel tracing. compose.prod.yml. |
| 8 | Frontend: login, admin pages, public status page, charts, uptime bars. |
| 9 | Integration tests. Load test. Grafana dashboard. |

---

## Current State

Phases 1–3 complete. Currently building Phase 4.

### Done
- `backend/Dockerfile` — multi-stage: builder → runner + worker
- `compose.yml` + `compose.prod.yml` — dev + prod (all 6 services)
- `env.example`, `Makefile`, `.gitignore`
- `backend/cmd/server/main.go` — clean arch, golang-migrate on startup, config-driven
- `backend/cmd/migrate/main.go` — standalone migrate CLI (up/down)
- `backend/cmd/create-admin/main.go` — interactive admin seed
- `backend/internal/config/config.go` — typed env config
- `backend/internal/domain/` — errors, monitor, service, incident, apikey, user
- `backend/internal/repository/postgres/` — user, service, monitor, check_result, incident, apikey + convert.go + interfaces.go
- `backend/internal/service/` — user.go, svc.go (services), monitor.go
- `backend/internal/handler/` — router.go (error handler), service.go, monitor.go, user.go
- `backend/internal/middleware/auth.go` — JWT issue + RequireAuth + RequireAdmin
- `backend/internal/checks/http.go` — HTTP check implementation
- `backend/pkg/hash/hash.go` — bcrypt + SHA-256
- `backend/pkg/encrypt/encrypt.go` — AES-256-GCM
- `backend/db/migrations/000001–000009` — all tables
- `backend/db/queries/` — all SQL queries (users, services, monitors, check_results, incidents, api_keys, notification_channels, webhooks, audit_log)
- `backend/db/sqlc/` — fully regenerated generated Go code
- `backend/go.mod` — golang-migrate, asynq, zap, prometheus, jwt, bcrypt, term

### Not Yet Started
- Frontend (all of it)
- API key middleware + service + handler
- /api/v1 public REST API
- Notifications (Discord + email)
- Webhooks (outgoing + delivery logs)
- Telemetry (Zap, Prometheus, OTel)
- Frontend (all of it)

---

## Environment Variables

```env
DATABASE_URL=postgres://statususer:password@postgres:5432/statusdb
REDIS_URL=redis://redis:6379
JWT_SECRET=<32+ random bytes>
ENCRYPTION_KEY=<exactly 32 bytes for AES-256>
APP_ENV=development   # or production
PORT=8080
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
SMTP_FROM=
```

---

## Testing Strategy

| Layer | Approach |
|---|---|
| Domain | Pure unit tests, no mocks needed |
| Repository | Integration tests with testcontainers-go (real Postgres) |
| Service | Unit tests with mock repositories (implement domain interfaces with in-memory maps) |
| Handler | HTTP integration tests using Fiber's `Test()` helper |
| Check engine | Integration tests against `httptest.NewServer` |

---

## Key Rules & Gotchas

- **Never import `fiber` inside a service** — layer violation
- **Never import `pgx` inside a service** — layer violation
- **Never edit files in `db/sqlc/`** — run `sqlc generate` instead
- **Never log decrypted credential values** in audit_log diff — field names only
- **Notifications vs Webhooks** — notifications = Discord/email alert channels; webhooks = generic user-configured HTTP callbacks with event filtering
- `asynq.ErrTaskIDConflict` on enqueue is expected (deduplication) — ignore it
- `consecutive_failures` resets to 0 on first successful check after a failure
- Incident creation is idempotent — always check for open incident before creating
- `SetNextCheckAt` must run even when the check fails — always advance the schedule
- API key stored as SHA-256 hash, prefix shown for display — full key shown once on creation only
- Webhook URL + secret stored AES-256-GCM encrypted; sign delivery with `X-Webhook-Signature: sha256=<hmac>`

---

## Progress Tracker

### Phase 1 — Foundation ✅
- [x] Fiber v3 + pgx/v5 pool + GET /health
- [x] Migrations 000001–000004
- [x] sqlc queries + generated code for services, monitors, check_results
- [x] lib/util.go helpers
- [x] internal/checks/http.go
- [x] golang-migrate wired into server startup
- [x] Dockerfile (multi-stage: builder → runner + worker targets)
- [x] compose.yml (dev — all services)
- [x] compose.prod.yml (prod overrides)
- [x] .env.example
- [x] Makefile

### Phase 2 — Domain + Repository Layer ✅
- [x] internal/domain/errors.go
- [x] internal/domain/monitor.go (struct + interfaces)
- [x] internal/domain/service.go
- [x] internal/domain/incident.go
- [x] internal/domain/apikey.go
- [x] internal/domain/user.go
- [x] internal/repository/postgres/monitor.go
- [x] internal/repository/postgres/service.go
- [x] internal/repository/postgres/check_result.go
- [x] internal/repository/postgres/user.go
- [x] internal/repository/postgres/incident.go
- [x] internal/repository/postgres/apikey.go
- [x] Migrations 000005–000009
- [x] sqlc queries for new tables + regenerate

### Phase 3 — Service + Handler Layer + Auth ✅
- [x] internal/config/config.go
- [x] pkg/hash/hash.go (bcrypt + SHA-256)
- [x] pkg/encrypt/encrypt.go (AES-256-GCM)
- [x] internal/handler/router.go (Fiber factory + error handler)
- [x] internal/middleware/auth.go (JWT)
- [x] cmd/create-admin/main.go
- [x] cmd/migrate/main.go
- [x] internal/service/monitor.go + svc.go + user.go
- [x] internal/handler/service.go + monitor.go + user.go
- [x] Refactor cmd/server/main.go to use clean layers

### Phase 4 — Worker + Check Engine ✅
- [x] internal/checks/tcp.go + runner.go
- [x] internal/worker/tasks.go + scheduler.go + check_monitor.go
- [x] cmd/worker/main.go
- [x] internal/service/checker.go (RunCheck, handleOutage, handleRecovery)
- [x] internal/service/incident.go
- [x] internal/handler/incident.go

### Phase 5 — Security + Public API ✅
- [x] internal/middleware/ratelimit.go + audit.go + apikey.go
- [x] internal/repository/redis/ratelimit.go
- [x] internal/service/apikey.go + handler/apikey.go
- [x] internal/handler/v1/status.go + incidents.go (public REST API)

### Phase 6 — Notifications + Webhooks ✅
- [x] internal/domain/notification.go + webhook.go (types + interfaces)
- [x] internal/repository/postgres/notification.go + webhook.go
- [x] internal/notifications/discord.go + email.go + dispatcher.go
- [x] internal/webhooks/dispatcher.go + delivery.go (HMAC-SHA256 + retry + logging)
- [x] internal/service/notification.go + webhook.go (AES-256-GCM encrypt/decrypt)
- [x] internal/handler/notification.go + webhook.go (CRUD + delivery logs)
- [x] checker.go integrated — fires notifications + webhooks on outage/recovery

### Phase 7 — Observability ✅
- [x] internal/telemetry/logger.go (Zap — JSON prod / console dev)
- [x] internal/telemetry/metrics.go (Prometheus — 5 metrics defined)
- [x] internal/telemetry/tracing.go (OTel — OTLP HTTP or stdout exporter)
- [x] internal/middleware/logger.go (Zap request logger middleware)
- [x] internal/middleware/metrics.go (Prometheus HTTP duration middleware)
- [x] Zap wired into server + worker entrypoints + checker service
- [x] Prometheus /metrics on port 2112 (separate http.Server goroutine)
- [x] OTel span on RunCheck with monitor attributes

### Phase 8 — Frontend
- [x] Next.js 15 setup (shadcn/ui, React Query, Zustand, Tailwind, dark theme)
- [x] Login page + JWT cookie auth
- [x] Admin: services, monitors, incidents, api-keys, notifications, webhooks, users
- [x] Public status page (togglable via NEXT_PUBLIC_STATUS_PAGE env var)

### Phase 9 — Testing
- [ ] Domain unit tests
- [ ] Service unit tests (mock repos)
- [ ] Repository integration tests (testcontainers)
- [ ] Handler HTTP tests (Fiber Test())
- [ ] Check engine tests (httptest.NewServer)
