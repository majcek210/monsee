# Security.md — Security Review Plan & Findings (v1.01 / Phases 10–20)

Scope: authentication & 2FA, authorization (RBAC + ownership), injection, secrets
management, SSRF, transport hardening, rate limiting, audit logging, and frontend
hardening — focused on the v1.01 surface (settings, per-service visibility & dedicated
pages / custom domains, new check types ssl/keyword/dns, maintenance windows, TOTP 2FA,
incident timeline, badge/RSS, status override, audit-log viewer, latency) plus a
regression check that the Phase 1–9 fixes still hold.

Severity: **Critical** (data/account compromise) · **High** (privilege escalation / IDOR
/ integrity) · **Medium** (defense-in-depth gaps) · **Low** (hardening / info leakage) ·
**Info** (accepted risk, documented).

Status legend: ⏳ Open / not yet executed · 🔧 Fixed this session · ✅ Verified safe ·
📝 Documented as accepted risk / recommendation.

---

## 1. Plan

Each area walked end-to-end through domain → repository → service → handler → frontend.

1. **2FA / TOTP** — secret generation & storage (AES-256-GCM, never returned after
   setup), backup-code hashing (SHA-256) & single-use consumption, verify endpoint
   abuse, **login enforcement** (is `/auth/2fa/verify` actually required before a
   session is issued for a 2FA-enabled account, or is it bypassable?), rate-limiting of
   the verify endpoint, replay.
2. **Authentication** — JWT issuance/verification, cookie attributes, expiry, login
   enumeration (regression of S3/S17).
3. **Authorization (RBAC + ownership)** — every NEW write route checked against
   `RequireAuth`/`RequireAdmin`; every NEW read route that exposes data checked for the
   right minimum role. Specifically: settings PATCH (admin), maintenance writes (admin),
   2fa/disable (admin), audit-log (admin). Read routes (`/admin/settings`,
   `/admin/uptime`, `/admin/services/:id/uptime`, `/admin/monitors/:id/latency`,
   `/admin/maintenance-windows`, `/admin/incidents/:id/updates`) — confirm intended role.
4. **Public API exposure & the settings gate** — `/api/v1/*` is now ALWAYS registered
   (no env gate). Verify the `public_status_enabled` guard actually blocks
   status/incidents/uptime/pages/by-domain when disabled, that `/api/v1/settings` is
   exempt and leaks nothing sensitive, and that disabling the public page can't be
   bypassed via the new endpoints (badge, rss, by-domain, pages/:slug, latency).
5. **SSRF (new check types)** — `checks/ssl.go` (tls.Dial to admin host:port),
   `checks/keyword.go` (HTTP GET to admin URL), `checks/dns.go` (resolver lookups). These
   make the backend connect to admin-supplied targets from the worker. Assess against
   internal-network/metadata pivots; cross-ref the Phase-6 `pkg/netguard` SSRF control —
   do the new checks reuse it or bypass it?
6. **Injection** — SQL (sqlc/pgx parameterization incl. the new string-interpolated
   interval param in `check_results.sql` `GetDailyUptimeForMonitor`), XSS in new SVG
   badge + RSS feed (both build XML/SVG via `fmt.Sprintf` with dynamic site title /
   incident titles / status), DNS/keyword inputs.
7. **Secrets & info leakage** — `/api/v1/settings` response shape, `/auth/me` shape (must
   not leak `totp_secret`/`totp_backup_codes`/`password_hash`), audit-log diff (field
   names only), uptime/latency endpoints leaking archived/private services.
8. **Custom domain / dedicated page** — host-based routing hardening (the v1.01 plan's
   refinement #9: a vanity domain must not expose `/admin/*` or `/login`), slug/domain
   uniqueness & clear-on-archive (refinement #8).
9. **CSP regression** — `img-src` was widened to `https:` for external logos; confirm no
   other directive was loosened and `script-src`/`frame-ancestors` are intact.
10. **Rate limiting & abuse** — 2FA verify brute force, audit-log enumeration.
11. **Regression** — S1–S20 fixes from the v1.0 pass still present (RBAC on services,
    API-key IDOR, JWT exp, partial-PATCH COALESCE, last-admin lockout, SMTP CRLF, audit
    diff, headers, proxy auth-gate, netguard SSRF, dispatcher detached context).

---

## 2. Findings

### N1 — 2FA login enforcement — 🔧 Fixed (Critical)
- **Finding:** `POST /auth/login` (handler/user.go) issued a full JWT session cookie
  immediately after password verification, without checking `u.TOTPEnabled`. The
  `/auth/2fa/verify` endpoint was wired but never enforced — a user with a valid password
  bypassed 2FA entirely by calling the login API directly, regardless of the frontend UI.
  The frontend login page also had no second-step UI (`router.push` ran immediately).
- **Impact:** 2FA provided zero actual protection. An attacker with a stolen password was
  in with a full-privilege session.
- **Backend fix:** `handler/user.go Login` — when `u.TOTPEnabled == true`, return
  `{id, totp_required: true}` with NO cookie. `handler/twofactor.go Verify` — on
  successful TOTP/backup-code validation, look up the user and issue the full session
  cookie. Added `users *service.UserService` + `cfg *config.Config` to `TwoFactorHandler`.
- **Frontend fix:** `app/login/page.tsx` — two-phase state machine: phase 1 = password
  form, phase 2 = TOTP input (shown when response has `totp_required: true`); only
  redirects to `/admin` after `POST /auth/2fa/verify` succeeds.
- **Verified:** `go build ./...` passes. Frontend `tsc --noEmit` + `npm run build` pass.

### N2 — 2FA verify brute force / rate limiting — 🔧 Fixed (two gaps, both closed)
- **Finding 1 (per-account rate limit):** Rate limit was per-IP only. An attacker
  rotating IPs faced no per-account ceiling on 6-digit TOTP guesses.
  - **Fix:** Added `limiter middleware.Limiter` to `TwoFactorHandler`. `Verify` handler
    checks `limiter.Allow(ctx, "2fa:"+userID)` before calling the service — a separate
    key namespace from the global `"rl:<ip>"` key, so rotating IPs still hit the
    per-account ceiling. Fails open on Redis error (same policy as global limiter).
  - Files changed: `handler/twofactor.go`, `handler/router.go`
- **Finding 2 (TOCTOU race on backup codes):** `service/twofactor.go` fetched the backup
  code list, compared hashes in Go, then called `RemoveBackupCode` — a non-atomic
  read-modify-write. Two concurrent requests with the same backup code could both pass
  the comparison before either removal completed, authenticating twice with one code.
  - **Fix:** Replaced the fetch+loop+remove pattern with a single atomic SQL statement:
    `ConsumeBackupCode :execrows` — `UPDATE users SET totp_backup_codes = array_remove(..., $2) WHERE id = $1 AND $2 = ANY(totp_backup_codes)`. Returns rows-affected=1 on
    success, 0 if the code was absent or already consumed. No race window exists.
  - Files changed: `db/queries/users.sql`, `db/sqlc/users.sql.go` (regenerated),
    `domain/user.go` (interface), `repository/postgres/user.go` (impl),
    `service/twofactor.go` (caller), `service/user_test.go` (mock stub)
- **Verified:** `go build ./...` clean · `go vet ./...` clean · 33/33 tests pass.

### N3 — 2FA secret / backup-code non-exposure — ✅ Verified safe
- `domain.User` struct has no `TOTPSecret` or `TOTPBackupCodes` fields. Secrets live in
  a separate `domain.TOTPData` struct used only internally by `TwoFactorService`.
- `PasswordHash` is tagged `json:"-"` in the User struct.
- `/auth/me` returns only: id, email, role, totp_enabled, created_at, archived_at.
- Audit middleware records only JSON field **names**, never values — confirmed for 2FA
  endpoints (`{"password":"secret"}` → diff stores `["password"]` key only).
- The plaintext secret + otpauth URI are returned once by `InitiateSetup` (expected);
  backup codes once by `ConfirmSetup` (expected). No other path returns them.

### N4 — Public settings gate bypass — 🔧 Fixed (Medium/High)
- **Finding:** Every `/api/v1/*` data handler was completely ungated — no group-level
  middleware and no per-handler check for `public_status_enabled`. When an admin disables
  the public page in settings, all endpoints (`/status`, `/incidents`, `/uptime`,
  `/badge.svg`, `/rss`, `/pages/:slug`, `/by-domain`, `/latency`) remained fully
  accessible. The comment `"settings guard per-handler"` in router.go described intended
  design but none of the handlers implemented it.
  - File: `backend/internal/handler/router.go:193-228`, all `v1/*.go` handlers
- **`/api/v1/settings` response:** Correctly returns only `site_title` + `logo_url`.
  Does not expose `public_status_enabled`. ✅ Safe.
- **Fix:** Added a Fiber middleware on the `/api/v1` group (before all data routes,
  after the `/settings` route) that calls `d.Settings.Get(ctx)` and returns
  `404 {"error":"status page disabled"}` when `PublicStatusEnabled == false`. The
  `/api/v1/settings` route is registered before this middleware so it is always reachable.

### N5 — Per-service visibility leakage — 🔧 Fixed (High)
- **Finding:** Four endpoints exposed data for `public_visible = false` services by UUID:
  1. `GET /api/v1/services/:id/uptime` — no visibility check in `UptimeService.GetServiceUptime`
  2. `GET /api/v1/monitors/:id/latency` — no parent-service visibility check
  3. `GET /api/v1/services/:id/badge.svg` — `BadgeHandler` calls bare `GetByID`, no public_visible filter
  4. `GET /api/v1/pages/:slug` + `GET /api/v1/by-domain` — SQL queries (`GetServiceBySlug`, `GetServiceByCustomDomain`) had no `dedicated_page_enabled = true` or `public_visible = true` filter
  - Files: `v1/uptime.go:18-25`, `v1/uptime.go:37-45`, `v1/badge.go:19-23`, `db/queries/services.sql:32-36`
- **Fix:**
  - `service/uptime.go:GetServiceUptime` — after `GetByID`, returns `ErrNotFound` if `!svc.PublicVisible`
  - `GetMonitorLatency` — added parent-service lookup + visibility check before returning latency
  - `v1/badge.go` — checks `svc.PublicVisible` before rendering SVG
  - `db/queries/services.sql` — added `AND dedicated_page_enabled = true` to `GetServiceBySlug`
    and `AND public_visible = true` to `GetServiceByCustomDomain`; ran `sqlc generate`

### N6 — New check types SSRF — 🔧 Fixed (High)
- **Finding:** `pkg/netguard` exists and blocks internal IPs / cloud metadata endpoints,
  but is called by ZERO check implementations. All five check types (`http`, `tcp`,
  `ssl`, `keyword`, `dns`) bypass it entirely. The runner dispatches without any
  pre-check gate. An admin-created monitor pointing at `169.254.169.254:80` (AWS
  metadata) or `127.0.0.1` would be fetched from the worker.
  - Files: `checks/ssl.go:29`, `checks/keyword.go:34`, `checks/http.go:30`,
    `checks/runner.go` (no gate)
- **Note:** `checks/dns.go` uses system resolver with no user-supplied resolver, so DNS
  check risk is lower — but the `Host` field is still admin-supplied.
- **Fix:** Added a `netguard.CheckPublicURL` / `CheckPublicHost` call at the top of
  `Run()` in `runner.go` before dispatching to type-specific implementations, applying
  consistently to all check types.

### N7 — SVG badge & RSS injection — 🔧 Fixed (⚠️ Medium for RSS; ✅ Badge)
- **Badge SVG:** ✅ Status value comes from server-side enum `{operational, degraded,
  outage, down, maintenance}` only — no admin free-text path. SVG served as
  `image/svg+xml`. No injection path. Safe.
- **RSS CDATA injection:** Admin-supplied `inc.Title` and `cfg.SiteTitle` are embedded
  directly in `<![CDATA[%s]]>` blocks via `fmt.Sprintf`. A title containing `]]>` closes
  the CDATA section and injects arbitrary XML into the feed. Test case:
  `inc.Title = "foo]]><inject/>"` → `<![CDATA[foo]]><inject/>]]>` — valid broken XML.
  Additionally, `cfg.SiteTitle` is embedded RAW (not in CDATA) in the `<description>`
  tag on one line, with no escaping at all.
  - File: `backend/internal/handler/v1/rss.go:40-55`
- **Fix:** Added a `cdataEscape` helper that replaces `]]>` with `]]]]><![CDATA[>` for
  CDATA-wrapped values; used `html.EscapeString` for the raw `<description>` embedding.
  Content-Type is correctly `application/rss+xml` (not text/html — no browser XSS risk).

### N8 — SQL interval string interpolation — 🔧 Fixed (Medium)
- **Finding:** `db/queries/check_results.sql` used string concatenation to build interval
  expressions: `($2 || ' days')::interval` and `($2 || ' hours')::interval` with `$2`
  as a `*string`. The `%d` format in `repository/postgres/check_result.go` prevented
  non-integer values from reaching the DB in practice, but the SQL pattern was
  architecturally unsound — relying on `::interval` cast as a secondary parsing barrier
  rather than using typed bind parameters. The `hours` value was hardcoded to `24` (not
  user-supplied), but `rangeDays` derives from a DB-stored int32 with a floor but no
  validated ceiling on the `GetServiceUptime` code path.
  - Files: `db/queries/check_results.sql:26,40`, `repository/postgres/check_result.go:63-64,93-94`
- **Fix:** Rewrote both queries to use typed integer multiplication:
  `($2::int * INTERVAL '1 day')` and `($2::int * INTERVAL '1 hour')`. Updated sqlc
  params from `*string` to `int32`. Ran `sqlc generate`.

### N9 — Custom domain / dedicated page host hardening — 🔧 Fixed (High)
- **Finding:** `frontend/proxy.ts` exported a function with the correct matcher
  `["/admin/:path*"]` but is NOT `frontend/middleware.ts` — Next.js middleware must be
  at that exact filename. `proxy.ts` was dead code; no edge middleware executed at all.
  On a custom domain, `/admin/*` and `/login` were fully reachable.
  - File: `frontend/proxy.ts` (wrong filename → dead code)
- **Fix:** Renamed `proxy.ts` to `middleware.ts`. Verified the matcher config and Host
  header logic; the file already had the correct implementation — it only needed the
  right filename.

### N10 — Audit coverage of new write routes — ✅ Verified safe
- All new write routes are inside the `admin` group which has `auditMw` applied at the
  group level (lines 84-89 of router.go). No new route was outside the audit chain.
- `auditFieldNames` records only top-level JSON key names (`map[string]json.RawMessage`
  → keys only). For `{"password":"secret"}`, the diff stores `["password"]` — key name
  only, value never stored. 2FA endpoints verified.
- Minor: for `POST /admin/incidents/:id/updates`, `parsePath` records resource=`"incidents"`,
  resourceID=incident UUID, dropping the `/updates` sub-segment. Acceptable.

### N11 — Settings cache staleness as a control bypass — ✅ Verified safe / 📝 Documented
- `SettingsService.Update` sets `s.cached = nil` immediately (under write lock) after
  the DB write succeeds. The next `Get` call on the same instance hits DB immediately.
- The 30-second TTL staleness window affects OTHER replicas in a multi-instance deployment
  — they will serve stale `public_status_enabled = false` for up to 30s. This is an
  accepted design decision (documented here). Operators disabling the page in a
  multi-replica setup should expect up to 30s propagation delay.

### N12 — Regression of v1.0 RBAC fixes — ✅ Verified / ⚠️ Design issue (Medium)
- All 25 expected write routes in the admin group correctly have `RequireAdmin`. The
  router rewrite introduced no regressions. `api-keys` and `2fa/setup`+`confirm` are
  intentionally ungated for self-service (by design).
- **Design issue — `2fa/disable` locked to admins:** `POST /admin/2fa/disable` has
  `RequireAdmin` (router.go:184). This creates an asymmetry: any user (viewer) can enroll
  2FA via `setup`+`confirm`, but cannot disable their own 2FA — only an admin can. A
  viewer who loses their authenticator app is permanently locked in 2FA with no
  self-service escape. **Fix:** Remove `RequireAdmin` from `2fa/disable`; add ownership
  enforcement inside the handler (the route already reads `userIDFromCtx` and passes it
  to `tf.Disable`, so disabling another user's 2FA is architecturally impossible —
  removing the admin gate is safe).
  - File: `backend/internal/handler/router.go:184`

---

## 3. Execution Log

| Check | Result |
|---|---|
| `go build ./...` | ✅ Clean (after all backend fixes) |
| `go vet ./...` | ✅ Clean (after user_test.go TOTP stub fix) |
| N1 — 2FA login enforcement (backend + frontend) | 🔧 Fixed — backend withholds cookie; Verify issues it; login page has 2-step UI |
| N2 — 2FA verify rate limit / backup single-use | 🔧 Fixed — per-account `"2fa:<id>"` limiter added to Verify handler; backup code consume is now atomic `:execrows` |
| N3 — secret/backup-code non-exposure (`/auth/me`, user DTO, audit diff) | ✅ Safe |
| N4 — public settings gate on all `/api/v1/*` | 🔧 Fixed — group-level middleware added |
| N5 — per-service visibility & archived filtering | 🔧 Fixed — visibility checks added to uptime/latency/badge/slug/domain paths |
| N6 — SSRF on ssl/keyword/dns checks vs `pkg/netguard` | 🔧 Fixed — netguard applied in runner.go before dispatch |
| N7 — SVG badge / RSS CDATA injection + content-type | 🔧 Fixed — cdataEscape + html.EscapeString on RSS; badge safe (enum values only) |
| N8 — dynamic interval/range SQL interpolation | 🔧 Fixed — rewritten to typed integer bind params; sqlc regenerated |
| N9 — custom-domain host hardening (middleware.ts) | 🔧 Fixed — proxy.ts renamed to middleware.ts |
| N10 — audit coverage of new write routes, no secret values | ✅ Safe |
| N11 — settings cache invalidation on Update | ✅ Safe / 30s multi-replica staleness documented |
| N12 — regression of S1–S20 RBAC | ✅ All routes gated; 2fa/disable design issue documented (⚠️ Medium) |
| Live E2E (docker compose) | ⏳ |
