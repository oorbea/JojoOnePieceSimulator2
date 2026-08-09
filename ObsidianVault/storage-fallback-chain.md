---
title: Storage fallback chain (R2 → B2 → Supabase)
tags:
  - project
  - decision
  - jojo-onepiece-simulator
  - storage
---

# Storage fallback chain (R2 → B2 → Supabase)

Added 2026-08-09. R2's free tier only gives 10 GB stored; to stretch further
without paying, pictures now fall through an ordered chain of S3-compatible
providers, each with its own free tier: **Cloudflare R2 (10 GB) → Backblaze
B2 (10 GB) → Supabase Storage (1 GB)**. All three speak the S3 API, so one
adapter (`internal/infrastructure/storage/s3store`) covers all of them —
only endpoint/region/credentials differ per provider.

## How it works

- `internal/domain/ports.IPictureStorage.Upload` now returns
  `StoredPicture{Provider, Key}` instead of just an error — callers learn
  where an upload actually landed. `ports.Picture` gained
  `PreferProvider`, so `picture_worker.go` pins a thumbnail to whatever
  provider its main rendition landed on (a Stand/DevilFruit/avatar's two
  renditions never end up split across two buckets).
- `internal/infrastructure/storage/fallback.PictureStorage` is the chain
  itself: for each upload it tries tiers in configured order, skipping a
  tier once it's within `STORAGE_QUOTA_THRESHOLD_PCT` (95% by default) of
  its quota, and falling through to the next tier on a runtime `Put` error
  too (not just on quota). It never migrates an object once written — a
  picture stays on whichever provider it first landed on.
- `internal/domain/ports.IStorageLedger` (Postgres-backed:
  `internal/infrastructure/repositories/storage_ledger.go`, table
  `storage_objects`, migration `00007_storage_objects.sql`) is the single
  source of truth for which provider a given key lives on. `PresignGetURL`/
  `Delete` look the key up there; a key predating the ledger (or the whole
  feature) defaults to the first configured tier (R2).
- `powers.picture`/`picture_thumb` and `users.avatar_key`/`avatar_thumb_key`
  are **unchanged** — still bare object keys, no provider suffix/column.
  Keeping the provider mapping in its own table (rather than a column on
  every picture-bearing entity) meant zero migration on those tables and no
  sqlc/DTO changes outside the new table.
- `internal/application/services.StorageReconciler` periodically (default
  every 6h, `STORAGE_RECONCILE_INTERVAL=0` disables it) walks every
  configured bucket and replaces the ledger's inventory for that provider
  with what's actually there, correcting drift from the ledger writes in
  `fallback.PictureStorage` being best-effort (an object is never "lost" if
  its `Record`/`Forget` call fails after the S3 operation already
  succeeded — it just goes stale in the ledger until the next
  reconciliation pass).

## Config

`STORAGE_PROVIDERS` (default `r2`, R2 must be first) picks which tiers are
active; each additional provider (`b2`, `supabase`) only needs its
credentials/endpoint/bucket/quota env vars once it's listed there. See
`deployments/.env.example`'s "Object Storage" sections for the exact var
names — B2 and Supabase's S3-compatible endpoints are taken as-is from each
provider's own dashboard rather than reconstructed from an account id, to
avoid guessing their URL shape wrong (unlike R2, whose endpoint is
`https://<account>.r2.cloudflarestorage.com` and is built from
`R2_ACCOUNT_ID` via `s3store.R2Endpoint`).

## Test coverage (2026-08-09)

Exhaustive pass across every layer, happy/bad/edge paths:

- `config`: every `STORAGE_*`/`B2_*`/`SUPABASE_*` Load validation rule
  (unknown/duplicate/empty provider, R2-must-be-first, threshold 1-100
  bounds, negative reconcile interval, each B2/Supabase var individually
  required only when listed, quota defaults).
- `fallback`: construction errors (empty tiers, bad threshold, ledger.Usage
  failure), exact quota-threshold boundary (fits at N%, rejected at N%+1
  byte), zero-quota-is-unlimited, ledger.Record failure being genuinely
  best-effort (upload still succeeds, usage counter deliberately NOT bumped
  until the next reconciliation), unknown `PreferProvider` falling back to
  chain order, non-seekable content shorter than declared `Size` erroring,
  3-tier exhaustion joining all three Put errors, every `Delete`/
  `PresignGetURL` error path (backend error, ledger.Get error, unknown
  recorded provider, Forget failure swallowed but usage not decremented),
  and a 50-goroutine concurrent-Upload test proving the atomic usage counter
  never loses an update.
- `s3store`: Put/Del/Walk error-wrapping and Walk pagination/early-stop
  against a local `httptest` server (no real bucket needed), PresignGet's
  URL/TTL correctness (pure local computation, no network).
- `storage_reconciler`: Walk/Replace/Usage-refresh error propagation,
  multiple backends reconciled independently, `Start`'s ticker loop actually
  firing and stopping on context cancellation.
- `storage_ledger` (new `//go:build integration` test, same pattern as
  `user_repository_test.go`): Record/Get/Forget/Usage/Replace round-tripped
  against a **real** throwaway Postgres container with the actual
  migrations applied - upsert-on-conflict, not-found returns `ok=false` with
  no error, `Replace` only touching its own provider's rows, empty-`Replace`
  clearing a provider entirely. All green.

**Gotcha, fixed 2026-08-09**: this dev machine's Windows Application Control
blocks freshly built `go test` binaries essentially at random for anything
importing `net/http/httptest` (opens a local listener) - confirmed by
bisecting: the identical file compiles clean (`go vet` passes) but
`go test` sometimes fails at process-start with "Una directiva de Control
de aplicaciones bloqueó este archivo" before any test code runs. Not a code
problem, and deliberately **not** fixed by loosening the OS policy - no
idea whether it's Smart App Control, WDAC, or a third-party EDR on this
machine, and blanket-disabling application control to unblock Go binaries
would weaken the whole system's security for a Go-test-specific problem.
Fixed properly instead by adding `deployments/docker-compose.test.yml` (a
`backend-test` service on `golang:1.26-alpine3.22`, the same image
`Dockerfile.backend`'s builder stage uses) plus three Makefile targets:
`test-docker` (`go test ./...`), `test-vips-docker` (installs
`build-base pkgconfig vips-dev` and runs `-tags vips`), and
`test-integration-docker` (joins `public-net`, needs `db-up` first, runs
`-tags integration` against the real compose Postgres/Redis by service
name). Also sidesteps the app-control problem entirely (Linux container) and
is a closer match to CI's `ubuntu-latest` than the bare host ever was. All
three verified green against real Docker: unit, vips-tagged, and
integration (after seeding the schema by briefly running the real `backend`
service so its startup `goose up` applies).

## Deliberately not done

- `ErrStorageExhausted` (all tiers full/erroring) is **not** mapped to an
  HTTP status code. `IPictureStorage.Upload` is only ever called from
  `picture_worker.go`'s async job processing — the HTTP handler already
  returned `202 Accepted` before the worker runs — so a storage-exhausted
  upload just flows through the exact same path any other transcode/upload
  failure already does: logged, `PictureStatus` set to `FAILED`. No new
  error code, no i18n string needed.
- No B2/Supabase account has actually been provisioned yet — see the setup
  guide (delivered separately, not in the vault) for how to fill in
  `B2_*`/`SUPABASE_*` once ready to activate those tiers.

Related: [[ADR]], [[docker-setup]], [[cicd-deployment]].
