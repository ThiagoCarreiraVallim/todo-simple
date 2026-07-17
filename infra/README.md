# Infra

Compose files, Dockerfiles, and (implicitly) deployment notes.

## Development

```bash
# Postgres only (API + Web run on the host via `pnpm dev`)
pnpm infra:up

# API + Web + Postgres in containers (closer to prod, no hot reload)
docker compose -f infra/docker-compose.yml --profile full up -d

# Logs
pnpm infra:logs

# Tear down
pnpm infra:down
```

## Production

`infra/docker-compose.prod.yml` targets a Traefik-fronted host (Dokploy,
Coolify, or a bare Docker host with Traefik + an external network). Postgres
is expected to run outside this compose (a managed instance, or its own
container elsewhere).

1. Provision Postgres and point `DATABASE_URL` at it.
2. Set the env vars listed in `.env.example` in your platform's secrets store.
3. Deploy — the `api` binary applies migrations on boot (see
   `apps/api/cmd/api/main.go`).

Rename the `edge-network` network and the `app-*` Traefik labels/routers to
match your project name and reverse proxy setup.
