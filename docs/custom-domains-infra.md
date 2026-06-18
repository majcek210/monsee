# Custom Domains — Infrastructure Setup

This document describes the container/DNS setup that makes customer custom
domains work. The **application** side (admin UI, the dedicated public page,
the `custom_domains_enabled` toggle, and the `/api/v1/tls-check` allowlist
endpoint) is already built. This is the remaining infra you deploy in Coolify.

> Nothing here is auto-deployed by the app. Apply it yourself.

---

## Topology

```
customer DNS:  status.theircompany.com  CNAME  proxy.cryvex.xyz
                                  │
                          (Cloudflare edge — see caveat)
                                  │
                          cloudflared tunnel  (outbound only, no open ports)
                                  │
                          Caddy  (terminates TLS, on-demand Let's Encrypt)
                                  │  asks backend: is this domain allowed?
                                  ├──────────────► GET /api/v1/tls-check?domain=<sni>
                                  ▼
                          Next.js frontend :3000  (Host-aware middleware)
                                  │  resolves the service by Host header
                                  ▼
                          renders /status/by-domain?domain=<host>
```

How a request flows once set up:

1. Visitor hits `https://status.theircompany.com`.
2. Caddy receives the TLS handshake. If it has no cert for that SNI, it calls
   the **ask** endpoint `GET http://backend:8080/api/v1/tls-check?domain=status.theircompany.com`.
   - `200` → Caddy obtains a Let's Encrypt cert on demand and caches it.
   - `404` → Caddy refuses (prevents cert issuance for arbitrary hostnames).
3. Caddy reverse-proxies to the Next.js frontend.
4. The frontend `middleware.ts` sees a non-primary `Host` on `/` and rewrites to
   `/status/by-domain?domain=status.theircompany.com`.
5. That page calls the backend `GET /api/v1/by-domain?domain=…`, which returns the
   service only if `custom_domains_enabled` is on AND a non-archived service has
   that `custom_domain` with `dedicated_page_enabled = true`.

---

## App-side prerequisites (already implemented)

- **Settings → Custom Domains** toggle must be **on** (`custom_domains_enabled`).
- The service must have **dedicated page enabled** + a **custom domain** set
  (service detail → Public Page tab).
- Backend endpoints (registered independently of the global public-status gate):
  - `GET /api/v1/tls-check?domain=<host>` → `200`/`404` allowlist for Caddy.
  - `GET /api/v1/by-domain?domain=<host>` → `{ service, uptime }`.
  - `GET /api/v1/pages/:slug` → `{ service, uptime }` (slug-based, same gate).
- Frontend env:
  - `APP_PRIMARY_HOST` (or `NEXT_PUBLIC_APP_HOST`) — the host the app normally
    serves on, e.g. `status.cryvex.xyz`. Any *other* host on `/` is treated as a
    custom domain. Without it, custom-domain rewriting is disabled.
  - `NEXT_PUBLIC_PROXY_HOST` — shown in the in-app DNS guide as the CNAME target
    (cosmetic), e.g. `proxy.cryvex.xyz`.

---

## Compose file: `compose.custom-domains.yml`

For plain `docker compose` deployments (not Coolify — see the Coolify section
below), the Caddy + cloudflared services are checked into the repo as an
**additive** compose file, not merged into `compose.yml`/`compose.prod.yml`
(the feature is opt-in infra, separate from the opt-in app-level toggle):

```bash
docker compose -f compose.yml -f compose.custom-domains.yml up -d
# or layered on prod the same way:
docker compose -f compose.prod.yml -f compose.custom-domains.yml up -d
```

It references a root-level `Caddyfile` (also checked in):

```caddyfile
{
    # Caddy calls this before issuing a cert for an unknown SNI.
    # Caddy automatically appends ?domain=<sni>.
    on_demand_tls {
        ask http://backend:8080/api/v1/tls-check
    }
}

# Custom domains (any host) — on-demand TLS, proxied to the frontend.
https:// {
    tls {
        on_demand
    }
    reverse_proxy frontend:3000
}
```

```yaml
# compose.custom-domains.yml
services:
  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on: [frontend, backend]

  cloudflared:
    image: cloudflare/cloudflared:latest
    restart: unless-stopped
    command: tunnel --no-autoupdate run
    environment:
      - TUNNEL_TOKEN=${CLOUDFLARED_TUNNEL_TOKEN}
    depends_on: [caddy]

volumes:
  caddy_data:
  caddy_config:
```

`caddy_data` **must** be a persistent volume — it stores issued certificates.
`CLOUDFLARED_TUNNEL_TOKEN` goes in your `.env` (see `env.example`) or as a
Coolify secret — get it from the Cloudflare Zero Trust dashboard when creating
the tunnel.

If exposing Caddy directly instead of fronting it with cloudflared, publish
ports `80`/`443` on the `caddy` service and skip the `cloudflared` service.

---

## Deploying on Coolify

Coolify's Docker Compose resources only support **one** compose file per
resource (the "Docker Compose Location" field), so the `-f file1 -f file2`
layering above doesn't apply here. Instead, the `caddy` and `cloudflared`
service blocks (and their `caddy_data`/`caddy_config` volumes) are merged
directly into `compose.coolify.yml` — the file your Coolify resource already
points at — so they land on the same Coolify-managed network as `backend`
and `frontend`, and the Caddyfile's `backend:8080` / `frontend:3000`
references resolve by service name.

Steps in the Coolify UI:

1. **Environment Variables tab** (not the General tab) — add
   `CLOUDFLARED_TUNNEL_TOKEN` with the token from the Cloudflare Zero Trust
   dashboard (Networks → Tunnels → create/select your tunnel → install
   connector). This is a plain secret, not a Coolify "magic" `SERVICE_*` var.
2. **General tab → Reload Compose File**, then redeploy. Coolify will detect
   the two new services and add "Domains for caddy" / "Domains for
   cloudflared" sections next to the existing ones for backend/worker/
   asynqmon/frontend.
3. **Leave those two domain fields blank — do not click Generate Domain for
   `caddy` or `cloudflared`.** Coolify's own domain/Traefik system is for
   services *you* want reached through Coolify's built-in reverse proxy
   (that's what the existing "Domains for frontend" entry,
   `https://<random>.cryvex.xyz:3000`, is — your normal way to reach the
   admin app). Customer custom domains are a completely separate path: they
   reach `caddy` only via the Cloudflare Tunnel, never through Coolify's
   Traefik.
4. In the **Cloudflare Zero Trust dashboard** (outside Coolify), set the
   tunnel's public hostname / ingress rule to route your proxy hostname
   (e.g. `proxy.cryvex.xyz`) to `https://caddy:443` — see the ingress example
   below. This is the step that actually connects a customer's CNAME target
   to this stack.
5. The `caddy` service publishes no host ports, so it doesn't conflict with
   Coolify's own Traefik listening on 80/443 for your other domains.

Tunnel ingress (configured in the Cloudflare Zero Trust dashboard, or a
`config.yml`) should route the proxy hostname to Caddy:

```yaml
ingress:
  - hostname: proxy.cryvex.xyz
    service: https://caddy:443
    originRequest:
      noTLSVerify: true   # Caddy presents on-demand certs; tunnel→Caddy is internal
  - service: http_status:404
```

Store `CLOUDFLARED_TUNNEL_TOKEN` as a Coolify secret.

---

## DNS

- **Primary app host** (`status.cryvex.xyz`): route via the tunnel as usual.
- **Proxy hostname** (`proxy.cryvex.xyz`): the CNAME target customers point at.
- **Customer side**: `status.theircompany.com  CNAME  proxy.cryvex.xyz`.

---

## ⚠️ Honest Cloudflare caveat (read before going live)

A plain tunnel + Caddy hides your IP cleanly **for hostnames in zones you
control**. For a truly *external* customer domain there's a constraint:

- If `proxy.cryvex.xyz` is **orange-clouded (proxied)** in your Cloudflare zone,
  an external customer CNAME to it returns **Error 1014 "CNAME Cross-User
  Banned."** Cloudflare only allows proxying someone else's domain through your
  zone via **Cloudflare for SaaS (Custom Hostnames)**.
- With the setup above, make the customer CNAME target **DNS-only / grey-cloud**
  so the request reaches Caddy and Caddy terminates TLS itself. This works and
  keeps the *tunnel* hiding your origin, but the customer domain is not behind
  Cloudflare's edge.

### Upgrade path: Cloudflare for SaaS

When you want edge TLS / DDoS / WAF on customer domains *and* full origin hiding:

1. Enable **Cloudflare for SaaS** on your zone.
2. Register each customer hostname as a **Custom Hostname** via the Cloudflare
   API when an admin saves a custom domain (hook into the service update flow).
3. Cloudflare issues + serves the cert at the edge and forwards to your origin
   (the tunnel). Caddy's on-demand TLS is then optional.

This is the only way to get external-domain edge-proxying with the IP hidden;
it's a paid feature, so it's left as a documented future step.

---

## Verification

1. Settings → Custom Domains is **on**.
2. A service has dedicated page + `custom_domain = status.theircompany.com`.
3. `curl -s -o /dev/null -w "%{http_code}" http://backend:8080/api/v1/tls-check?domain=status.theircompany.com` → `200`; unknown domain → `404`.
4. `dig status.theircompany.com` resolves to the proxy hostname.
5. `https://status.theircompany.com` loads the service's status page (first load
   may take a few seconds while the cert is issued).
