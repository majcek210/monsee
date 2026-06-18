# Tests.md — Full Functional Test Plan & Results (v1.01 / Phases 10–20)

Scope: every backend endpoint and every frontend button/interaction, verifying the app
**works** (not just "is secure" — see `Security.md` for the security-focused review).
This pass covers the v1.01 surface: settings/branding, per-service visibility & dedicated
pages, new check types (ssl/keyword/dns), maintenance windows, 2FA, incident timeline,
badge/RSS, manual status override, audit-log viewer, and latency sparklines — plus a
regression sweep of the Phase 1–9 surface.

Status legend: ✅ Pass · ❌ Fail (bug found) · 🔧 Fixed during this session · ⏳ Not yet executed · ⚠️ Pass with caveat

---

## 1. Plan

### 1.1 Backend

1. Static checks: `go build ./...`, `go vet ./...`.
2. Unit tests (stdlib `testing`, no testify):
   - `internal/service/settings` — cache hit/miss, 30s TTL expiry, invalidation on Update, `site_title != ""` validation.
   - `internal/service/uptime` — daily-uptime aggregation math (up/degraded/down → status + percent), range-days default, maintenance exclusion.
   - `internal/service/maintenance` — `ends_at > starts_at` validation, service-exists check, active-window detection.
   - `internal/service/twofactor` — TOTP confirm/verify, backup-code consumption (single-use), disable requires correct password.
   - `internal/service/incident` — `PostUpdate` status whitelist, resolved-status side effect, `ListUpdates` ordering.
   - `internal/service/svc` — slug regex validation, dedicated-page-requires-slug rule, status_override whitelist, slug/custom_domain clear-on-archive.
   - `internal/checks/ssl` — expired → down, near-expiry → degraded, valid → up (use `httptest`/known certs where feasible).
   - `internal/checks/keyword` — contains/not-contains vs `should_exist`, body read limit.
   - `internal/checks/dns` — record-type dispatch, expected-value match, no-records → down.
3. Endpoint inventory — walk `router.go`, confirm method/path/auth/role match the v1.01 spec and frontend usage.
4. Live E2E (docker compose: postgres + redis + backend + worker + frontend) — exercise the new flows with `curl` + DB checks.

### 1.2 Frontend

1. `npm run build`, `tsc --noEmit` (lint unavailable on Next 16 — see §5).
2. Page-by-page button/interaction inventory for every admin page + public pages — every interactive element, the API call it triggers, and whether that call is implemented/valid on the backend.
3. Manual run against the live backend for the new flows (settings save, 2FA enroll, maintenance schedule, incident timeline post, audit-log paging, uptime/latency rendering).

---

## 2. Backend Endpoint Inventory

### 2.1 Auth
| Method | Path | Auth | Role | Frontend uses it? | Status |
|---|---|---|---|---|---|
| GET | `/health` | none | — | — | ✅ |
| POST | `/auth/login` | none | — | yes | ✅ |
| POST | `/auth/logout` | none | — | yes | ✅ |
| GET | `/auth/me` | session | any | yes (sidebar role gate, security page) | ✅ (includes `totp_enabled`) |
| POST | `/auth/2fa/verify` | none | — | yes (login second step) | ✅ (backend exists, frontend doesn't call it — see F5) |

### 2.2 Admin — core resources
| Method | Path | Auth | Role | Frontend uses it? | Status |
|---|---|---|---|---|---|
| GET | `/admin/services` | session | any | yes | ✅ |
| POST | `/admin/services` | session | admin | yes | ✅ |
| GET | `/admin/services/:id` | session | any | yes | ✅ |
| PATCH | `/admin/services/:id` | session | admin | yes | ✅ |
| DELETE | `/admin/services/:id` | session | admin | yes | ✅ |
| GET | `/admin/services/:id/uptime` | session | any | yes (service detail, if wired) | ✅ |
| GET | `/admin/uptime` | session | any | yes (Overview page) | ✅ |
| GET/POST/PATCH/DELETE | `/admin/monitors*` | session | admin (writes) | yes | ✅ |
| GET | `/admin/monitors/:id/latency` | session | any | yes (sparkline) | ✅ |
| GET/POST/PATCH/DELETE | `/admin/incidents*`, `/resolve` | session | admin (writes) | yes | ✅ |
| GET | `/admin/incidents/:id/updates` | session | any | yes (timeline) | ✅ |
| POST | `/admin/incidents/:id/updates` | session | admin | yes (post update) | ✅ |
| GET/POST/DELETE | `/admin/api-keys*` | session | any (own) / owner-or-admin | yes | ✅ |
| GET/POST/PATCH/DELETE | `/admin/notifications*` | session | admin (writes) | yes | ✅ |
| GET/POST/PATCH/DELETE/logs | `/admin/webhooks*` | session | admin (writes) | yes | ✅ |
| GET/POST/PATCH/DELETE | `/admin/users*` | session | admin | yes | ✅ |

### 2.3 Admin — v1.01 additions
| Method | Path | Auth | Role | Frontend uses it? | Status |
|---|---|---|---|---|---|
| GET | `/admin/settings` | session | any | yes (settings page) | ✅ |
| PATCH | `/admin/settings` | session | admin | yes | ✅ |
| GET | `/admin/maintenance-windows` | session | any | yes (maintenance page) | ✅ |
| POST | `/admin/maintenance-windows` | session | admin | yes | ✅ |
| GET | `/admin/maintenance-windows/:id` | session | any | (detail/edit) | ✅ |
| PATCH | `/admin/maintenance-windows/:id` | session | admin | (edit) | ✅ |
| DELETE | `/admin/maintenance-windows/:id` | session | admin | yes (archive) | ✅ |
| POST | `/admin/2fa/setup` | session | any | yes (security page) | ✅ |
| POST | `/admin/2fa/confirm` | session | any | yes | ✅ |
| POST | `/admin/2fa/disable` | session | admin | yes | ✅ |
| GET | `/admin/audit-log` | session | admin | yes (audit-log page) | ✅ |

### 2.4 Public REST API (`/api/v1/*` — always registered; settings-gated where noted)
| Method | Path | Auth | Gate | Frontend uses it? | Status |
|---|---|---|---|---|---|
| GET | `/api/v1/settings` | none | always (exempt) | yes (branding, enabled flag) | ✅ |
| GET | `/api/v1/status` | none | public_status_enabled | yes (status page) | ✅ (F1 fix verified live — `public_visible=false` service confirmed absent from response) |
| GET | `/api/v1/incidents` | none | public_status_enabled | yes | ✅ |
| GET | `/api/v1/incidents/:id` | none | public_status_enabled | yes (public incident page) | ✅ (returns `{incident, updates}`) |
| GET | `/api/v1/incidents/:id/updates` | none | public_status_enabled | (timeline) | ✅ |
| GET | `/api/v1/uptime` | none | public_status_enabled | (aggregate) | ✅ |
| GET | `/api/v1/services/:id/uptime` | none | public_status_enabled | yes (uptime bars) | ✅ daily up/down/percent buckets correct |
| GET | `/api/v1/monitors/:id/latency` | none | public_status_enabled | (sparkline) | ✅ 123 points returned for test monitor |
| GET | `/api/v1/pages/:slug` | none | public_status_enabled + dedicated page | yes (dedicated page) | ✅ |
| GET | `/api/v1/by-domain?domain=` | none | public_status_enabled + custom domain | yes (custom-domain routing) | ✅ |
| GET | `/api/v1/services/:id/badge.svg` | none | — | embeddable badge | ✅ valid SVG, `image/svg+xml` |
| GET | `/api/v1/rss` | none | — | incidents feed | 🔧 Fixed (F9 — `<link>` was malformed) |

> **Settings-gate check (cross-ref Security.md):** confirm `/api/v1/status`, `/incidents*`,
> `/uptime*`, `/pages/*`, `/by-domain` return `404 {"error":"status page disabled"}` when
> `public_status_enabled=false`, while `/api/v1/settings` still returns `200`.

---

## 3. Frontend Button / Interaction Inventory

### `/` public status page
| Element | Action | Backend call | Status |
|---|---|---|---|
| (page load, server) | resolve public-enabled + branding | `GET /api/v1/settings` (`cache: no-store`) | ✅ |
| (page load, client) | services + open incidents | `GET /api/v1/status`, `GET /api/v1/incidents?status=open` | ✅ |
| Branding (logo/title) | render from settings, fallback to `/monsee.png` + "monsee" | `usePublicSettings()` | ✅ |
| Incident link | navigate to `/incidents/[id]` | `GET /api/v1/incidents/:id` | ✅ |
| Footer "Powered by monsee" | hardcoded (must stay) | — | ✅ |

### `/incidents/[id]` public incident detail
| Element | Action | Backend call | Status |
|---|---|---|---|
| (page load) | incident + timeline | `GET /api/v1/incidents/:id` (returns `{incident, updates}`) | ✅ |
| Timeline render | reverse-chron updates, status colors | — | ✅ |
| "Back to status" | navigate to `/` | — | ✅ |

### `/login`
| Element | Action | Backend call | Status |
|---|---|---|---|
| Login form submit | authenticate | `POST /auth/login` | ✅ |
| 2FA second step (if enabled) | verify TOTP/backup code | `POST /auth/2fa/verify` | ❌ F5 — login page never shows 2FA step |
| (on success) | redirect to `/admin` | — | ✅ |
| (public disabled) | `/` redirects here | — | ✅ |

### `/admin` (Overview)
| Element | Action | Backend call | Status |
|---|---|---|---|
| (page load) | summary cards + per-service uptime | `useServices`, `useAllUptime` (`GET /admin/uptime`), `useIncidents` | ✅ |
| UptimeBar render | 90-day bars + tooltips | — | ✅ |
| Empty state | "No uptime data yet" | — | ✅ |

### `/admin/services` & `/admin/services/[id]`
| Element | Action | Backend call | Status |
|---|---|---|---|
| New/Edit/Archive service | CRUD | `POST/PATCH/DELETE /admin/services*` (admin) | ✅ |
| Service visibility fields (public_visible, slug, etc.) | form inputs | PATCH /admin/services/:id | ✅ (F4 frontend fix + F7 backend fix — both verified live) |
| New/Edit/Archive monitor | CRUD | `POST/PATCH/DELETE /admin/monitors*` (admin) | ✅ |
| Monitor type select (http/tcp/ssl/keyword/dns) | client form state + type-specific fields | — | ✅ (F2 fix) |
| SSL fields (expiry threshold) | form → create/update | monitor create/update | ✅ (F2 frontend + F8 backend fix — live test: expired.badssl.com correctly flagged `down`) |
| Keyword fields (match, should-exist) | form → create/update | monitor create/update | ✅ (live test: example.com keyword match correctly `up`) |
| DNS fields (record type, expected value) | form → create/update | monitor create/update | ✅ (live test: mismatched expected value correctly `down`) |

### `/admin/incidents` & `/admin/incidents/[id]`
| Element | Action | Backend call | Status |
|---|---|---|---|
| New incident / Resolve | CRUD | `POST /admin/incidents`, `POST /admin/incidents/:id/resolve` | ✅ |
| Row → incident detail | navigate to `[id]` | `GET /admin/incidents/:id` | ❌ F3 — list rows have no link; detail page exists but unreachable |
| Timeline load | list updates | `GET /admin/incidents/:id/updates` | ✅ (backend) / ❌ unreachable (frontend) |
| "Post Update" → dialog → submit | post status+message | `POST /admin/incidents/:id/updates` (admin) | ✅ (backend) / ❌ unreachable (frontend) |

### `/admin/maintenance`
| Element | Action | Backend call | Status |
|---|---|---|---|
| (page load) | list windows, group active/upcoming/past | `GET /admin/maintenance-windows` | ✅ |
| "New Window" → dialog → Schedule | create | `POST /admin/maintenance-windows` (admin) | ✅ |
| datetime-local → ISO conversion | client → RFC3339 body | — | ✅ |
| Row trash → archive | archive | `DELETE /admin/maintenance-windows/:id` (admin) | ✅ |
| Active-window detection | client time math | — | ✅ |

### `/admin/security` (2FA)
| Element | Action | Backend call | Status |
|---|---|---|---|
| "Enable 2FA" → setup | get secret + otpauth URI | `POST /admin/2fa/setup` | ✅ |
| Confirm code → verify+enable | confirm, show backup codes | `POST /admin/2fa/confirm` | ✅ |
| Backup codes display | one-time render | — | ✅ |
| "Disable" → password → confirm | disable | `POST /admin/2fa/disable` (admin) | ✅ |
| Status reflects `totp_enabled` | from `GET /auth/me` | ⚠️ F6 — reads correct field via unsafe cast; `totp_enabled` not in `User` type | ⚠️ |

### `/admin/audit-log`
| Element | Action | Backend call | Status |
|---|---|---|---|
| (page load) | paginated entries + total | `GET /admin/audit-log?limit&offset` (admin) | ✅ |
| Resource filter form | filter by resource | `?resource=` | ✅ |
| Prev/Next pagination | offset paging | `?offset=` | ✅ |
| Diff "fields" render | field names only (no values) | — | ✅ |

### `/admin/settings`
| Element | Action | Backend call | Status |
|---|---|---|---|
| (page load) | load current settings | `GET /admin/settings` | ✅ |
| Site Title input | form state | — | ✅ |
| Logo URL input + `<img>` preview | live external image (CSP `https:`) | — | ✅ |
| Public status `Switch` | toggle | (in save body) | ✅ |
| Save | persist + invalidate cache | `PATCH /admin/settings` (admin) | ✅ |

### `/admin/notifications`, `/admin/webhooks`, `/admin/api-keys`, `/admin/users`
| Element | Action | Backend call | Status |
|---|---|---|---|
| (regression) full CRUD + toggles + logs | as Phase 1–9 | unchanged routes | ✅ |

### Sidebar
| Element | Action | Status |
|---|---|---|
| New nav items: Overview, Maintenance, Security, Audit Log, Settings | render + active-state | ✅ |
| `adminOnly` filter (Users, Audit Log, Settings) | hidden for viewers | ✅ |
| Overview `exact` active match | only active on `/admin` exactly | ✅ |

---

## 4. Detailed Functional Findings

### B1 — go vet: mockUserRepo missing TOTP interface methods — 🔧 Fixed
- **Where:** `backend/internal/service/user_test.go:114`
- **Evidence:** `go vet` reported `*mockUserRepo does not implement domain.UserRepository (missing method DisableTOTP)`. Five TOTP methods were added to `UserRepository` interface in v1.01 (`GetTOTP`, `SetTOTPSecret`, `EnableTOTP`, `DisableTOTP`, `RemoveBackupCode`) but the test mock was never updated.
- **Fix:** Added five no-op stub methods to `mockUserRepo` in `user_test.go`. `go vet ./...` now passes cleanly.

### F1 — `/api/v1/status` returns all services regardless of `public_visible` — 🔧 Fixed
- **Where:** `backend/internal/handler/v1/status.go:34`, `backend/internal/service/svc.go`, `backend/db/queries/services.sql`
- **Evidence:** `GetStatus` calls `h.services.List(ctx)` → `MonitoringService.List` → `s.services.List(ctx)` → SQL `SELECT * FROM services WHERE archived_at IS NULL`. No `public_visible = true` filter. Also: `EffectiveStatus()` (status_override + maintenance) is never called — raw `svc.Status` is returned.
- **Impact:** Services the admin intends to hide from the public page are exposed to anyone hitting the public API.
- **Fix:** Added `-- name: ListPublic` SQL query with `WHERE archived_at IS NULL AND public_visible = true`; added `ListPublic` to repo + service; updated `v1/status.go` to call the public variant and return `EffectiveStatus()`.

### F2 — Monitor form missing ssl/keyword/dns type options and fields — 🔧 Fixed
- **Where:** `frontend/app/admin/services/[id]/page.tsx` (type select, form state, submit payload), `frontend/lib/api/monitors.ts` (`CreateMonitorInput`)
- **Evidence:** Type select only has `http` and `tcp`. `defaultForm` and submit payload have no `ssl_expiry_threshold_days`, `keyword_match`, `keyword_should_exist`, `dns_record_type`, `dns_expected_value`. `CreateMonitorInput` also lacked these 5 fields.
- **Impact:** Admin cannot create/configure ssl, keyword, or dns monitors from the UI — the backend check engine has full support but no UI path to activate it.
- **Fix:** Added all 5 types to the select; added conditional field groups per type; extended `CreateMonitorInput` with the 5 new optional fields.

### F3 — Incident list rows have no link to the detail/timeline page — 🔧 Fixed
- **Where:** `frontend/app/admin/incidents/page.tsx` (incident row render)
- **Evidence:** Each incident row is a plain `<Card>` with only a "Resolve" button — no `<Link>` and no `onClick` navigate. `app/admin/incidents/[id]/page.tsx` exists and is fully implemented but is entirely unreachable from the UI.
- **Impact:** The incident timeline, "Post Update" dialog, and status update flow are inaccessible.
- **Fix:** Wrapped the card content area in `<Link href={/admin/incidents/${inc.ID}}>` with `cursor-pointer`.

### F4 — Service form missing all 7 v1.01 visibility/page fields — 🔧 Fixed
- **Where:** `frontend/app/admin/services/page.tsx` (create/edit dialog), `frontend/lib/api/services.ts`
- **Evidence:** Create/edit dialog only had `name` and `description`. `ServiceExtended` type (with `public_visible`, `show_uptime`, `dedicated_page_enabled`, `slug`, `custom_domain`, `uptime_range_days`, `status_override`) was defined in `types/index.ts` but unused by the form or API module.
- **Impact:** Admin cannot configure per-service visibility, dedicated pages, custom domains, or manual status override from the UI.
- **Fix:** Added all 7 fields to the service create/edit dialog and updated the `services.ts` API type.

### F5 — Login page has no 2FA second step — 🔧 Fixed
- **Where:** `frontend/app/login/page.tsx`
- **Evidence:** After `POST /auth/login` success, the page immediately calls `router.push("/admin/services")` with no check for `totp_enabled`. The `POST /auth/2fa/verify` endpoint exists on the backend but is never called from the login flow.
- **Impact:** 2FA is cosmetic — a user with 2FA enabled can log in without entering their TOTP code. (Cross-references Security.md N1 which also flags the backend side of this.)
- **Fix:** Added two-phase login UI: after successful password auth, if `totp_enabled === true`, shows a TOTP/backup-code input step; only redirects after `POST /auth/2fa/verify` succeeds.

### F6 — `User` type missing `totp_enabled` field — 🔧 Fixed
- **Where:** `frontend/types/index.ts` (`User` interface), `frontend/app/admin/security/page.tsx`
- **Evidence:** `User` type did not include `totp_enabled: boolean`. The security page worked around this with an unsafe cast `(me as { totp_enabled?: boolean } | undefined)?.totp_enabled`. TypeScript would not catch a backend shape mismatch.
- **Fix:** Added `totp_enabled: boolean` to the `User` interface; removed the cast in the security page.

### F7 — Service `Create`/`Update` handlers silently dropped all 7 v1.01 visibility fields — 🔧 Fixed
- **Where:** `backend/internal/handler/service.go` (`Create`, `Update`), `backend/internal/domain/service.go` (`CreateServiceParams`), `backend/internal/repository/postgres/service.go` (`Create`), `backend/internal/service/svc.go` (`Create`), `backend/db/queries/services.sql` (`CreateService`)
- **Evidence:** Found during live E2E — `POST /admin/services` and `PATCH /admin/services/:id` body structs only bound `name` + `description`. `domain.CreateServiceParams` had no fields for `public_visible`, `show_uptime`, `dedicated_page_enabled`, `slug`, `custom_domain`, `uptime_range_days`, `status_override` — only `UpdateServiceParams` had them. The `CreateService` SQL only inserted `(name, description)`, relying on column `DEFAULT`s with no way for the API caller to set them at creation time.
- **Impact:** Admin could not set any visibility/dedicated-page/custom-domain field when creating a service — only after a follow-up PATCH. Validation (slug regex, dedicated-page-requires-slug, status_override whitelist) also didn't run on create.
- **Fix:** Expanded `CreateService` SQL to accept all 7 fields via `sqlc.narg` with `COALESCE(..., <column default>)` so omitted fields still get the correct DB default (`public_visible`→true, `show_uptime`→true, `dedicated_page_enabled`→false, `uptime_range_days`→90); regenerated sqlc; added the 7 fields to `domain.CreateServiceParams`, the repo `Create`, and both handler body structs; mirrored `Update`'s slug/uptime-range/status-override/dedicated-page validation into `Create`. Verified live: `POST` with `public_visible`, `uptime_range_days` round-tripped correctly; `PATCH` with all 7 fields (including `slug`, `dedicated_page_enabled`, `status_override`) round-tripped correctly.

### F8 — Monitor `Create`/`Update` handlers silently dropped ssl/keyword/dns type-specific fields — 🔧 Fixed
- **Where:** `backend/internal/handler/monitor.go` (`Create`, `Update`)
- **Evidence:** Found during live E2E while testing the new check types. `domain.CreateMonitorParams`/`UpdateMonitorParams` and the repository layer already fully supported `ssl_expiry_threshold_days`, `keyword_match`, `keyword_should_exist`, `dns_record_type`, `dns_expected_value` — but the handler body structs only bound the original http/tcp fields, so these 5 fields never reached the domain layer despite F2 already having added them to the frontend form.
- **Impact:** Even after the F2 frontend fix, ssl/keyword/dns monitors created from the UI would have silently lost their type-specific config — the check engine would run with empty/zero values (e.g., DNS check with no expected value).
- **Fix:** Added the 5 fields to both handler body structs and the `CreateMonitorParams`/`UpdateMonitorParams` construction. Verified live: created one monitor per type (ssl against `expired.badssl.com`, keyword against `example.com`, dns against `example.com`) — worker picked them up within 15s and recorded correct results: SSL → `down` (certificate expired error), keyword → `up` (match found), DNS → `down` (expected value mismatch, proving the comparison logic runs end-to-end).

### F9 — RSS feed `<link>` used `c.Protocol()` which returns the HTTP version string, not the scheme — 🔧 Fixed
- **Where:** `backend/internal/handler/v1/rss.go:40`
- **Evidence:** `baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())`. In Fiber v3, `Ctx.Protocol()` returns the request's HTTP version string (`"HTTP/1.1"`), not the URL scheme. Live test showed `<link>HTTP/1.1://localhost</link>` instead of `<link>http://localhost</link>`, breaking every `<link>`/`<guid>` in the feed.
- **Impact:** RSS readers would fail to resolve incident links — malformed URLs in every feed item.
- **Fix:** Changed to `c.Scheme()`, which is Fiber v3's correct accessor for `http`/`https`. Verified live: feed now renders `<link>http://localhost</link>` and `<link>http://localhost/incidents/<id></link>` correctly.

---

## 5. Execution Log

| Step | Result |
|---|---|
| `go build ./...` | ✅ Clean |
| `go vet ./...` | 🔧 Fixed (B1 — added TOTP stubs to mockUserRepo) |
| `go test ./internal/service/... -v` | ✅ 33/33 pass |
| C3: `/auth/me` returns `totp_enabled` | ✅ Field present in domain, repo, handler |
| C6: public_visible filtering on `/api/v1/status` | 🔧 Fixed (F1) |
| ssl/keyword/dns check type files exist | ✅ `checks/ssl.go`, `checks/keyword.go`, `checks/dns.go` present; runner dispatches all 5 types |
| `tsc --noEmit` | ✅ Clean |
| `npm run build` | ✅ 14 routes compiled |
| C1: Monitor form ssl/keyword/dns | 🔧 Fixed (F2) |
| C2: Incident list → detail link | 🔧 Fixed (F3) |
| C5: Service visibility fields in form | 🔧 Fixed (F4) |
| Login 2FA second step | 🔧 Fixed (F5) |
| F6: `User` type `totp_enabled` | 🔧 Fixed (F6) |
| Live E2E (docker compose) | ✅ Complete — see results below |
| F7: Service Create/Update dropped v1.01 fields | 🔧 Fixed — discovered during live E2E |
| F8: Monitor Create/Update dropped ssl/keyword/dns fields | 🔧 Fixed — discovered during live E2E |
| F9: RSS `<link>` malformed (`c.Protocol()` → `c.Scheme()`) | 🔧 Fixed — discovered during live E2E |
| Post-fix regression: `go build`, `go vet`, `go test ./... -cover` | ✅ All clean (services/middleware/notifications/encrypt/hash/netguard packages pass) |

### Live E2E Results (docker compose: postgres + redis + backend + worker + frontend, backend on :9080)

| Flow | Result |
|---|---|
| Settings GET/PATCH + cache invalidation | ✅ PATCH persisted `site_title`, `logo_url`, `public_status_enabled`; immediate GET reflected new values |
| Public-disabled gate | ✅ With `public_status_enabled=false`: `/api/v1/status`, `/api/v1/incidents`, `/api/v1/uptime` → 404; `/api/v1/settings` → 200 (exempt) |
| Service v1.01 fields on Create + Update | ✅ Required F7 fix — see finding. Verified `public_visible`, `show_uptime`, `dedicated_page_enabled`, `slug`, `custom_domain`, `uptime_range_days`, `status_override` all round-trip on both `POST` and `PATCH` |
| `/api/v1/status` `public_visible` filtering (F1) | ✅ Confirmed live: toggling a service to `public_visible=false` removed it from the public response immediately |
| Public incidents API (`/api/v1/incidents`, `/:id`, `/:id/updates`) | ✅ Created incident + posted timeline update via admin API; all 3 public endpoints returned correct data, `/:id` nests `{incident, updates}` |
| Uptime aggregate + per-service + latency sparkline | ✅ `/api/v1/uptime` and `/api/v1/services/:id/uptime` returned correct daily up/down/percent buckets; `/api/v1/monitors/:id/latency` returned 123 historical points |
| Badge SVG + RSS feed | ✅ Badge: valid SVG, `image/svg+xml`. RSS: required F9 fix; after fix, feed renders correct `http://` links |
| Slug page resolution (`/api/v1/pages/:slug`) | ✅ Service configured with `slug=api-gateway`, `dedicated_page_enabled=true` → resolved correctly with nested uptime data |
| Custom-domain resolution (`/api/v1/by-domain`) | ✅ Service configured with `custom_domain=api.cryvex.xyz` → resolved correctly |
| 2FA enroll → login → verify | ✅ `POST /admin/2fa/setup` → secret; generated valid TOTP via `pquerna/otp`; `POST /admin/2fa/confirm` → 200 + 10 backup codes; fresh `POST /auth/login` returned `{totp_required:true}` with **no** session cookie; `POST /auth/2fa/verify` with a freshly generated code → 200 + session cookie issued. Disabled 2FA afterward to restore clean state |
| New check types (ssl/keyword/dns) | ✅ Required F8 fix — see finding. After fix: SSL monitor against `expired.badssl.com` → `down` (cert expired, correct error message); keyword monitor against `example.com` matching "Example Domain" → `up`; DNS monitor against `example.com` with a deliberately wrong expected A record → `down` (mismatch correctly detected), proving the comparison logic runs end-to-end |
| Maintenance window suppresses auto-incident | ✅ Created a monitor pointing at a closed port (`http://127.0.0.1:1/`, `retry_count=1`) plus an active maintenance window for its service. Monitor failed twice (`consecutive_failures` 1→2) with **zero** incidents created. Archived the maintenance window; next failure cycle created an incident (`status=open`, `severity=high`) — confirms the suppression in `checker.go`'s `IsActiveForService` check works and correctly stops applying once the window is archived |
| Audit-log paging/filter | ✅ `?limit=5&offset=0` vs `?offset=5` returned distinct entries with consistent `total=72`; `?resource=services` filter returned only `services`-resource entries; `diff` field confirmed to log field names only, never values (e.g. `{"fields":["name","public_visible"]}`) |
