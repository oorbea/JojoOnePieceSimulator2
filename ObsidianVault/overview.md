---
title: JoJoOnePieceSimulator2 — Overview
tags:
  - project
  - jojo-onepiece-simulator
---

# JoJoOnePieceSimulator2

Repo: `C:\code\JojoOnePieceSimulator2`

Monorepo-ish layout, but **not** a real monorepo:
- `apps/backend` — Go/chi backend, complete
- `apps/frontend` — Expo PWA, standalone pnpm project (own lockfile, no root workspace)
- `deployments/` — Dockerfiles + docker-compose

See:
- [[backend-contract]] — API shape, auth, caching, enums
- [[frontend-stack]] — full frontend stack, config decisions, gotchas
- [[docker-setup]] — dev/prod images, compose wiring

## Status (2026-07-28)

Frontend scaffold complete and verified (`tsc`, `eslint`, `expo export -p web`, docker build+run all clean). No feature code yet — skeleton only, by explicit user choice.

## Flagged / not done

- No websocket endpoint on backend — `socket.io-client` installed but unwired.
- Google OAuth client IDs blank in `.env.example` — auth feature not built yet.
- No CI. `.github/cicd.yml` exists but empty and in wrong location (should be `.github/workflows/`). Must cover **both** backend (`-tags vips` build/test) and frontend when built.
