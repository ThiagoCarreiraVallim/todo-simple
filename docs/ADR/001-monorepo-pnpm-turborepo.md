# ADR 001 — Monorepo with pnpm + Turborepo (including a Go app)

**Status:** Accepted
**Date:** 2026-07-15

## Context

The web app (Next.js) and the API (Go) are developed together and deployed
together. Turborepo is a Node-oriented tool, but it can orchestrate *any*
language as long as each package exposes its tasks through a `package.json`
with `scripts` — it just shells out to them and caches by declared
inputs/outputs. `apps/api` has a `package.json` whose scripts call `go`
directly (`go run`, `go build`, `go test`, `gofmt`/`go vet`); there is no
Node code in that package.

## Decision

A single monorepo using pnpm workspaces + Turborepo, with `apps/api` (Go) as
a workspace member alongside `apps/web` (Next.js), even though only the
latter has JS dependencies.

## Consequences

### Positive
- One command (`pnpm dev` / `pnpm build` / `pnpm lint`) drives both apps
- Turborepo's task caching still applies to the Go app's `build`/`lint`/`test` outputs
- Docker/CI scripts, env files, and onboarding docs are shared across both apps

### Negative
- `pnpm install` at the root does nothing for `apps/api` — Go dependencies are
  managed separately via `go.mod`/`go mod tidy`, which can confuse contributors
  expecting a single install step
- Anyone unfamiliar with this pattern may not expect a Go module inside a
  pnpm workspace

### Neutral
- `apps/api/package.json` has no `dependencies` — it exists purely so Turborepo can see and orchestrate the package

## Alternatives considered
- **Two separate repos (Go API, Next.js web):** avoids the mixed-workspace oddity, but splits CI, env management, and docs across repos for two halves of one product.
- **Makefile / plain shell scripts instead of Turborepo:** simpler mentally, but loses caching and the single `turbo run <task>` entrypoint across both apps.
