---
title: JojoOnePieceSimulator2 - ADR
tags:
  - project
  - adr
  - jojo-onepiece-simulator
---

# JojoOnePieceSimulator2 — ADR

## Status
Solo hobby/learning project. CI/CD to prod set up 2026-08-02 (see [[cicd-deployment]]): single server reachable via an external Docker network `public-net` (shared across services on that host, created outside this repo), Nginx Proxy Manager on that same network fronting `jojo-one-piece-simulator.duckdns.org`. All 4 compose services join `public-net`; local dev is Docker-only (see [[docker-setup]]).

## Stack
- Backend: Go, hexagonal/clean architecture (`internal/domain`, `internal/application`, `internal/infrastructure`), sqlc + Postgres, Chi-style router, Swagger docs, Redis-like `ICache` port.
- Domain: Stands, DevilFruits, Powers, Users/Auth (Google OAuth).
- Picture pipeline: libvips-based image processing (thumbnails/renditions), async status (`PictureStatus`), stored via a picture-storage port.
- Frontend: Expo/React Native + Tamagui, TanStack Query, secure-storage session store.
- Deploy: docker-compose (base + dev/prod overrides) + nginx (frontend), CI in `.github/workflows/ci.yml` (PR→main, required check), CD in `.github/workflows/cd.yml` (push→main, Tailscale+SSH) — see [[cicd-deployment]].

## Known weak spots (flagged by owner, not derivable from code)
- Picture pipeline (vips) is fragile — CI must build/test with `-tags vips`, easy to break with a plain `go test ./...` step. **Fixed 2026-08-02**: `.github/workflows/ci.yml`'s backend job now runs both `go test ./...` and `go test -tags vips ./...` (libvips-dev installed on the runner first).
- Auth/session flow is incomplete — Google login exists, rest of the session/auth lifecycle is WIP.

## Decisions worth remembering
- Solo project: docs are for the owner's own future reference, not team onboarding.
- Multi-language support (en-GB/es-ES/ca-ES) added 2026-08-06, backend + frontend infra done and verified; UI copy migration partial. See [[i18n-multi-language]].

## Repo stats (indexed 2026-07-28)
1601 nodes, 6484 edges. Languages: Go (104 files), TypeScript (21), YAML (7), SQL (7), JS (4).
