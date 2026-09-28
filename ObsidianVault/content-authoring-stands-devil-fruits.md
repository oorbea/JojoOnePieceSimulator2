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

**Second partial exception, 2026-09-28:** [[catalog-seed-part4-stands]] added 29 brand-new Stand
rows (Diamond is Unbreakable) via a hand-authored migration, not the admin UI — the inverse case
from catalogsync (new content, not translation catch-up on existing content). Still
Claude-authored-then-human-reviewed before merging, same trust model as catalogsync's own
translation step.

**Why this matters for future work:** content volume grows only as fast as an admin types it in, or
as fast as a reviewed content-migration adds it — there's no *unreviewed* batch-content risk to
design around (e.g. no need for import validation, dedup-on-import, or large-payload handling from
an untrusted source). Any feature touching Stand/DevilFruit data should still assume single-record
mutations from the admin UI as the primary write path in steady state, not bulk writes.
