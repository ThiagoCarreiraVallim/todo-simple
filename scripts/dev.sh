#!/usr/bin/env bash
# One-command dev: postgres (Docker) + api + web (host, hot reload for Web).
#
# Idempotent — `docker compose up -d` is a no-op if Postgres is already
# running. Once infra is healthy, hands off to `turbo run dev` in the
# foreground so logs from API and Web stream until you Ctrl-C.
#
# Run `pnpm bootstrap` first (once) to install deps and migrate.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

say() { printf '\033[1;36m▶ %s\033[0m\n' "$*"; }
warn() { printf '\033[1;33m! %s\033[0m\n' "$*" >&2; }
die() { printf '\033[1;31m✗ %s\033[0m\n' "$*" >&2; exit 1; }

command -v docker >/dev/null 2>&1 || die "docker not found. See docs/ONBOARDING.md."

if [ ! -f .env ]; then
  die ".env not found. Run 'pnpm bootstrap' first."
fi

say "Ensuring Postgres is up"
docker compose --env-file .env -f infra/docker-compose.yml up -d postgres >/dev/null

say "Waiting for Postgres"
for i in $(seq 1 30); do
  if docker exec todo-postgres pg_isready -U todo >/dev/null 2>&1; then
    break
  fi
  if [ "$i" -eq 30 ]; then
    die "Postgres did not become ready in 30s. Try 'pnpm infra:logs'."
  fi
  sleep 1
done

cat <<EOF

✓ Infra is up. Starting API + Web (Ctrl-C to stop).

  API:    http://localhost:8080
  Web:    http://localhost:3000

EOF

exec pnpm turbo run dev
