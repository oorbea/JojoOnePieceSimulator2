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
- Domain: Stands, DevilFruits, Powers, Users/Auth (Google OAuth), Games (Gauntlet/Versus modes, full lobby lifecycle incl. host config edit/team-switch/kick/transfer-host/lock/public-browser — see [[gameplay-domain-design]], [[game-lobby-management]]).
- Picture pipeline: libvips-based image processing (thumbnails/renditions), async status (`PictureStatus`), stored via a picture-storage port that fronts a multi-provider fallback chain (R2 → B2 → Supabase) — see [[storage-fallback-chain]].
- Frontend: Expo/React Native + Tamagui, TanStack Query, secure-storage session store.
- Deploy: docker-compose (base + dev/prod overrides) + nginx (frontend), CI in `.github/workflows/ci.yml` (PR→main, required check), CD in `.github/workflows/cd.yml` (push→main, Tailscale+SSH) — see [[cicd-deployment]].

## Known weak spots (flagged by owner, not derivable from code)
- Picture pipeline (vips) is fragile — CI must build/test with `-tags vips`, easy to break with a plain `go test ./...` step. **Fixed 2026-08-02**: `.github/workflows/ci.yml`'s backend job now runs both `go test ./...` and `go test -tags vips ./...` (libvips-dev installed on the runner first). Neither libvips nor its Go binding are touched by the storage fallback chain added 2026-08-09 — see [[storage-fallback-chain]].
- ~~Auth/session flow is incomplete — Google login exists, rest of the session/auth lifecycle is WIP.~~ **No longer true, corrected 2026-09-02** — see [[auth-hardening-2026-09-02]]. The flow is complete and verified in prod: Google ID token verification (`aud` pinned to `GOOGLE_CLIENT_ID`), HS256 JWT with alg/issuer/exp validation and a ≥32-char secret enforced at boot, per-IP login rate limit, `RequireAuth`/`RequireAdmin`, admin promotion by `ADMIN_EMAILS`. Three risks are *accepted*, not missing work: JWT in the query string for SSE/WS ([[picture-events-sse]], [[game-realtime-transport]]), session in `localStorage` on web ([[frontend-stack]]), and no refresh tokens ([[backend-contract]] — 24h JWT, 401 means re-login).

## Decisions worth remembering
- Solo project: docs are for the owner's own future reference, not team onboarding.
- Multi-language support (en-GB/es-ES/ca-ES) added 2026-08-06, backend + frontend infra done and verified; UI copy migration partial. See [[i18n-multi-language]].
- Storage fallback chain (R2 → B2 → Supabase) added 2026-08-09 to stretch free-tier object storage past R2's 10 GB — see [[storage-fallback-chain]].
- Game domain layer (Gauntlet + Versus modes) added 2026-08-10, domain-only pass (State/Strategy/Template Method, no infra/routes/migrations yet). Dead `game`/`user.Player` skeletons replaced; `user.Player` deleted outright. See [[gameplay-game-modes]] (rules) and [[gameplay-domain-design]] (technical design).
- Game application layer added the same day (2026-08-10), right after the domain pass: `GameService` (create/join/start/vote/tiebreak/disconnect/finalize), an in-memory `IGameStore` + reaper, a per-game `GameEventHub`, a Clock-driven voting timer, and cheap adapters for `ITiebreaker`/`IAssignmentWeights`/`IGamePowerPool`/`IStageCatalog` (the last one a hardcoded stub — no schema yet). See [[gameplay-application-layer]].
- **2026-08-11**: the game feature's infrastructure closed out — Redis-backed `IGameStore` (fail-closed, unlike the read-through cache's fail-open contract), a Postgres-backed stage catalog with admin CRUD, a persistent `IGameHistory`, and a native-WebSocket (`coder/websocket`, not socket.io) transport + HTTP routes for the whole `GameService` surface. See [[game-lobby-persistence]] and [[game-realtime-transport]]. Decisions worth remembering: single backend instance assumed throughout (no distributed lock, no cross-instance pub/sub for `GameEventHub` — scaling horizontally later needs a Redis pub/sub sibling for the hub, not just the store); votes stay hidden until a round resolves (loadouts don't — see [[game-realtime-transport]]); `/games/{id}/ws` is mounted outside the router's `Timeout(60s)` group, same precedent as `/events`.
- **2026-09-02**: `apps/frontend/src/shared/contracts/` (enums, DTOs, WS commands/frames, error
  codes) is now generated from Go by `apps/backend/cmd/typegen`, ending independent hand-mirroring
  on the frontend — the drift had already cost a real incident ([[user-profile-feature]]'s
  frontend/backend contract coupling) and had live drift at the time (`backend-contract.md`
  documented rarity without the `MYTHICAL` tier; a stale `EMPTY_MANGAS` i18n key and a missing
  `GAME_NOT_OVER` one, both caught by the new guard test). Chose a reflection-based Go generator
  over swaggo/OpenAPI (swaggo can't express the WS surface or error codes) and over spec-first
  (would invert Go from decider to consumer). CI gate (`contracts` job) + a guard test
  (`contracts.test.ts`) enforce it can't silently drift again. See [[contratos-tipos-generados]].
- **2026-08-28**: round-resolved feedback (per-option vote tally + winner, shown inline where the vote bar was) shipped, closing the bullet open since [[game-vote-buttons-2026-08-26]]. Required splitting `Game.resolveRound`/new `Game.CompleteRound()` so `RESOLVING` is a real, observable pause (`GameService.scheduleResultDelay`, `game.ResultDuration` = 6s fixed) instead of a same-call pass-through — the same shape the sorteo's `scheduleRevealDelay` already has. Also added `Round.TiedVotes`, the one deliberate exception to "votes hidden while live": a tie's vote breakdown is now preserved (was previously wiped by `Ballot.Reset()` with nothing kept) and revealed before the revote replaces it. See [[game-round-result-2026-08-28]].

## Repo stats (indexed 2026-07-28)
1601 nodes, 6484 edges. Languages: Go (104 files), TypeScript (21), YAML (7), SQL (7), JS (4).
