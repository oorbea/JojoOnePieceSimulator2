---
name: verify
description: Run the CI-equivalent checks (backend go build/vet/test, frontend typecheck/lint/test) in local Docker, only for whatever this branch actually touched. Use before calling any backend/frontend change done — the project norm (ObsidianVault/norma-verificacion-docker.md) is that a tanda isn't finished until this passes.
---

# /verify — CI-equivalent checks in local Docker

This project's norm (owner-set, `ObsidianVault/norma-verificacion-docker.md`): a plan or tanda is
only done once the checks CI would run pass in **local Docker**, and only the relevant ones —
backend-only changes never need the frontend suite, and vice versa.

## 1. Decide what's relevant

```
base=$(git merge-base HEAD origin/develop 2>/dev/null || git merge-base HEAD develop 2>/dev/null \
  || git merge-base HEAD origin/main 2>/dev/null || git merge-base HEAD main 2>/dev/null)
files=$( { git status --porcelain | awk '{print $2}'; git diff --name-only "$base" HEAD; } | sort -u)
```

- Any `apps/backend/**` in `$files` → run the backend checks (§2).
- Any `apps/frontend/**` in `$files` → run the frontend checks (§3).
- Neither → nothing to verify, say so and stop.

## 2. Backend — Docker, never host `go test`

Windows App Control blocks freshly-built `go test` binaries on this machine (see
`ObsidianVault/norma-verificacion-docker.md`, `feedback_backend_tests_via_docker` in Claude's own
memory). Always via Docker:

```
docker compose -f deployments/docker-compose.yml -f deployments/docker-compose.test.yml \
  run --rm backend-test sh -c "go build ./... && go vet ./... && go test ./..."
```

Run from the repo root (or `apps/backend` with `-f ../../deployments/...`). Redis-backed packages
(`internal/infrastructure/cache/redis`, `internal/infrastructure/gamestore/redis`) fail without
`db-up` running first — that's expected and not part of this check unless the diff actually touches
those packages, in which case bring `db-up` up first (see `docker-setup.md`).

## 3. Frontend — Docker, no dedicated compose service yet

No `frontend-test` service exists in `docker-compose.test.yml` today. The pattern that works
(2026-08-26, see `norma-verificacion-docker.md` for the full rationale and the copy-not-bind-mount
reasoning):

```bash
docker run --rm -v "$(pwd):/repo:ro" -v jojo-frontend-work:/work node:22-alpine sh -c '
  corepack enable && corepack prepare pnpm@11.18.0 --activate
  rm -rf /work/repo/frontend_new && mkdir -p /work/repo/frontend_new
  cp -r /repo/apps/frontend/. /work/repo/frontend_new/
  rm -rf /work/repo/frontend_new/node_modules
  [ -d /work/repo/frontend/node_modules ] && mv /work/repo/frontend/node_modules /work/repo/frontend_new/node_modules
  rm -rf /work/repo/frontend && mv /work/repo/frontend_new /work/repo/frontend
  cd /work/repo/frontend
  CI=true pnpm install --frozen-lockfile --prefer-offline
  pnpm typecheck && pnpm lint && pnpm test:ci
'
```

The named volume (`jojo-frontend-work`) caches `node_modules`/the pnpm store across runs so repeat
verifications in the same session are fast. **Never** bind-mount the real `apps/frontend` directly
into `pnpm install` — it would overwrite the Windows host's `node_modules` with Linux binaries.

If `pnpm test:ci` shows a handful of failures under full worker parallelism that pass in isolation
or with `pnpm jest --ci --maxWorkers=2`, that's the known Docker flake pattern (see
`norma-verificacion-docker.md`) — re-run once before treating it as a real regression.

## 4. Report

State plainly, per file/suite: build ✓/✗, vet ✓/✗, tests N/N, lint errors (warnings from this
Windows checkout's `core.autocrlf=true` are not failures — 0 errors is the bar). If anything failed,
show the actual output, don't summarize it away.
