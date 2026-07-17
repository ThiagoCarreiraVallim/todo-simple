# Onboarding

Get from `git clone` to a running API + Web stack in one command.

## Prerequisites

| Tool                        | Version | Notes                                                                        |
| ---------------------------- | ------- | ------------------------------------------------------------------------------ |
| **Node.js**                   | `>= 20` | `.nvmrc` is provided — run `nvm use`.                                          |
| **pnpm**                      | `9`     | `corepack enable` — version is pinned by `packageManager` in `package.json`. |
| **Go**                        | `>= 1.22` | <https://go.dev/doc/install>                                                |
| **Docker** + Docker Compose    | latest  | Used for Postgres (and optionally the API/Web containers).                    |
| **A POSIX shell**              | —       | macOS / Linux work out of the box. On Windows, use WSL2.                      |

Verify:

```bash
node -v          # v20.x
go version        # go1.22+
docker --version
docker compose version
```

## Quick start

```bash
git clone <your-repo-url>
cd <your-repo>
pnpm bootstrap
```

`pnpm bootstrap` runs `scripts/setup.sh`, which:

1. installs web dependencies (pnpm);
2. copies `.env.example` → `.env` if it doesn't exist;
3. starts the Postgres container (`infra/docker-compose.yml`);
4. waits for Postgres to be healthy;
5. runs `go mod tidy` in `apps/api` (resolves dependencies, generates `go.sum`);
6. applies database migrations.

Then start the app:

```bash
pnpm dev
```

- API: <http://localhost:8080>
- Web: <http://localhost:3000>

No further setup is needed — there's no auth provider to configure in this
template.

## Development modes

### Mode A — host `pnpm dev` + dockerized Postgres (default, fastest feedback)

```bash
pnpm dev   # postgres + api (go run) + web (next dev) — one command
```

If you want the pieces separately:

```bash
pnpm infra:up      # postgres only
pnpm dev:apps       # turbo run dev only (assumes infra is up)
```

### Mode B — full stack in Docker

```bash
pnpm infra:up:full   # postgres + api + web, all in containers
```

No hot reload; useful for reproducing a production-only bug or validating
Dockerfile changes. Tear down with `pnpm infra:down`.

## Common commands

```bash
pnpm db:migrate        # apply pending migrations
pnpm db:migrate:down   # roll back the last migration
pnpm lint              # gofmt/go vet (api) + eslint (web)
pnpm test              # go test (api) + vitest (web)
pnpm build             # turbo build, both apps
pnpm reset:db          # wipe the local DB volume and re-migrate (destructive)
```
