# API

A small Go HTTP service (stdlib `net/http` + [chi](https://github.com/go-chi/chi)
for routing) backed by Postgres via [pgx](https://github.com/jackc/pgx).

```bash
pnpm --filter api dev       # go run ./cmd/api (needs Postgres up + DATABASE_URL)
pnpm --filter api test      # go test ./...
pnpm --filter api build     # go build -o bin/api ./cmd/api
pnpm --filter api migrate   # apply pending migrations
```

`go.sum` is committed and verified (`go vet`, `go build`, and `go test ./...`
all pass against it). Re-run `go mod tidy` whenever you add or change a
dependency.

## Layout

- `cmd/api` — entrypoint: loads config, runs migrations, starts the HTTP server with graceful shutdown.
- `cmd/migrate` — small CLI wrapping the embedded migrations (`go run ./cmd/migrate [-down]`).
- `internal/config` — env-based config, loaded from `../../.env` in development.
- `internal/database` — `pgxpool` connection pool + embedded SQL migrations (`go:embed`).
- `internal/httpx` — tiny JSON response helpers shared by handlers.
- `internal/health` — public `GET /health`, pings the DB.
- `internal/items` — example domain: model + repository (SQL) + service (validation) + HTTP handlers, mounted at `/api/items`. Copy this package's shape for new domains.

## Adding a migration

Add a pair of files to `internal/database/migrations/`, following
`golang-migrate`'s naming convention:

```
NNNNNN_description.up.sql
NNNNNN_description.down.sql
```

They're embedded into the binary at build time — no separate migration
runner or file copy needed in Docker.
