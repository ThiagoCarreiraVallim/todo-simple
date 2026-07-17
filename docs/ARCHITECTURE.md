# Architecture

## High-level view

```
┌───────────────────┐
│   Web (Next.js)    │
│   browser           │
└─────────┬───────────┘
          │ fetch (NEXT_PUBLIC_API_URL)
          │ CORS: WEB_ORIGIN
          │
┌─────────▼───────────┐
│      Go API           │
│ ┌──────────────────┐ │
│ │  chi router        │ │  HTTP handlers (decode/encode JSON)
│ └────────┬──────────┘ │
│ ┌────────▼──────────┐ │
│ │    Services         │ │  ← business logic
│ └────────┬──────────┘ │
│ ┌────────▼──────────┐ │
│ │   Repositories      │ │  plain SQL via pgx
│ └────────┬──────────┘ │
└──────────┼─────────────┘
           │
┌──────────▼─────────────┐
│     Postgres 16          │
└───────────────────────────┘
```

There's no auth in this template and no server-side proxy: the web app calls
the API directly from the browser, and the API allows only `WEB_ORIGIN` via
CORS (see `apps/api/internal/server/server.go`).

## Stack

### Backend (`apps/api`)

- **Go 1.22**, stdlib `net/http` + [chi](https://github.com/go-chi/chi) for routing
- **pgx/v5** (`pgxpool`) — no ORM, repositories write plain SQL
- **golang-migrate**, migrations embedded via `go:embed` (see [ADR 002](./ADR/002-go-stdlib-chi-pgx.md))
- **`log/slog`** — structured logging, stdlib only

See [`apps/api/README.md`](../apps/api/README.md) for the package layout.

### Frontend (`apps/web`)

- **Next.js 15 (App Router)**
- **TanStack Query** for data fetching/caching
- **Tailwind CSS** + a couple of shadcn/ui-style primitives

See [`apps/web/README.md`](../apps/web/README.md).

## Adding a new domain (backend)

Follow the shape of `apps/api/internal/lists/`:

1. `model.go` — the Go struct + JSON tags.
2. `repository.go` — plain SQL against `*pgxpool.Pool`.
3. `service.go` — validation and orchestration; this is what you unit test without a DB.
4. `handler.go` — chi routes, JSON decode/encode via `internal/httpx`.
5. Wire it into `internal/server/server.go` (`r.Route("/api/<domain>", handler.Routes)`).
6. Add a migration under `internal/database/migrations/` for any new tables.

## Adding a new page (frontend)

1. `src/lib/api/<domain>.ts` — a thin wrapper around `apiFetch`.
2. `src/app/<route>/page.tsx` — the UI, using TanStack Query for data fetching.
