---
title: "picturesync: filling missing Devil Fruit pictures from the wiki (2026-10-08)"
tags:
  - project
  - jojo-onepiece-simulator
  - backend
  - content
  - tooling
---

# picturesync — Devil Fruit pictures from the wiki, with human review

## Status

Shipped and run against prod 2026-10-08. Before: 78 of 152 Devil Fruits had no picture
(`pictureStatus: NONE`); all 169 Stands already had one. After: 63 uploaded, all `READY`
(worker transcoded them), 15 left empty on purpose — 10 "chulos" (EPIC/LEGENDARY/MYTHICAL, for admins to
upload by hand) + 5 the owner saw no good candidate for (Giro Giro, Iku Iku, Muchi Muchi,
Shari Shari, Inu Inu Model Hound). Code: `apps/backend/cmd/picturesync/` (stdlib only).

## Flow

`list` → `candidates` → human picks in `review.html` → `upload`. Talks to the backend over its
**public API only** (no DB tunnel), so it works against prod with an ADMIN access token.

1. `list`: `GET /devil-fruits`, keeps rows with `picture == ""` minus `--exclude-rarity`
   (default `EPIC,LEGENDARY,MYTHICAL`) → `out/missing.json`. Read-only.
2. `candidates`: MediaWiki API on `onepiece.fandom.com` (page by exact title, search as fallback;
   infobox image first, then files named like the fruit) → up to 3 images per fruit downloaded
   under `--out` + a self-contained `review.html` (radios default to "ninguna", so nothing is
   uploaded without an explicit pick; button downloads `selections.json`).
3. `upload`: reads `selections.json`, re-lists live state and **skips fruits that already have a
   picture** (safe to re-run after a 401 or partial failure), then `PATCH /devil-fruits/{id}/picture`
   one by one with a pause (writes are rate-limited; retries 429). `--dry-run` first.

Images live only in the gitignored `picturesync-out/` and in R2 after upload; never committed.

## Gotchas

- **The wiki's image CDN (`static.wikia.nocookie.net`) 403s hotlinks without a wiki `Referer`.**
  The API itself is fine with any User-Agent. Sending `Referer: https://onepiece.fandom.com/`
  fixes it. (Same as viewing the page in a browser.)
- Fandom's `"missing":""` in `action=query&titles=` is an empty **string**, not an object —
  decode as `json.RawMessage`.
- Fruit titles are `Inu Inu no Mi, Model: Hound` (comma + colon), not the app's
  `Inu Inu no mi: Model Hound`; `wikiTitle` converts, search is the fallback.
- Selection export is a browser download: it lands in `Downloads`, not the project dir. Copy it into
  `--out`.
- Max upload is `PICTURE_MAX_BYTES` (5 MiB default); the tool skips candidates over ~4.5 MiB.

## Getting a prod ADMIN token without the Google login

Prod auth is Google-only; the web access token is memory-only (see
[[session-token-storage-2026-09-05]]). With the browser already logged in as admin, calling
`POST /api/v1/auth/refresh` from the page (`credentials: 'include'`, header `X-JOPS-Refresh: 1`)
returns `accessToken`; the browser keeps the rotated refresh cookie, so the session survives.
Done through Claude in Chrome's `javascript_tool`, written to the session scratchpad (outside the
repo), deleted afterwards. **Prod's `JWT_TTL` is 8h** (`deployments/.env`), not the 15m of
`.env.example`. The tool takes it from `PICTURE_SYNC_TOKEN` or `--token-file`, never a flag value.
Network-request inspection in Claude in Chrome does not expose request headers, so the refresh
call is the route.

Related: [[content-authoring-stands-devil-fruits]], [[catalog-seed-from-prod]],
[[media-proxy-content-addressed]], [[storage-fallback-chain]], [[admin-search-and-filters]]

## Follow-up: "Sin foto" admin filter (commits `6cb779f`, `5696fcc`)

So admins can find the fruits/stands still lacking a picture (the 15 above).
`?hasPicture=false|true` on `GET /devil-fruits` and `GET /stands`; chip "Sin foto" next to the rarity
controls in both admin screens (shared `FilterChip`, with tooltip).

- **Semantics: `powers.picture = ''`, not `picture_status`.** A re-upload sets PENDING/FAILED while
  the old rendition is still served, so status alone would call those rows picture-less. Side effect:
  a first upload still PENDING (or FAILED with no earlier picture) counts as "without picture" —
  there is nothing to show yet, so that is right.
- Invalid value → 400 `ValidationError` field `hasPicture`.
- Applied in Filter, Page **and Count** SQL for both entities. Stand's lives inside the recursive
  `base` CTE (SELECT list untouched); a picture-less child whose parent has a picture returns with
  its parent hydrated, parent not returned as a match.
- Cache keys and cursor fingerprints come from `ports.Canonical()` (new `optBool`), not a hand-edited
  `keys.go` — a missed field is caught by `canonical_test`'s field-count test. Update the older trap
  note in [[admin-search-and-filters]] accordingly.
- Trap: `FilterStandRows`' WHERE ends in `)),` (closes the CTE too) — scripted edits broke the parens.
- sqlc regenerated with the `sqlc/sqlc` image (`MSYS_NO_PATHCONV=1`, mount `apps/backend`); `make` is
  not installed in this Git Bash, run the compose commands directly.
- `swag@latest` regen produces ~2000 lines of unrelated drift (checked-in docs are stale) — not
  committed; annotations are in source. A separate swagger regen commit is pending.
- Frontend verify recipe: `cp -r /repo/apps/frontend/.` copied the Windows `node_modules` and hung for
  hours; replaced with `tar --exclude=node_modules` and jest now runs with `--forceExit` + `timeout`
  (updated in `.claude/skills/verify/SKILL.md`; see [[norma-verificacion-docker]]).
