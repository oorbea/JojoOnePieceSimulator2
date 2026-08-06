---
title: "i18n: English (UK), Castilian, Catalan"
tags:
  - project
  - jojo-onepiece-simulator
  - decision
  - i18n
---

# i18n — multi-language support

## Status

Implemented 2026-08-06. Backend + frontend infra done and verified (unit +
integration tests, real Postgres). UI copy migration is partial — the
pattern is proven on a handful of files (login, error fallback, theme
toggle), the rest of `src/` still has hardcoded English. Admin multi-locale
edit form for Stand/DevilFruit content not built yet.

## Scope

Three locales: `en-GB`, `es-ES`, `ca-ES`. Two independent things get
translated:
1. UI chrome (frontend-only, i18next).
2. Power content (`description`, `skills` on Stands/DevilFruits) — backend
   data, resolved server-side per request.

## Decisions

- **Names are never translated.** "Star Platinum", "Gomu Gomu no Mi" read
  the same in every locale — they're proper nouns. `powers.name` and its
  `UNIQUE` constraint stay untouched; only `description`/`skills` moved to
  a new `power_translations` table. This is what made the whole migration
  tractable — no need to redesign name-based lookups
  (`GetStandIDByName`, `ORDER BY p.name`).
- **UI**: `i18next` + `react-i18next`, installed via `pnpm add` (never
  hand-edited into package.json). Catalogs at
  `apps/frontend/src/shared/i18n/locales/*.json`. i18next initializes
  synchronously at module-import time with `en-GB` so `useTranslation()`
  never renders before the instance is ready — `language.store.ts`'s async
  `hydrate()` (AsyncStorage + `expo-localization`) then calls
  `i18n.changeLanguage()` once the real preference is known.
- **Content fallback chain**: `ca-ES → es-ES → en-GB`. `en-GB` is
  mandatory — every Power must have an `en-GB` translation, enforced in
  the application layer (DTO validation), not a DB constraint. Resolved in
  a single SQL query via a `LEFT JOIN LATERAL` against
  `power_translations`, ordered by `array_position()` over the caller's
  fallback-chain array — no N+1, no Go-side COALESCE juggling.
- **API contract**: hybrid. Public reads (`GET /stands`, `GET
  /devil-fruits`) resolve one locale via `Accept-Language` (or `?lang=`
  override) — response shape unchanged (`description: string`, `skills:
  string[]`), so existing Zod schemas keep working. Admin-only `GET
  .../translations` returns every locale at once (`map[locale]content`),
  for edit forms. Writes (`POST`/`PUT`) always take a `translations` map
  keyed by locale, `en-GB` mandatory.
- **Caching — the actual risk in this whole feature.** Two independent
  cache layers had zero locale dimension before this:
  - HTTP ETag/Cache-Control middleware varied on `Authorization` only →
    added `Accept-Language` to `Vary`. Lower-risk than it sounds (the ETag
    hash already differs per locale, so a stale 304 was never actually
    reachable) but correct per HTTP semantics regardless.
  - The Redis read-through decorator (`internal/infrastructure/cache/`)
    keyed reads by id/name/"all" with **no locale at all** — this one was
    the real bug-in-waiting: an `es-ES` read would have poisoned the cache
    for every other locale's request to the same stand. Fixed by putting
    locale into every cache key (`id:es-ES:<uuid>`, `all:ca-ES`, ...).
    Writes still invalidate the whole namespace (unchanged), which
    correctly drops every locale's entries together since a write's
    translations touch all locales at once. Verified with a dedicated test
    per repository (`TestStandRepository_FindByID_NeverCrossesLocales`,
    mirrored for DevilFruit) asserting three locales produce three
    independent cache entries, and an integration test
    (`TestStandRepository_Locale_ResolvesAndFallsBack`) against real
    Postgres proving the SQL fallback chain itself.
  - Frontend: TanStack Query keys (`standKeys`, `devilFruitKeys`) branch by
    the active locale (`allLocales` unlocalized prefix for
    mutation-invalidation, `all()` locale-branched for reads) — otherwise
    switching language in the UI would keep serving the previous
    language's cached list/detail.
- **Preference storage**: `users.language` column (Postgres enum
  `locale`), synced with the frontend's `language.store.ts` — the backend
  value wins once a session exists (`app/_layout.tsx` overrides the
  device-detected/stored value with `session.user.language`). Device
  detection only matters pre-login.

## Deliberately not done (follow-up)

- Backend error messages (`shared/api/errors.ts`) are still English-only —
  translating them needs the backend to emit stable error codes instead of
  free-text messages, a separate, bigger change.
- Full UI copy migration: ~20 files still have hardcoded English (profile
  screen, stand/devil-fruit forms and screens, admin hub, toast copy in the
  mutation hooks, enum value labels for rarity/stand-stat/fruit-type/role).
  The pattern to follow is in `login-screen.tsx` / `error-fallback.tsx` /
  `theme-toggle.tsx`.
- Admin multi-locale edit form (Stand/DevilFruit create/update UI) still
  posts a single-locale body shape conceptually — needs a tabbed/
  accordion per-locale input before it can actually exercise the new
  `translations` map the backend now expects.

Related: [[ADR]], [[frontend-stack]], [[backend-contract]]
