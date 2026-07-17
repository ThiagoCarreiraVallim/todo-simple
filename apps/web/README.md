# Web

Next.js (App Router) frontend. Talks to the Go API directly from the browser
via `NEXT_PUBLIC_API_URL` — there's no proxy route or auth token to hide in
this template, so no server-side indirection is needed.

```bash
pnpm --filter web dev
pnpm --filter web test
pnpm --filter web build
```

## Layout

- `src/app/page.tsx` — lists items from `GET /api/items` and lets you add/remove one, proving the web ↔ API ↔ Postgres wiring end to end.
- `src/lib/api.ts` — thin `fetch` wrapper (`apiFetch`) used by `src/lib/api/*`.
- `src/lib/api/items.ts` — typed wrapper around the `items` endpoints.
- `src/components/ui` — a couple of shadcn/ui-style primitives (`button`, `card`, `input`).

## CORS

The API only accepts requests from `WEB_ORIGIN` (see `apps/api/internal/server/server.go`).
Keep `WEB_ORIGIN` (API side) and `NEXT_PUBLIC_API_URL` (this app) in sync when
you change ports or domains.
