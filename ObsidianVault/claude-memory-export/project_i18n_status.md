---
name: project_i18n_status
description: "i18n (en-GB/es-ES/ca-ES) fully closed 2026-08-06: infra, admin translations UI, full copy migration, error codes"
metadata: 
  node_type: memory
  type: project
  originSessionId: 58de9bcf-8ec3-4e33-8782-7fc908f74d28
  modified: 2026-08-06T21:28:40.071Z
---

Multi-language support (en-GB/es-ES/ca-ES) landed on `develop` 2026-08-06 across two sessions. First session shipped backend+frontend infra; this session closed all three follow-ups it left open. All verified: backend `go test ./...` green, frontend typecheck/lint/test (184 tests) green.

**Done this session**:
1. Admin Stand/DevilFruit forms now have a `LocaleTabs` per-locale input (tabs, not accordion) for the backend's `translations` map — this also fixed a real regression: the forms were still posting flat `description`/`skills`, so every admin create/update had been returning `400 translations.en-GB is required` since the first i18n session.
2. Full UI copy migration — every remaining hardcoded-English file now uses `useTranslation()`/`t()`: profile, stands/devil-fruits screens+forms+cards, admin hub, home, confirm sheets, skills field, glass select, loading screen, app shell, not-found, all mutation-hook toasts, enum value labels (rarity/stand-stat/fruit-type/role).
3. Backend error codes: `dto.ErrorResponse.Code` + `endpoints/error_codes.go`, surfaced through `AppError.code` and `showErrorToast` (`t('errors.<code>', {defaultValue: error.message})`).

**Incidental bug fixed**: `shared/i18n/index.ts`'s `resources` object was missing i18next's required namespace wrapper (`{lng: {translation: {...}}}` vs the flat catalog it was passing) — every `t()` call had been silently returning the raw key back since the first i18n session. Nothing caught it because no test rendered a component through `useTranslation()` with a real `I18nextProvider` until this session added the first ones.

**Not done (explicit follow-up, low priority)**: validation `details` (per-field messages on a 400) stay English-only — needs a structured `{field, code}` DTO on the backend, not just the top-level `code` this session added. Lower priority since `zodResolver` already validates client-side, in-language, before those details would ever reach a user.

**Why this matters**: full decision record, rationale, and the same follow-up live in `ObsidianVault/i18n-multi-language.md` — read that first before touching this area again.

**How to apply**: this feature area is now closed short of the one noted follow-up. If asked to touch i18n again, read the vault entry first for the LocaleTabs/zod-as-i18n-key/error-code decisions before changing anything.
