# Gogal fast runtime setup

## Goal

Keep Gogal lightweight, fast, and operationally simple.

## What to use now

### API runtime
- Use the compiled Go binary directly.
- Manage it with **Supervisor** or **systemd**.
- Do **not** use Gunicorn. Gunicorn is for Python WSGI apps and is not relevant for Gogal's Go API.

### Frontend runtime
- Build `UI Studio` with Vite.
- Serve the generated `frontend/dist/` as static files behind **Caddy**, **Nginx**, or **Traefik**.
- Do not keep `npm run dev` in production.

### Reverse proxy
Use one of:
- **Caddy** for easiest TLS and simple config
- **Nginx** for familiar static + reverse proxy setup
- **Traefik** if you already want dynamic multi-site routing later

### Database
- PostgreSQL is the primary store.
- Add **PgBouncer** only when you start seeing real connection churn or many concurrent app instances.

## Do we need Redis?

### Not mandatory right now
For the current platform state, Redis is **optional**.

### Add Redis when you implement
- background jobs / workers
- caching for metadata and heavy list queries
- rate limiting
- websocket/session coordination across multiple instances
- scheduled tasks / queue orchestration

So the fast path is:
- Go binary
- PostgreSQL
- Caddy/Nginx/Traefik
- Supervisor/systemd

And the next scale path is:
- add Redis
- optionally add PgBouncer
- run multiple Go API instances behind the reverse proxy

## Process management recommendation

### Lightweight default
- **API**: Supervisor or systemd
- **UI**: static files served by Caddy/Nginx

### Why this stays efficient
- Go already gives you native concurrency without Gunicorn-style worker orchestration.
- A single optimized Go binary is much lighter than a Python-style app stack.
- Static serving for UI Studio removes the need for a separate Node runtime in production.

## Suggested production shape

1. Build the Go API binary
2. Run it on `127.0.0.1:8080`
3. Build `frontend/dist`
4. Serve `dist` from `/opt/gogal/ui`
5. Use Caddy/Nginx/Traefik in front
6. Use Supervisor or systemd to restart the API automatically

## Files included in this repo

- `deploy/supervisor/gogal-api.conf` — Supervisor sample for the Go API
- `deploy/caddy/Caddyfile` — Simple reverse proxy + static UI example

## Next ops features worth adding later

- Redis-backed job queue
- dedicated worker binary/process
- metrics endpoint + Prometheus
- structured request logging
- health checks and readiness endpoints
- optional PgBouncer for pooled PostgreSQL connections
