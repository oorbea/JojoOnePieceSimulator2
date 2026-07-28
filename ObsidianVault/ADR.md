---
title: JojoOnePieceSimulator2 - ADR
tags:
  - project
  - adr
  - jojo-onepiece-simulator
---

# JojoOnePieceSimulator2 — ADR

## Status
Solo hobby/learning project. Not deployed to prod yet — no target infra decided (docker-compose is dev-only for now).

## Stack
- Backend: Go, hexagonal/clean architecture (`internal/domain`, `internal/application`, `internal/infrastructure`), sqlc + Postgres, Chi-style router, Swagger docs, Redis-like `ICache` port.
- Domain: Stands, DevilFruits, Powers, Users/Auth (Google OAuth).
- Picture pipeline: libvips-based image processing (thumbnails/renditions), async status (`PictureStatus`), stored via a picture-storage port.
- Frontend: Expo/React Native + Tamagui, TanStack Query, secure-storage session store.
- Deploy: docker-compose + nginx (frontend), CI in `.github/cicd.yml`.

## Known weak spots (flagged by owner, not derivable from code)
- Picture pipeline (vips) is fragile — CI must build/test with `-tags vips`, easy to break with a plain `go test ./...` step.
- Auth/session flow is incomplete — Google login exists, rest of the session/auth lifecycle is WIP.

## Decisions worth remembering
- Solo project: docs are for the owner's own future reference, not team onboarding.

## Repo stats (indexed 2026-07-28)
1601 nodes, 6484 edges. Languages: Go (104 files), TypeScript (21), YAML (7), SQL (7), JS (4).
