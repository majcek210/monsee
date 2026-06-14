# Tests.md — Full Functional Test Plan & Results

Scope: every backend endpoint and every frontend button/interaction, verifying the app
**works** (not just "is secure" — see `Security.md` for the security-focused review).

Status legend: ✅ Pass · ❌ Fail (bug found) · 🔧 Fixed during this session · ⏳ Not yet executed · ⚠️ Pass with caveat

---

## 1. Plan

### 1.1 Backend

1. Static checks: `go build ./...`, `go vet ./...`.
2. Unit tests (new, stdlib `testing`, no testify available):
   - `pkg/encrypt` — AES-256-GCM round trip, tamper detection, wrong-key failure.
   - `pkg/hash` — bcrypt round trip, SHA-256 determinism.
   - `internal/middleware/auth` — IssueToken/RequireAuth/RequireAdmin, expiry.
   - `internal/service/apikey` — ownership enforcement on Revoke.
   - `internal/service/webhook`, `internal/service/notification` — partial-update merge
     logic (the bug found below).
   - `internal/service/user` — Login (no enumeration), Register, role validation.
3. Endpoint inventory — walk `router.go` and confirm method/path/auth/role match
   `CLAUDE.md` spec and frontend usage.
4. Live E2E (docker compose: postgres + redis + backend + worker) — exercise auth,
   RBAC, CRUD, rate limiting, public API with `curl`.

### 1.2 Frontend

1. `npm run build`, `npm run lint`, `tsc --noEmit`.
2. Page-by-page button/interaction inventory — for every admin page and the public
   status page, list every interactive element, the API call it triggers, and whether
   that call is implemented/valid on the backend.
3. Manual run against the live backend (dev server) for at least the broken flows
   found below, to confirm the fix.

---

## 2. Backend Endpoint Inventory

| Method | Path | Auth | Role | Frontend uses it? | Status |
|---|---|---|---|---|---|
| GET | `/health` | none | — | — | ✅ |
| POST | `/auth/login` | none | — | yes | ✅ |
| POST | `/auth/logout` | none | — | yes | ✅ |
| GET | `/admin/services` | session | any | yes | ✅ |
| POST | `/admin/services` | session | admin | yes | ✅ Fixed — viewer now 403 (verified live) |
| GET | `/admin/services/:id` | session | any | yes | ✅ |
| PATCH | `/admin/services/:id` | session | admin | yes | ✅ Fixed — viewer now 403 (verified live); see also F5 |
| DELETE | `/admin/services/:id` | session | admin | yes | ✅ |
| GET/POST/PATCH/DELETE | `/admin/monitors*` | session | admin (writes) | yes | ✅ |
| GET/POST/PATCH | `/admin/incidents*`, `/resolve` | session | admin (writes) | yes | ✅ |
| GET | `/admin/api-keys` | session | any (own keys) | yes | ✅ |
| POST | `/admin/api-keys` | session | any | yes | ✅ |
| DELETE | `/admin/api-keys/:id` | session | owner or admin | yes | ✅ Fixed — IDOR closed, verified live (403 cross-user, 204 own/admin) |
| GET/POST/GET/PATCH/DELETE | `/admin/notifications*` | session | admin (writes) | yes | ✅ Fixed — partial PATCH preserves config, toggle returns 200 (verified live) |
| GET/POST/GET/PATCH/DELETE/logs | `/admin/webhooks*` | session | admin (writes) | yes | ✅ Fixed — partial PATCH preserves url/secret, toggle returns 200 (verified live) |
| GET | `/admin/users` | session | admin | yes | ✅ |
| POST | `/admin/users` | session | admin | yes | ✅ Fixed — 201, verified live |
| PATCH | `/admin/users/:id` | session | admin | yes | ✅ Fixed — 200, last-admin lockout verified (409) |
| DELETE | `/admin/users/:id` | session | admin | yes | ✅ (last-admin lockout verified, 409) |
| GET | `/api/v1/status` | none (if enabled) | — | yes | ✅ verified live, no auth required |
| GET | `/api/v1/incidents` | none (if enabled) | — | yes | ✅ verified live, no auth required |
| GET | `/api/v1/incidents/:id` | none (if enabled) | — | yes | ✅ |

---

## 3. Frontend Button / Interaction Inventory

### `/` public status page
| Element | Action | Backend call | Status |
|---|---|---|---|
| (page load) | fetch services + monitors + open incidents | `GET /api/v1/status`, `GET /api/v1/incidents?status=open` | ✅ |
| Incident link | navigate to `/incidents/[id]` | `GET /api/v1/incidents/:id` | ✅ |

### `/login`
| Element | Action | Backend call | Status |
|---|---|---|---|
| Login form submit | authenticate | `POST /auth/login` | ✅ |
| (on success) | redirect to `/admin/services` | — | ✅ |
| (unauthenticated visit to `/admin/*`) | **no redirect — renders shell with empty/error widgets** | — | ❌→🔧 (see Security.md S-FE1) |

### `/admin/services`
| Element | Action | Backend call | Status |
|---|---|---|---|
| "New Service" button → dialog → Save | create service | `POST /admin/services` | ✅ Fixed — `RequireAdmin` added; viewer gets 403, admin 201 (verified live) |
| Row dropdown → Edit → Save | update service | `PATCH /admin/services/:id` | ✅ Fixed (RBAC) + ✅ Fixed (F5 partial-update, verified live) |
| Row dropdown → Archive | archive service | `DELETE /admin/services/:id` | ✅ (admin only) |

### `/admin/services/[id]`
| Element | Action | Backend call | Status |
|---|---|---|---|
| "New Monitor" button → dialog → Save | create monitor | `POST /admin/monitors` (admin) | ✅ |
| Row dropdown → Edit → Save | update monitor | `PATCH /admin/monitors/:id` (admin) | ✅ |
| Row dropdown → Archive | archive monitor | `DELETE /admin/monitors/:id` (admin) | ✅ |
| Type toggle (http/tcp) | client-side form state | — | ✅ |

### `/admin/incidents`
| Element | Action | Backend call | Status |
|---|---|---|---|
| "New Incident" → dialog → Save | create incident | `POST /admin/incidents` (admin) | ✅ |
| Row "Resolve" | resolve incident | `POST /admin/incidents/:id/resolve` (admin) | ✅ |

### `/admin/api-keys`
| Element | Action | Backend call | Status |
|---|---|---|---|
| "New Key" → dialog → Save | create key | `POST /admin/api-keys` | ✅ |
| "Show key once" dialog → Copy | clipboard copy | — | ✅ |
| Row revoke (trash) | revoke key | `DELETE /admin/api-keys/:id` | ✅ Fixed — IDOR closed (verified live: cross-user 403, own/admin 204) |

### `/admin/notifications`
| Element | Action | Backend call | Status |
|---|---|---|---|
| "New Channel" → dialog → Save | create channel | `POST /admin/notifications` (admin) | ✅ |
| Row Enabled switch (inline toggle) | toggle enabled | `PATCH /admin/notifications/:id` with `{id, enabled}` only | ✅ Fixed — returns 200, `config` length unchanged in DB (verified live) |
| Row dropdown → Edit → Save | update channel | `PATCH /admin/notifications/:id` | ✅ Fixed — partial update preserves `config` (see §4) |
| Row dropdown → Archive | archive channel | `DELETE /admin/notifications/:id` (admin) | ✅ |

### `/admin/webhooks`
| Element | Action | Backend call | Status |
|---|---|---|---|
| "New Webhook" → dialog → Save | create webhook | `POST /admin/webhooks` (admin) | ✅ |
| Row Enabled switch (inline toggle) | toggle enabled | `PATCH /admin/webhooks/:id` with `{id, enabled}` only | ✅ Fixed — returns 200, `url`/`secret` lengths unchanged in DB (verified live) |
| Row dropdown → Edit → Save | update webhook | `PATCH /admin/webhooks/:id` | ✅ Fixed — partial update preserves `url`/`secret` (see §4) |
| Row dropdown → Delete | archive webhook | `DELETE /admin/webhooks/:id` (admin) | ✅ |
| Row → "Delivery Logs" dialog | view logs | `GET /admin/webhooks/:id/logs` | ✅ |

### `/admin/users`
| Element | Action | Backend call | Status |
|---|---|---|---|
| "New User" → dialog → Save | create user | `POST /admin/users` | ✅ Fixed — 201 created, verified live |
| Row role `Select` | change role | `PATCH /admin/users/:id` | ✅ Fixed — 200, last-admin lockout returns 409 (verified live) |
| Row archive (trash) | archive user | `DELETE /admin/users/:id` (admin) | ✅ (last-admin lockout returns 409, verified live) |

---

## 4. Detailed Functional Findings

### F1 — Webhook/Notification PATCH silently destroys stored secrets (CRITICAL) — 🔧 Fixed — ✅ Verified live this session
- `WebhookService.Update` **always** re-encrypts `rawURL` and writes it, even when the
  caller sends `""` (e.g. inline "Enabled" toggle, or Edit dialog with the URL field
  left blank per its "leave blank to keep existing" UX). `db/queries/webhooks.sql`
  `UpdateWebhook` does an unconditional `SET url = $3` — no `COALESCE`.
- Same for `secret` (`Secret *string`, nil when blank → repo sets `secret = NULL`).
- `NotificationService.Update` always re-encrypts `config` (even `nil` → `"null"`),
  and `UpdateNotificationChannel` does unconditional `SET config = $3`.
- **Net effect**: any edit that doesn't re-supply the secret fields breaks the
  webhook/notification channel silently — deliveries start failing
  (`parse "": empty url` / `discord channel missing webhook_url`), logged only in
  `webhook_logs`/server logs, never surfaced to the admin UI.
- Additionally both `Update` methods require non-empty `name`, so the inline
  "Enabled" toggle (`{id, enabled}` body) fails outright with `422 name is required`
  — meaning the toggle switch on both list pages is **completely broken** today.
- **Fix**: switch to pointer/optional fields end-to-end (handler binds
  `*string`/`*map[string]any`/`*bool`; service only overwrites a field when the
  pointer is non-nil; repository query keeps existing values via `COALESCE`).
- **Verified live**: inline `{"enabled":false}` toggle on both webhooks and
  notification channels now returns `200` (previously `422 name is required`), and
  `url`/`secret`/`config` byte lengths in the DB are unchanged after the toggle.

### F2 — `/admin/users` Create & Update-role routes missing (HIGH) — 🔧 Fixed — ✅ Verified live this session
- Frontend (`lib/api/users.ts`) calls `POST /admin/users` and `PATCH /admin/users/:id`.
- `router.go` only registers `GET /admin/users` and `DELETE /admin/users/:id`.
- `UserHandler`/`UserService` have no `Create`/`UpdateRole` methods.
- **Fix**: implement `UserService.Create` (bcrypt hash, role validation) and
  `UserService.UpdateRole`, add handler methods, wire
  `POST /admin/users` + `PATCH /admin/users/:id` behind `RequireAdmin`.
- **Verified live**: `POST /admin/users` → `201`; `PATCH /admin/users/:id` (role
  change) → `200`; last-admin lockout (S5a) → `409` on demote/archive of the sole
  remaining admin.

### F3 — API key Revoke is an IDOR (see Security.md S2) — 🔧 Fixed — ✅ Verified live this session
Functionally "works" (200/204) but lets any authenticated user revoke *any* user's
API key by guessing/observing UUIDs.
- **Verified live**: cross-user revoke attempt → `403`; revoking your own key, or any
  key as `admin`, → `204`.

### F4 — RequireAdmin missing on Service Create/Update (see Security.md S1) — 🔧 Fixed — ✅ Verified live this session
Functionally "works" for any logged-in role — but per spec, viewers should get 403.
- **Verified live**: `viewer` → `403` on `POST`/`PATCH /admin/services`; `admin` →
  `201`/`200`.

### F5 — `PATCH /admin/services/:id` silently nulled `description` on partial body (HIGH) — 🔧 Fixed
- Found during live E2E: `curl -X PATCH /admin/services/:id -d '{"name":"x"}'` (omitting
  `description`) set `description` to `NULL` in the DB. Same bug class as F1, in
  `db/queries/services.sql` `UpdateService` (`SET name = $2, description = $3, status = $4`
  — unconditional, no `COALESCE`).
- **Not reachable via the current frontend** — `app/admin/services/page.tsx`'s Edit dialog
  always pre-fills and resends `description` (string, possibly `""`), so the shipped UI
  never sends a body that omits the key. Found via direct API testing, not a button click.
- **Fix**: same pattern as F1 — `UpdateService` now uses
  `SET name = COALESCE(sqlc.narg(name), name), description = COALESCE(sqlc.narg(description), description), status = COALESCE(sqlc.narg(status), status)`.
  `ServiceRepo.Update`/`UpdateStatus` updated accordingly (also removed the now-unnecessary
  pre-fetch of the current row). Regenerated `db/sqlc/services.sql.go`.
- **Verified live**: PATCH `{"name":"Main API v3"}` after previously setting a description
  preserved it; PATCH `{"name":"Main API v3","description":""}` correctly cleared it.
- **Note**: `monitors` (`UpdateMonitor`) and `incidents` (`UpdateIncident`) have the same
  unconditional-`SET` shape with non-pointer required fields, but their only frontend
  callers (`app/admin/services/[id]/page.tsx`, incident edit dialog) always resend the
  full object — not reachable via any button today. Documented as a defense-in-depth
  recommendation in Security.md (S19) rather than fixed, to keep this pass scoped to
  reachable/found issues plus the one bug already demonstrated (F5).

---

## 5. Execution Log

| Step | Result |
|---|---|
| `go build ./...` | ✅ Clean build, no errors |
| `go vet ./...` | ✅ Clean, no warnings |
| `go test ./... -race -cover` | ⚠️ Ran without `-race` (no CGO/gcc on this Windows host — pre-existing host limitation, not a regression). `go test ./...` — all packages pass: `internal/middleware`, `internal/notifications`, `internal/service`, `pkg/encrypt`, `pkg/hash` (others report "no test files", expected) |
| Backend unit tests added | ✅ New tests this session: `pkg/encrypt` (round trip, tamper detection, wrong-key failure), `pkg/hash` (bcrypt round trip, SHA-256 determinism), `internal/middleware` (IssueToken/RequireAuth/RequireAdmin/expiry), `internal/service` (apikey ownership on Revoke, webhook/notification partial-update merge, user Login/Register/role validation, last-admin lockout), `internal/notifications/email_test.go` (3 new `stripCRLF` tests for S6) |
| `npm run build` | ✅ Production build succeeds, no errors. `proxy.ts` migration (Next.js 16 `middleware.ts`→`proxy.ts`) confirmed — deprecation warning gone, route table shows `ƒ Proxy (Middleware)` for `/admin/:path*` |
| `npm run lint` | ⚠️ Unavailable — Next.js 16 removed `next lint` and no standalone ESLint config exists in this project. Pre-existing condition, not introduced this session; not a regression to fix as part of this review |
| `tsc --noEmit` | ✅ No type errors |
| Live E2E (docker compose) | ✅ Full stack (`postgres`, `redis`, `backend`, `worker`, `frontend`, `asynqmon`) brought up healthy via `docker compose up -d`. Verified via `curl`/direct DB checks: <br>• S1/F4 — RBAC on `POST`/`PATCH /admin/services`: viewer → 403, admin → 201/200 <br>• S2/F3 — API key Revoke IDOR closed: cross-user → 403, own/admin → 204 <br>• S3 — JWT contains `exp`/`iat`, 24h apart <br>• F1/S4 — webhook & notification partial PATCH (`{"enabled":false}`) → 200, `url`/`secret`/`config` unchanged in DB <br>• F2/S5 — `POST /admin/users` → 201, `PATCH /admin/users/:id` role change → 200 <br>• S5a — demoting/archiving the last admin → 409 (then reverted test user back to admin to restore pre-existing data) <br>• F5 — `PATCH /admin/services/:id` partial update preserves `description`; explicit `description:""` still clears it <br>• S9 — full CSP + security headers present on `/` <br>• S10 — unauthenticated `/admin/services` → 307 redirect to `/login`; authenticated → 200 <br>• Rate limiting — burst of 110 requests → exactly 100×200 + 10×429 <br>• Public `/api/v1/status` and `/api/v1/incidents` — no auth required, both return live data including an auto-created incident from the worker/checker flow <br>Stack torn down afterward with `docker compose down` (named volumes retained) |

All planned checks executed. No outstanding ⏳ items.
