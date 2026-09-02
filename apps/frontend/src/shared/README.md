# Shared

Cross-feature primitives, mirroring each feature's container/presentational
split (see `src/features/README.md`). Anything two or more features need
goes here — never re-implemented per feature.

## `api/`

- **`client.ts`** — single axios instance (`apiClient`), 15s timeout, treats
  HTTP 304 as a success response.
- **`interceptors.ts`** (`registerInterceptors`) — attaches
  `Authorization: Bearer`, `Accept-Language` (from `language.store`), and
  conditional-GET `If-None-Match`; on a 304 replays the cached body (via
  `etag.ts`'s `etagCacheKey`/`getCachedResponse`/`rememberResponse`); on a
  401 clears the session (no refresh token, user re-logs in).
- **`errors.ts`** — normalizes axios errors into `AppError`.
- **`query-keys.ts`** — root query key `queryKeys.root = ['jops']`, extended
  per feature.
- **`assert-contract.ts`** (`assertContract`) — dev-only, non-throwing Zod
  check that a REST response still matches the generated contract; call it
  wrapped in `if (__DEV__)` so it's stripped from the prod bundle.

## `contracts/`

**Generated — do not edit.** Produced from Go by `apps/backend/cmd/typegen`
(`make types`). `index.ts` is a type-only barrel re-exporting `enums.ts`,
`errors.ts`, `dto.ts`, `ws.ts`; import *schemas* (values, not just types)
directly from those files rather than the barrel, so unused schemas don't
end up in the bundle.

## `stores/`

Zustand. `session.store.ts` (JWT + user, persisted via `secure-storage`;
exposes a non-hook `getSessionToken()` for interceptors), `theme.store.ts`,
`language.store.ts`.

## `config/env.ts`

Zod-validated `EXPO_PUBLIC_*` env vars, parsed once at boot — throws loudly
on a missing/malformed var instead of a mystery `undefined` later.

## `i18n/`

i18next setup for `en-GB`/`es-ES`/`ca-ES`: `SUPPORTED_LOCALES`,
`detectDeviceLocale()`, endonym labels. Initialized synchronously with the
default locale at import time.

## `lib/`

Platform helpers: `a11y.ts` (`a11yProps()` — routes RN accessibility props
to ARIA on web instead of leaking them as raw DOM attributes),
`secure-storage.ts` / `async-storage.ts`, `layout.ts`,
`overlay-position.ts`, `tamagui-token.ts`, `power-translations.ts` /
`stage-translations.ts` (hand-written schemas composing generated
contracts — never a hand-mirror of a backend enum), `toast.ts`,
`scroll-bus.ts`, `nav-insets.tsx`, `web-blur.ts`.

## `hooks/`

`use-app-state`, `use-debounced-value`, `use-now`, `use-picture-picker`,
`use-reduced-motion`, `use-roving-group`.

## `components/`

- **`containers/`** — `app-shell-container.tsx` (wires nav/session/theme
  into `AppShell`), `error-boundary.tsx`.
- **`presentational/`** — the shared design system (`GlossButton`,
  `GlassPanel`, `WiiCard`, `TooltipBubble`, `AppShell`, `PageShell`, ...),
  barreled via `index.ts`, each with a colocated `__tests__` file.

## Other

`types/css.d.ts`, `assets.ts`.
