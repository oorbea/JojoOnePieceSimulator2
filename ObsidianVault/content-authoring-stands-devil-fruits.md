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

**Why this matters for future work:** content volume grows only as fast as an admin types it in —
there's no batch-content risk to design around (e.g. no need for import validation, dedup-on-import,
or large-payload handling). Any feature touching Stand/DevilFruit data should assume single-record
mutations from the admin UI as the only write path, not bulk writes.
