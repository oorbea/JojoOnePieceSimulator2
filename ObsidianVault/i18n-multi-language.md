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

Implemented 2026-08-06, follow-ups closed 2026-08-06. Backend + frontend
infra, admin multi-locale form, full UI copy migration, and stable backend
error codes are all done and verified (unit + integration tests, real
Postgres, full frontend suite). Only remaining gap: validation `details`
(field-level messages on a 400) stay English-only — see Deliberately not
done below.

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
  - **Incident**: the original init call passed the flat per-locale
    catalogs straight into i18next's `resources` — but i18next requires
    `{ [lng]: { [namespace]: {...} } }`, not a flat object. Every `t()`
    call silently returned its own key back (e.g. `"stands.stats.attackRange"`
    on screen) instead of throwing, because `returnNull: false` degrades
    missing lookups to the key rather than failing loudly. This shipped
    undetected for the length of the original commit because no test ever
    rendered a component through `useTranslation()` with a real
    `I18nextProvider` until this follow-up added the first ones. Fixed by
    wrapping each locale's catalog in a `translation` key. Lesson: a UI
    string library needs at least one test that asserts *translated*
    output, not just that a key exists in the catalog — `i18n-keys.test.ts`
    checks catalog parity, but nothing checked resolution.
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
    language's cached list/detail. The admin-only `GET .../translations`
    query hangs off `allLocales` directly (`standKeys.translations(id)`,
    `devilFruitKeys.translations(id)`), **not** `all()` — it carries every
    locale at once, so it must not be branched by the active UI locale,
    and the mutations that already invalidate `allLocales` drop it for
    free.
- **Preference storage**: `users.language` column (Postgres enum
  `locale`), synced with the frontend's `language.store.ts` — the backend
  value wins once a session exists (`app/_layout.tsx` overrides the
  device-detected/stored value with `session.user.language`). Device
  detection only matters pre-login.
- **Admin multi-locale form: tabs, not accordion or locale-follows-UI.**
  `LocaleTabs` (`shared/components/presentational/locale-tabs.tsx`) is a
  row of `en-GB`/`es-ES`/`ca-ES` pills built on `ChannelBarItem`; only the
  active locale's `Description`/`Skills` render. Considered and rejected:
  an accordion (all three visible at once, but the modal grows too tall on
  mobile) and editing only the UI's current language (simplest, but makes
  it impossible to enter the mandatory `en-GB` content while using the app
  in `es-ES`). A locale left entirely blank is dropped from the submitted
  `translations` map (`toTranslationsPayload`); a half-filled locale
  (description without skills or vice versa) fails client-side validation
  before any request goes out — same "all or nothing per locale" rule the
  backend's `dto/translation_request.go` enforces, replicated in a
  `superRefine` on `power-translations.ts`'s zod schema. Opening "edit"
  awaits `GET .../translations` before showing the modal, rather than
  showing public-locale fields first and translations a beat later — avoids
  every non-`en-GB` tab flashing empty then jumping to real content.
- **Zod messages are i18n keys, not display strings.** Schemas
  (`stands.types.ts`, `devil-fruits.types.ts`, `profile.types.ts`,
  `power-translations.ts`) are module-level constants evaluated once at
  import time, so they can't call `t()` reactively. Each `.min()`/`.max()`
  message is instead the i18n key itself (e.g. `'validation.nameRequired'`),
  translated at the render site: `error={errors.name?.message && t(errors.name.message)}`.
  Only applies to fields with a custom validator — plain enum fields
  (`rarity`, `fruitType`, stand stats) have no custom message and are left
  as-is, since their default zod error is effectively unreachable (the
  form always starts from a valid enum default).
- **Stand stat abbreviations (PWR/SPD/RNG/END/PRE/DEV) stay in English** on
  the Stand card, deliberately — three-letter game-UI stat codes read the
  same convention in every locale (compare "HP" in any localized game),
  and translating them would make the card busier without adding clarity.
  The full stat labels in the edit form (`stands.stats.*`) are translated.
- **Backend error codes.** `dto.ErrorResponse` gained `code` (omitempty),
  a stable SCREAMING_SNAKE identifier (`STAND_NOT_FOUND`,
  `VALIDATION_FAILED`, `RATE_LIMITED`, ...) computed by
  `endpoints/error_codes.go`'s `errorCode()`, one switch mirroring
  `handleError`'s in `errors.go`. `error` (free text) stays as the
  fallback the frontend shows if a code isn't in its `errors.*` catalog
  yet, and what logs/curl still read directly.
  `AppError.code` (shared/api/errors.ts) carries it into the frontend;
  `showErrorToast` (shared/lib/toast.ts) resolves
  `t('errors.<code>', { defaultValue: error.message })` — i18next's
  `defaultValue` is what makes the degrade-to-English-text path work
  without extra branching. This function runs from
  `MutationCache.onError` (query-provider.tsx), configured once outside
  the React tree — no component to call `useTranslation()` from — so it
  reads the `i18next` singleton (`shared/i18n`'s default export) directly
  instead of the hook, same as every other non-component call site would
  need to.
  - **Not done**: validation `details` (the per-field messages inside a
    400's `ValidationError`) stay English-only. Encoding them would need a
    structured `{field, code}` shape instead of free-text strings — real
    scope, and lower priority since `zodResolver` already validates
    client-side, in the user's own language, before any request carrying
    those details ever leaves the client.

## Deliberately not done (follow-up)

- Validation `details` (see above) — needs a structured per-field DTO on
  the backend, not just a code on the top-level error.

Related: [[ADR]], [[frontend-stack]], [[backend-contract]]
