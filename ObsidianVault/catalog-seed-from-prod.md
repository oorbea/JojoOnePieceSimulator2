---
title: "catalogsync: seeding the prod catalogue + translations into a migration"
tags:
  - project
  - jojo-onepiece-simulator
  - decision
  - i18n
  - content
---

# catalogsync — repeatable prod-catalog seed migration

## Status

Shipped 2026-09-27. `db/migrations/00017_seed_catalog_20260927.sql` generated and applied; verified
end-to-end against a fresh local stack (full catalogue + images + focal points render through the
API, post-dump admin-edit guard confirmed live). Same day, the mandatory/default content locale
also moved from `en-GB` to `es-ES` (backend + frontend) — see [[i18n-multi-language]] for that half.

## Problem

Admins fill in Stands/Devil Fruits/Characters by hand in prod (see
[[content-authoring-stands-devil-fruits]]) in whatever locale they're comfortable in. In practice:
Stands were only reliably correct in `es-ES`; Devil Fruits only in `en-GB`; the 4 existing
Characters happened to already have all three locales filled well. Local dev also had none of
prod's real catalogue, images, or focal points to test against.

## Decision

`apps/backend/cmd/catalogsync` (`dump` / `plan` / `generate` subcommands — see its own doc comment)
reads prod read-only over the DB tunnel ([[cicd-deployment]]'s tunnel, same credentials DBeaver
uses), diffs the snapshot against a versioned `db/seed/catalog/translations.json` overlay, and
generates a goose migration once every translation is filled in. Key choices:

- **Translation authoring stays human-reviewed, not an API call.** `plan` writes a to-do list
  (`pending.json`); translating it into `translations.json` is a Claude Code session (batched across
  6 parallel subagents for the first real run — 75 Devil Fruits × 2 locales), reviewed in the git
  diff before `generate` runs. No `ANTHROPIC_API_KEY`, no unreviewed LLM text landing in a migration.
- **Source locale per kind:** Stand → `es-ES`, Devil Fruit → `en-GB`, fixed. Characters vary row by
  row — `plan` detects a best guess (small stopword-based per-locale scorer) into
  `sources.suggested.json` and refuses to proceed until a human confirms it into `sources.json`.
- **Placeholder detection isn't just "text differs."** A locale's existing text is flagged
  `placeholder` (needs retranslating, not trusted) if it's empty, under ~30 chars (real content here
  is paragraph-length; literal junk like `"A"` or `"placeholder"` was found this way), identical to
  the source locale's text, or confidently detected (by the same scorer) as being written in the
  source language instead of its own.
- **Environment-split migration, one file.** Goose's `ENVSUB` substitutes `${SEED_CATALOG_FULL:-false}`
  into a `WHERE` guard on every full-catalogue `INSERT ... SELECT` — true only in
  `docker-compose.dev.yml`/`.test.yml`, never in prod. Prod only ever gets the translations-only
  block, which is itself guarded per-row: `WHERE EXISTS (... AND updated_at <= <dump timestamp>)` —
  an admin edit made after the dump always wins, and a row deleted after the dump never resurrects.
  Verified live: touching a row's `updated_at` post-dump then re-running that row's own upsert
  statement affects 0 rows.
  - **Gotcha:** goose's `ENVSUB` (mfridman/interpolate) treats a bare `$$` as an escaped literal `$`
    (envsubst convention) — it does NOT understand Postgres's `$tag$` dollar-quoting. A `DO $$ ...
    END $$;` PL/pgSQL block silently corrupts into `END $;` (syntax error) the moment `ENVSUB ON` is
    active anywhere in the file. Avoid PL/pgSQL entirely here — every full-catalogue statement is a
    plain `INSERT INTO t (...) SELECT ... WHERE '${VAR}' = 'true'` instead of `VALUES (...)`, which
    sidesteps the whole class of problem.
- **Same R2 bucket, local and prod — image keys only, never bytes.** The migration seeds
  `picture`/`picture_thumb`/`picture_card`/`picture_media_id` (and the `storage_objects`/
  `media_objects` rows they reference) with prod's real keys; local dev's `.env` already points at
  the same bucket, so images and focal points render immediately, no file copy. This makes a local
  `Delete` dangerous — see below.
- **`STORAGE_DISABLE_DELETE`** (new config flag, `fallback.PictureStorage.SetDeleteDisabled`):
  no-ops `Delete` when true. Set in `docker-compose.dev.yml` only, so a local delete never removes
  an object prod still serves.
- **Cache invalidation on migration-writes.** A seed migration writes `power_translations`/
  `character_translations` etc. with raw SQL, bypassing every repository's own `invalidate()` call.
  `postgres.Migrate` now reports whether it actually applied anything (comparing goose's DB version
  before/after); `cmd/app/main.go` calls the new `cache.InvalidateCatalogNamespaces` right after
  building the Redis client when it did. Fail-open, like every other cache path.

## Re-running it

Admins keep adding content in prod, so this isn't a one-shot: `make catalog-dump-docker` (needs
`SOURCE_DATABASE_URL` — see [[cicd-deployment]]'s DB-tunnel section) → `make catalog-plan-docker` →
translate `pending.json` into `translations.json` → `plan` again until "nothing pending" → `make
catalog-generate-docker` writes the next `NNNNN_seed_catalog_*.sql`. `db/seed/catalog/README.md` has
the full loop. `snapshot.json`/`pending.json`/`sources.suggested.json` are gitignored (regenerated,
not a source of truth); `translations.json`/`sources.json` are committed.

## Security note

`deployments/README.md`'s DB-tunnel section has the prod Postgres password in plain text — flagged,
not fixed here (out of scope for this change; recommend rotating it and moving it out of the
README in a separate change).

Related: [[i18n-multi-language]], [[content-authoring-stands-devil-fruits]], [[cicd-deployment]],
[[storage-fallback-chain]], [[media-proxy-content-addressed]]
