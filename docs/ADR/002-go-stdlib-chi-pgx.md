# ADR 002 — Go backend: stdlib `net/http` + chi + pgx

**Status:** Accepted
**Date:** 2026-07-15

## Context

A Go HTTP backend needs: routing (path params, route groups), a Postgres
driver, and a migration mechanism. Go's ecosystem has strong, narrow-scope
libraries for each rather than one framework that does everything — that fits
a template meant to stay small and easy to read end to end.

## Decision

- **Routing:** [`chi`](https://github.com/go-chi/chi) on top of stdlib `net/http`. Adds route params, middleware chaining, and route groups without hiding the standard `http.Handler` contract.
- **Postgres driver:** [`pgx/v5`](https://github.com/jackc/pgx) with `pgxpool` for pooling. No ORM — repositories write plain SQL. For a template this keeps the query surface visible; swap in `sqlc` or an ORM later if the domain grows complex enough to want one.
- **Migrations:** [`golang-migrate`](https://github.com/golang-migrate/migrate), embedded via `go:embed` so migration files ship inside the compiled binary — no separate file copy step in Docker, no need to install the `migrate` CLI to run the app (a small `cmd/migrate` wraps it for explicit local use).
- **Logging:** stdlib `log/slog` — structured logging without adding a dependency.

## Consequences

### Positive
- Small dependency footprint; every dependency has one clear job
- `go:embed` migrations mean the Docker image is just one static binary
- Plain SQL in repositories is easy to read and reason about without ORM-generated query indirection

### Negative
- No ORM means more boilerplate as the schema grows (more `Scan()` calls, no auto-generated relations)
- `chi` is a routing library, not a framework — validation, request binding, etc. are done by hand in handlers

### Neutral
- Adding `sqlc` later is straightforward: point it at the same `migrations/` directory for its schema source

## Alternatives considered
- **A full framework (e.g. Gin, Echo, Fiber):** more built-in conveniences, but pulls in more indirection than a small template needs; stdlib + chi stays closer to what ships with Go.
- **GORM (ORM):** faster to prototype simple CRUD, but hides the SQL and adds real complexity for anything beyond basic queries.
- **`database/sql` directly instead of `pgx`:** more portable across drivers, but `pgx` has a friendlier API and better Postgres-specific type support (e.g. native UUID, JSONB).
