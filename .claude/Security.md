# Security.md — Security Review Plan & Findings

Scope: authentication, authorization, injection, secrets management, transport,
rate limiting, audit logging, SSRF, dependency hygiene, and frontend hardening for
monsee.

Severity: **Critical** (data/account compromise) · **High** (privilege escalation /
IDOR / integrity) · **Medium** (defense-in-depth gaps) · **Low** (hardening / info
leakage) · **Info** (accepted risk, documented for awareness).

Status legend: ⏳ Open · 🔧 Fixed this session · ✅ Verified safe (no action needed) ·
📝 Documented as accepted risk / recommendation only.

---

## 1. Plan

Review areas, each walked end-to-end through domain → repository → service → handler
→ frontend:

1. **Authentication** — JWT issuance/verification, cookie attributes, password
   hashing, login enumeration.
2. **Authorization (RBAC)** — every admin write route checked against
   `RequireAuth`/`RequireAdmin`, ownership checks for user-scoped resources
   (API keys).
3. **Injection** — SQL (sqlc/pgx parameterization), SMTP header injection,
   command injection, XSS (frontend rendering).
4. **Secrets management** — encryption at rest (AES-256-GCM), key handling,
   `.env` hygiene, what's returned to clients (`safeWebhook`/`safeNotifChannel`).
5. **Rate limiting & abuse** — Redis sliding window, fail-open behavior.
6. **CSRF** — cookie `SameSite`/`Secure` attributes, state-changing GETs.
7. **Audit logging** — coverage and `Diff` field correctness.
8. **Transport / HTTP hardening** — security headers, CORS, CSP.
9. **SSRF** — webhook delivery to admin-supplied URLs, notification dispatch.
10. **Dependency hygiene** — `go.mod`/`package.json` versions, `go vet`.
11. **Frontend** — auth-gating of `/admin/*`, dead/broken API calls that could mask
    silent failures of security-relevant settings (notifications/webhooks).

---

## 2. Findings

### S1 — `POST/PATCH /admin/services` missing `RequireAdmin` (HIGH) — 🔧 Fixed
- **Where**: `internal/handler/router.go`
- **Evidence**: every other write route (`monitors`, `incidents`, `notifications`,
  `webhooks`) has `middleware.RequireAdmin`; `services` Create/Update did not
  (only `Delete` did).
- **Impact**: a `viewer`-role session can create/rename services, contradicting
  the documented RBAC model (`role = viewer → GET only, 403 on writes`).
- **Fix**: add `middleware.RequireAdmin` to both routes.

### S2 — API key Revoke is an IDOR (HIGH) — 🔧 Fixed
- **Where**: `internal/service/apikey.go` `Revoke`, `internal/handler/apikey.go`
  `Revoke`, `router.go` `DELETE /admin/api-keys/:id`.
- **Evidence**: `Revoke(ctx, id)` calls `s.keys.Archive(ctx, id)` directly — no check
  that the key belongs to `middleware.UserIDFromCtx(c)`. `ListByUser`/`Create` are
  correctly user-scoped, `Revoke` is not.
- **Impact**: any authenticated user (including `viewer`) can revoke **any other
  user's** API key by ID, since UUIDs are returned/observable in the same response
  shape and not otherwise secret — a denial-of-service against another user's
  integrations.
- **Fix**: `Revoke(ctx, userID, id)` loads the key, returns `404` if missing,
  `403`(`ErrForbidden`) if `key.UserID != userID` **and** caller isn't `admin`.
  Added `APIKeyRepository.GetByID`.

### S3 — JWT has no `exp`/`iat` claim (MEDIUM) — 🔧 Fixed
- **Where**: `internal/middleware/auth.go` `IssueToken`.
- **Evidence**: `Claims{UserID, Role}` embeds `jwt.RegisteredClaims{}` but never sets
  `ExpiresAt`/`IssuedAt`. `jwt.ParseWithClaims` therefore never rejects the token on
  expiry — only the **cookie's** 24h expiry limits the browser from *sending* it.
  A token captured via XSS/log-leak before the cookie expires remains valid forever
  if replayed directly against the API.
- **Fix**: set `ExpiresAt: jwt.NewNumericDate(now.Add(24*time.Hour))` and `IssuedAt`
  to match the cookie lifetime, so the library enforces expiry server-side too.

### S4 — Webhook/Notification PATCH overwrites encrypted secrets with empty values (HIGH — integrity of security controls) — 🔧 Fixed
- Cross-ref `Tests.md` F1. Security angle: this silently disables the alerting
  channels that are supposed to notify admins of outages — a monitoring blind spot
  that an attacker causing an outage would benefit from, and that nobody would
  notice because the UI reports success (`200 OK`).
- **Fix**: partial-update semantics (pointer fields, `COALESCE` in SQL) — see
  Tests.md F1 fix description.

### S5 — `/admin/users` Create/Update missing (HIGH, functional→security) — 🔧 Fixed
- Cross-ref Tests.md F2. Security angle: there is currently **no way to create
  additional users or demote/promote roles via the running application** — the only
  path is the `create-admin` CLI (which always creates `admin` role). Once
  implemented, both new routes are placed behind `RequireAdmin`.
- Additional check while implementing: ensure an admin cannot demote/delete their
  **own** account in a way that locks everyone out — see S5a below.

### S5a — Self-demotion / last-admin lockout (MEDIUM) — 🔧 Mitigated
- **Where**: new `UpdateRole`/existing `Archive` for users.
- **Risk**: an admin could demote or archive the only remaining admin account
  (themselves or the last one), locking the team out of `/admin/*` entirely (no
  recovery path except direct DB access / re-running `create-admin`).
- **Fix**: `UserService.UpdateRole` and `Archive` reject the operation
  (`ErrConflict`) if it would leave zero active admins.

### S6 — SMTP header injection via notification config (LOW) — 🔧 Fixed
- **Where**: `internal/notifications/email.go`.
- **Evidence**: raw SMTP message built via string concatenation of `From`/`To`/
  `Subject`/`Body` without stripping `\r\n`.
- **Impact**: low — these values come from `notification_channels.config`
  (admin-only, AES-256-GCM encrypted) and monitor/service names (also admin-managed).
  Still, defense-in-depth: a compromised/careless admin entry (e.g. a monitor name
  containing CRLF) could inject extra SMTP headers/recipients.
- **Fix**: strip `\r` and `\n` from header-bound values before building the message.

### S7 — JWT/session cookie attributes (LOW) — ✅ Verified
- `HttpOnly: true`, `Secure: cfg.IsProd()`, `SameSite: Lax` — correct for a
  same-site admin panel proxied through Next.js. `SameSite=Lax` + `HttpOnly` covers
  CSRF for the cookie-based admin API (no state-changing `GET`s found — all writes
  are POST/PATCH/DELETE). No change required.
- Minor nit: `Logout` clears the cookie without repeating `Secure`/`SameSite`/`HttpOnly`
  attributes. Functionally fine (browsers match on name+path+domain for deletion),
  but tidied up for consistency — 📝 low-priority cosmetic fix applied.

### S8 — Audit log `Diff` never populated (MEDIUM) — 🔧 Fixed
- **Where**: `internal/middleware/audit.go` — `AuditEntry.Diff` was never set, so
  `audit_log.diff` was always `NULL`.
- **Impact**: audit trail recorded *who/what/when* but not *what changed* — reduced
  forensic value for investigating unauthorized admin changes.
- **Fix**: `Audit()` now captures `c.Body()` before `c.Next()` runs, and for
  `POST`/`PATCH` requests extracts the **sorted top-level JSON field names** via a
  new `auditFieldNames()` helper, setting `entry.Diff = map[string]any{"fields":
  [...]}`. Field **values are never inspected or logged** — this trivially satisfies
  the `CLAUDE.md` rule "never log decrypted credential values" since the function
  only ever sees key names, even for encrypted fields like webhook `url`/`secret` or
  notification `config`.
- **Tests**: 4 new unit tests in `internal/middleware/audit_test.go` — sorted
  field-name extraction, nil/empty body, invalid JSON, non-object JSON (array). All
  pass.

### S9 — No security headers / CSP on frontend (LOW) — 🔧 Fixed
- **Where**: `frontend/next.config.ts` had no `headers()`.
- **Fix**: added `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
  `Referrer-Policy: strict-origin-when-cross-origin`, and a baseline
  `Content-Security-Policy` (self + the few origins actually needed).

### S10 — `/admin/*` has no client-side auth guard (LOW) — 🔧 Fixed
- **Where**: `frontend/app/admin/layout.tsx`.
- **Evidence**: renders the sidebar + page unconditionally; an unauthenticated
  visitor sees the admin shell with failed (`401`) data fetches instead of being
  redirected to `/login`.
- **Impact**: low — the underlying API still returns `401`/enforces RBAC, so no data
  is exposed. But it's a poor signal to an attacker/UX, and a "fail open" pattern is
  worth tightening for defense-in-depth.
- **Fix**: added `frontend/proxy.ts` (Next.js 16 renamed `middleware.ts`/`middleware()`
  to `proxy.ts`/`proxy()`; `frontend/middleware.ts` was deleted) that checks for
  presence of the `session` cookie on `/admin/*` and redirects to `/login` if absent
  (presence check only — signature verification stays server-side where `JWT_SECRET`
  lives).

### S11 — Sidebar shows "Users" nav link to viewers (LOW) — 🔧 Fixed
- **Where**: `frontend/components/admin/sidebar.tsx`.
- Viewer role would see a "Users" link leading to a page where `GET /admin/users`
  is `RequireAdmin` (so the page errors for them). Hid the link for non-admin
  roles for a cleaner UX/least-privilege presentation. No data exposure either way.

### S12 — Rate limiting fails open on Redis errors (INFO) — 🔧 Mitigated
- **Where**: `internal/repository/redis/ratelimit.go`, `middleware/ratelimit.go`.
- If Redis is unreachable, requests are allowed through rather than blocked.
- **Decision**: kept the fail-open behavior — this remains a deliberate
  availability-over-strictness tradeoff appropriate for this app (the public status
  page must stay up even if Redis is down; fail-closed would turn a Redis outage into
  a full site outage, which is worse for an app whose purpose is reporting outages).
- **What changed**: the fail-open path was previously silent (a bare comment, nothing
  logged or counted). It's now **observable**: `RateLimit()` takes `*zap.Logger` and a
  `prometheus.Counter`, and on a Redis error it logs `zap.Warn("rate limiter backend
  unavailable, failing open", ...)` and increments a new
  `rate_limiter_fail_open_total` counter (registered in
  `internal/telemetry/metrics.go`, wired in `router.go`). An operator can now alert on
  sustained fail-open instead of the condition being invisible.
- **Verified live**: `/metrics` exposes `rate_limiter_fail_open_total` (0 under normal
  operation with Redis healthy); rate limiting itself still enforces the 100/min
  sliding window (burst of 105 requests → first 100 return `200`, remainder `429`).

### S13 — SQL injection — ✅ Verified safe
- All `db/queries/*.sql` use `sqlc`-generated parameterized queries via `pgx/v5`.
- `grep -rn 'Sprintf|" + |+ "'` across `internal/repository` only matches
  `redis/ratelimit.go` (Redis key construction, not SQL) and `postgres/convert.go`
  (type conversion helpers, no SQL). No string-built SQL found.

### S14 — XSS — ✅ Verified safe
- Public status page renders all dynamic content as React text nodes (no
  `dangerouslySetInnerHTML`, no `innerHTML`). React auto-escapes.

### S15 — Webhook/notification delivery SSRF (INFO → MEDIUM after re-assessment) — 🔧 Fixed
- `internal/webhooks/delivery.go` and `internal/notifications/discord.go` POST to
  admin-supplied URLs. Previously documented as an accepted risk ("trusted admin can
  make the backend issue requests to arbitrary hosts"). On `do what you think is best
  so the app won't get hacked`, re-assessed and mitigated rather than left open —
  even an admin-only SSRF lets a compromised/malicious admin account pivot to
  internal infrastructure (e.g. cloud metadata endpoints, internal-only admin APIs on
  other services in the same Docker network) with the backend's network identity, and
  defense-in-depth against a *compromised* admin session is in scope for "won't get
  hacked".
- **Fix**: new `pkg/netguard` package with two functions:
  - `CheckPublicURL(rawURL)` — resolves the host via DNS and blocks loopback,
    private (RFC1918), link-local (incl. `169.254.169.254` cloud metadata),
    unspecified, and multicast addresses, for both IPv4 and IPv6 (`::1`, `fe80::/10`,
    `fc00::/7`). Applied at the **delivery boundary**: top of
    `webhooks/delivery.go doDeliver` and `notifications/discord.go SendDiscord`. This
    re-checks DNS on every delivery, so a hostname that resolves to a public IP at
    config time but is later DNS-rebound to a private IP is still blocked.
  - `CheckPublicURLSyntax(rawURL)` — scheme + IP-literal check only, **no DNS
    lookup**. Applied at **config time** (`WebhookService.Create/Update`,
    `NotificationService` discord channel validation) for immediate UX feedback on
    obviously-bad URLs (`http://127.0.0.1/...`, `http://169.254.169.254/...`) without
    making every admin write synchronously depend on DNS resolution (avoids a
    reliability/test-fragility regression: `CheckPublicURL` on `example.com`-style
    hostnames would require live DNS in every environment, including CI).
- **Tests**: `pkg/netguard/netguard_test.go` — 7 unit tests covering blocked
  IP-literal/hostname targets (IPv4 + IPv6), allowed public IPs, bad schemes, missing
  host, and that `CheckPublicURLSyntax` does not perform DNS lookups. All pass.
  Existing `internal/service/webhook_test.go` / `notification_test.go` (which use
  `https://example.com/...` / `https://discord.com/...`) continue to pass unchanged.
- **Verified live**:
  - `POST /admin/webhooks` with `url: "http://127.0.0.1:8080/admin"` → `422` (`target
    address is not allowed: 127.0.0.1`); same for `http://169.254.169.254/...`
    (cloud metadata). A normal `https://example.com/webhook` → `201`.
  - DNS-rebinding case: created a webhook with `url: "http://localtest.me/webhook"`
    (a hostname that resolves to `127.0.0.1`) — config-time check allows it (no DNS
    lookup), but on actual delivery (`monitor.recovered` event) `webhook_logs` shows
    `error: "blocked target: target address is not allowed: localtest.me resolves to
    127.0.0.1"`, `status_code: NULL`, `duration_ms: 0` — confirmed no HTTP request
    was made to the loopback address.

### S16 — Dependency / build hygiene (INFO) — ✅ Verified
- `go vet ./...` clean after all fixes (see Execution Log).
- `go.mod` / `package.json` versions are current (Fiber v3.3.0, pgx v5.10.0,
  golang-jwt v5.3.1, Next 16.2.9, React 19).
- No `go.sum`/`package-lock` CVE-scanning audit tool (e.g. `govulncheck`, `npm audit`)
  was run — no such tool is installed in this environment and none was requested.
  📝 Documented as a recommendation for a future pass / CI integration, not a finding
  against current code.

### S19 — `services.UpdateService` overwrote `description` with `NULL` on partial PATCH (HIGH — integrity) — 🔧 Fixed
- Cross-ref `Tests.md` F5. Same bug class and same fix shape as S4: an unconditional
  `SET description = $3` in `db/queries/services.sql` meant any `PATCH
  /admin/services/:id` body that omitted `description` silently erased it in the
  database, even though the request "succeeded" (`200 OK`).
- **Security angle**: a service's description is the text shown to the public on the
  status page (`/`). Silent data loss on an admin-facing write that returns success
  is the same "monitoring/UI says everything is fine but data integrity is silently
  broken" pattern as S4 — admins have no signal that content was wiped.
- **Reachability**: not exploitable via the current frontend (the Edit dialog always
  resends `description`), so this is a latent API-level integrity bug rather than a
  live exploit path — but `/admin/services/:id` is a documented, usable API surface
  (same auth as the rest of `/admin/*`), so it's in scope.
- **Fix**: `UpdateService` query rewritten to
  `SET name = COALESCE(sqlc.narg(name), name), description = COALESCE(sqlc.narg(description), description), status = COALESCE(sqlc.narg(status), status)`
  — identical pattern to the S4/F1 fix for webhooks/notifications. Regenerated
  `db/sqlc/services.sql.go`; `ServiceRepo.Update`/`UpdateStatus` updated.
- **Verified live**: PATCH with `{"name":"Main API v3"}` (no `description` key)
  preserved the existing description; PATCH with `{"description":""}` still clears it
  as expected (explicit empty value = explicit intent).
- **Related, deliberately deferred (📝 decision, not a TODO)**: `monitors.UpdateMonitor`
  and `incidents.UpdateIncident` have the same unconditional-`SET` shape. Considered
  applying the same `COALESCE(sqlc.narg(...), ...)` pattern here too, but decided
  **not to** in this pass:
  - Unlike `services.description` (a single nullable text field), `UpdateMonitorParams`
    mixes required non-pointer fields (`Name`, `IntervalSeconds`, `TimeoutMs`,
    `RetryCount`, `Enabled`) with several *already-nullable* pointer fields (`URL`,
    `Host`, `Port`, `DegradedThresholdMs`, `HTTPMethod`, `HTTPExpectedStatus`).
    Mechanically applying `nil = "no change"` via `COALESCE` would make it impossible
    to ever explicitly clear `degraded_threshold_ms`/`port`/`http_expected_status`
    back to `NULL` again via PATCH (an absent key and an explicit `null` would become
    indistinguishable) — a *new* 3-state absent/null/value semantic would be needed to
    do this correctly, which is a larger, riskier change than the original bug.
  - No reachable partial-PATCH path was found for either resource — every current
    frontend caller resends the full object — so this is a latent, non-demonstrated
    gap, not a live exploit.
  - Given the choice between (a) shipping a quick mechanical port that risks a *new*
    silent-data-loss regression on nullable monitor fields, and (b) leaving the
    current (safe-but-unconditional) behavior in place until a real partial-PATCH
    caller exists and the 3-state semantic can be designed properly, (b) is the lower
    overall risk. **Recommendation stands for a future pass** if/when an inline-edit
    UI or `/api/v1` write endpoint introduces a real partial-PATCH path for monitors
    or incidents.

### S17 — Password handling — ✅ Verified safe
- `pkg/hash` uses bcrypt cost 12 for password hashing, SHA-256 for API key hashing
  (appropriate — API keys are high-entropy random tokens, not user-chosen secrets).
- `UserService.Login` returns the same generic `"invalid credentials"` for both
  unknown email and wrong password — no user-enumeration via timing/message
  differences (bcrypt comparison only runs on the not-found path's absence, but the
  message is identical either way).

### S20 — Webhook/notification dispatch goroutines used a context that's canceled before they finish (MEDIUM — functional, discovered while verifying S15) — 🔧 Fixed
- **Where**: `internal/webhooks/dispatcher.go` `Dispatch`, `internal/notifications/dispatcher.go` `Dispatch`.
- **Evidence**: both fire delivery in a `go func() { ... }()` using the `ctx`
  passed in from `CheckerService.RunCheck`, which in turn is the asynq task
  handler's context. asynq cancels that context as soon as `ProcessTask` returns —
  but the calling function returns immediately after spawning the goroutines, so the
  HTTP request and `InsertLog` call race the cancellation and lose almost every time.
  Live `docker compose` worker logs showed `webhooks: insert log: context canceled`
  and `notifications: send monitor.down via discord: ... context canceled` on every
  delivery attempt, and `webhook_logs` was permanently empty despite webhooks being
  configured and incidents firing.
- **Impact**: this silently broke the entire alerting/webhook system (Phase 6, marked
  complete) — Discord/email notifications and outgoing webhooks essentially never
  completed delivery, and `webhook_logs` provided no visibility into that failure.
  It also meant the new S15 delivery-time SSRF block would never be recorded.
- **Fix**: both dispatch goroutines now build a detached context —
  `context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)` — before calling
  `deliver`/`send`, so delivery survives the parent task context's cancellation but
  still has a bounded deadline. (Also dropped two now-redundant `x := x` loop-variable
  copies flagged by `forvar`, harmless under Go 1.22+ per-iteration loop vars.)
- **Verified live**: after rebuilding the `worker` image, triggered a real
  `monitor.recovered` event — `webhook_logs` now contains rows (previously zero), and
  the Discord notification dispatch completed with an actual HTTP response
  (`discord webhook returned 404` for a fake test webhook ID) instead of `context
  canceled`.

### S18 — CORS — ✅ Verified, no action
- No CORS middleware registered. Admin API is same-origin via the Next.js proxy
  (cookie auth), and `/api/v1/*` requires no auth but also sends no cookies — default
  same-origin policy is the safe default. If third-party browser-based consumers of
  `/api/v1/*` are needed in the future, add a permissive CORS policy scoped to that
  group only.

---

## 3. Execution Log

| Check | Result |
|---|---|
| S1 fix applied + verified | ✅ `RequireAdmin` added to `POST`/`PATCH /admin/services`; live: viewer → 403, admin → 201/200 |
| S2 fix applied + verified | ✅ `Revoke` now ownership-checked (`APIKeyRepository.GetByID` + 403/404); live: cross-user → 403, own/admin → 204 |
| S3 fix applied + verified | ✅ `IssueToken` sets `ExpiresAt`/`IssuedAt`; live: decoded JWT shows `exp`/`iat` 24h apart |
| S4 fix applied + verified | ✅ Partial-update (pointer fields + `COALESCE`) for webhooks/notifications; live: `{"enabled":false}` toggle → 200, `url`/`secret`/`config` unchanged in DB |
| S5/S5a fix applied + verified | ✅ `POST/PATCH /admin/users` implemented behind `RequireAdmin`; last-admin lockout returns 409. Live: create → 201, role change → 200, demote/archive last admin → 409 (test data reverted afterward) |
| S6 fix applied + verified | ✅ `stripCRLF` applied to SMTP header-bound values; 3 new unit tests in `internal/notifications/email_test.go`, all pass |
| S9 fix applied + verified | ✅ Security headers + CSP added in `frontend/next.config.ts`; live: all headers (incl. CSP) present on `GET /` |
| S10 fix applied + verified | ✅ `frontend/proxy.ts` added (Next.js 16 `proxy`/`middleware` rename); live: unauthenticated `/admin/services` → 307 to `/login`, authenticated → 200 |
| S11 fix applied + verified | ✅ "Users" nav link hidden for non-admin roles in `frontend/components/admin/sidebar.tsx` |
| S19 fix applied + verified | ✅ `UpdateService` query rewritten with `COALESCE(sqlc.narg(...), ...)`; live: name-only PATCH preserves `description`, explicit `description:""` clears it |
| `go vet ./...` after fixes | ✅ Clean, no warnings |
| Live RBAC test (viewer 403 on writes) | ✅ Verified across services/users; viewer receives 403 on all write routes, 200 on reads |
| Live API key ownership test | ✅ Cross-user revoke → 403; own key or admin → 204 |
| Live rate-limit test (429 after burst) | ✅ Burst of 110 requests in <1 min from one IP → 100×200 + 10×429 |
| S8 fix applied + verified | ✅ 4 new tests in `internal/middleware/audit_test.go` pass; live: `audit_log.diff` for a webhook `POST` shows `{"fields":["events","name","url"]}` (no values), a PATCH toggling `enabled` shows `{"fields":["enabled"]}` |
| S15 fix applied + verified | ✅ 7 new tests in `pkg/netguard/netguard_test.go` pass; existing webhook/notification service tests unaffected; live: `127.0.0.1`/`169.254.169.254` webhook URLs → `422` at create, `https://example.com/...` → `201`; DNS-rebinding hostname (`localtest.me` → `127.0.0.1`) passes config-time check but is blocked at delivery (`webhook_logs.error = "blocked target: ... resolves to 127.0.0.1"`) |
| S12 fix applied + verified | ✅ `rate_limiter_fail_open_total` registered and visible on `/metrics` (0 with Redis healthy); rate limiting still enforces 100/min (same burst test as above) |
| S20 fix applied + verified | ✅ `context.WithoutCancel` + 30s timeout in both dispatchers; rebuilt `worker` image; live: `webhook_logs` now populated (previously always empty), Discord dispatch completes with a real HTTP response instead of `context canceled` |
| `go build ./... && go vet ./... && go test ./...` after S8/S12/S15/S20 | ✅ Clean build, no vet warnings, all packages pass |
| E2E cleanup | ✅ Test webhooks/monitor/service archived after verification; rate-limit key cleared in Redis |

All planned checks executed via the live `docker compose` E2E stack documented in
`Tests.md` §5. No outstanding ⏳ items.
