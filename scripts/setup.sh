#!/usr/bin/env bash
# One-command bootstrap for local development.
#
# Steps:
#   1. install workspace deps (pnpm)
#   2. copy .env from .env.example if missing
#   3. start Postgres via docker compose
#   4. wait until Postgres is healthy
#   5. tidy Go modules (generates go.sum on first run)
#   6. run database migrations
#
# Re-running is safe: install/migrate are idempotent. Use `pnpm reset:db`
# to wipe the database volume and re-migrate from scratch.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

say() { printf '\n\033[1;36m▶ %s\033[0m\n' "$*"; }
warn() { printf '\033[1;33m! %s\033[0m\n' "$*" >&2; }
die() { printf '\033[1;31m✗ %s\033[0m\n' "$*" >&2; exit 1; }

require() {
  command -v "$1" >/dev/null 2>&1 || die "Missing prerequisite: $1. See docs/ONBOARDING.md."
}

require node
require docker
require go

say "Enabling corepack (pins pnpm to the version in package.json)"
if command -v corepack >/dev/null 2>&1; then
  corepack enable
else
  warn "corepack not found — falling back to global pnpm. Install Node >= 20."
fi

say "Installing web dependencies"
pnpm install

if [ ! -f .env ]; then
  say "Creating .env from .env.example"
  cp .env.example .env
  warn ".env was just created from the example — review it before exposing the API publicly."
else
  say ".env already exists — leaving it alone"
fi

say "Starting Postgres (docker compose)"
pnpm infra:up

say "Waiting for Postgres to be ready"
for i in $(seq 1 30); do
  if docker exec todo-postgres pg_isready -U todo >/dev/null 2>&1; then
    printf '  ready after %ss\n' "$i"
    break
  fi
  if [ "$i" -eq 30 ]; then
    die "Postgres did not become ready in 30s. Try 'pnpm infra:logs'."
  fi
  sleep 1
done

say "Tidying Go modules"
(cd apps/api && go mod tidy)

say "Running database migrations"
pnpm db:migrate

cat <<EOF

✓ Bootstrap complete. Infra is running, DB is migrated.

  Next: one command starts everything (with hot reload for Web; Go restarts
  on each request via 'go run', no separate hot-reload tool wired in)

    pnpm dev            # postgres + API + Web

  Or, full Docker mode (no hot reload, closer to prod):

    pnpm infra:up:full

  URLs:
    API:    http://localhost:8080
    Web:    http://localhost:3000

  Other commands:
    pnpm infra:logs     # tail Postgres logs
    pnpm reset:db       # wipe the DB volume and re-migrate (destructive)

EOF
