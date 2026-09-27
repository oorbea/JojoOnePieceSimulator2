---
title: "Stands and Devil Fruits are hand-authored by admins via the admin panel"
tags:
  - project
  - jojo-onepiece-simulator
  - content
  - admin
---

# Stands and Devil Fruits: manual admin authoring, no bulk import

There's no seed script, scraper, or bulk-import pipeline for game content. Admins create and edit
every Stand and Devil Fruit one at a time through the admin panel's CRUD screens (see
[[admin-panel-crud-ux-fixes]], [[admin-search-and-filters]], [[admin-crud-cache-stale-sw]] for the
UX/bug history of those screens).

**Partial exception, 2026-09-27:** [[catalog-seed-from-prod]] added `cmd/catalogsync`, a
repeatable tool that reads prod's existing catalogue and generates a goose migration to fill in
missing/placeholder/stale *translations* (and, local-only, mirror the whole catalogue for realistic
test data). It never creates or edits the actual game content itself — an admin still types every
Stand/Devil Fruit in by hand through the admin panel, same as always - it only catches up
translations for content that already exists. The "no batch-content risk" reasoning below still
holds for the admin-facing CRUD surface.

**Why this matters for future work:** content volume grows only as fast as an admin types it in —
there's no batch-content risk to design around (e.g. no need for import validation, dedup-on-import,
or large-payload handling). Any feature touching Stand/DevilFruit data should assume single-record
mutations from the admin UI as the only write path, not bulk writes.
