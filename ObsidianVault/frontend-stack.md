---
title: Frontend stack & decisions
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
---

# Frontend — apps/frontend

PWA, standalone pnpm project (own `pnpm-lock.yaml`, no root workspace — `pnpm-workspace.yaml` here is ONLY pnpm's build-script allowlist, not a monorepo root).

## Stack

TypeScript, Expo SDK 57, React Native 0.86, React Native Web, Expo Router (file routing), TanStack Query v5 + `@tanstack/query-async-storage-persister` (persisted via AsyncStorage), Zustand v5, React Hook Form + Zod, Axios, Tamagui v2.5.1 (UI kit — NativeWind dropped, one styling system only), browser-native `WebSocket` for the game's realtime transport (`src/features/game/stores/game-socket.store.ts`, see [[game-realtime-transport]]; `socket.io-client` was uninstalled 2026-09-02, see [[socket-io-cleanup-2026-09-02]]), Expo SecureStore (native) / localStorage (web), Expo PWA (manifest + custom SW), i18next + react-i18next + expo-localization (2026-08-06, see [[i18n-multi-language]]).

## i18n (2026-08-06)

`src/shared/i18n/` (init + `locales/{en-GB,es-ES,ca-ES}.json`), `src/shared/stores/language.store.ts` (device detect → AsyncStorage → overridden by `session.user.language` once logged in, mirrors `theme.store.ts`'s shape). `src/shared/api/interceptors.ts` sends `Accept-Language` on every request. TanStack Query keys for Stand/DevilFruit branch by locale (`standKeys.all()`/`devilFruitKeys.all()`) so switching language doesn't serve the previous language's cached list; mutations invalidate the unlocalized `allLocales` prefix to drop every locale's cache at once. `copy.test.ts`'s em/en-dash guard now also scans the locale JSON catalogs, and a new `i18n-keys.test.ts` asserts all three catalogs have identical key sets. **Not done yet**: most of `src/` still has hardcoded English copy (only login/error-fallback/theme-toggle migrated so far), and the Stand/DevilFruit admin forms don't yet have a per-locale input UI for the backend's `translations` map. See [[i18n-multi-language]] for the full decision record and the follow-up list.

All deps installed via CLI only, never hand-edited into `package.json` (user hard requirement). Package manager: pnpm.

## Decisions made with user

| Question | Decision |
|---|---|
| UI kit | Tamagui only (dropped Gluestack and NativeWind) |
| Repo layout | Standalone pnpm project, no root workspace |
| Docker | Both dev (hot reload) + prod (static export via nginx) |
| Platforms | Web-first, native-ready (iOS/Android configured, not fully asset-built) |
| Feature scaffolding | Skeleton only, no feature code this pass |
| socket.io | ~~Install only, no wiring — backend's game WS is native, not socket.io (see [[game-realtime-transport]])~~ **Reversed 2026-09-02**: it was never wired and never imported, so the dep was uninstalled. Realtime is the browser's own `WebSocket` — see [[socket-io-cleanup-2026-09-02]] |

## Architecture: Feature-Driven + Container/Presentational

- `app/` = routes only, thin shims rendering a container.
- `src/features/<feature>/` owns its api/hooks/stores/types/components. Cross-feature imports ONLY through the feature's `index.ts` barrel.
- Inside a feature: `components/containers` (data/nav/forms wiring) vs `components/presentational` (pure Tamagui UI, props in/JSX out).
- `src/shared/` mirrors the same split — axios client, ETag/auth interceptors, Zod schemas mirroring backend enums, storage wrappers, session store.
- `src/providers/` composes app-wide providers (Query, Tamagui, SafeArea) once in `AppProviders`.
- Full convention doc lives in-repo: `apps/frontend/src/features/README.md`. Copy `src/features/_template/` to start a feature.

## Gotchas hit while building this (don't re-derive)

- **Tamagui typing**: module augmentation target is `@tamagui/web`, NOT `tamagui` or `@tamagui/core`:
  `declare module '@tamagui/web' { interface TamaguiCustomConfig extends AppConfig {} }`
- **Tamagui shorthands**: `@tamagui/config/v4`'s `defaultConfig.settings.onlyAllowShorthands: true` forces shorthand-only props at the type level — use `p`, `items`, `justify`, NOT `padding`, `alignItems`, `justifyContent`.
- **react-native-reanimated v4** needs separate explicit dep `react-native-worklets` + babel plugin `react-native-worklets/plugin` (not `react-native-reanimated/plugin`), must be last in babel plugins array.
- **TanStack Query persister**: `createAsyncStoragePersister` lives in `@tanstack/query-async-storage-persister`, NOT `@tanstack/react-query-persist-client`. `PersistQueryClientProvider` already wraps `QueryClientProvider` — don't nest another one inside it.
- **ESLint 10 incompatible** with `eslint-plugin-react@7.37.5` (pulled by `eslint-config-expo`) — crashes on `react/display-name`. Pin `eslint@^9`. Flat config required (`eslint.config.js`, not `.eslintrc.js`).
- **Metro + Tamagui CSS extraction**: `@tamagui/metro-plugin`'s `outputCSS: './tamagui-web.css'` file must exist as a placeholder BEFORE build/typecheck/export — Metro's resolver looks for it before the plugin writes it. Needs `touch tamagui-web.css` (or equivalent) pre-build, including inside Docker, since it's gitignored/dockerignored as generated output.
- **`@tamagui/core` and `@tamagui/web`** must be explicit direct deps — they're transitive-only otherwise and `expo export` fails to resolve them under pnpm's hoisted linking (`.npmrc`: `node-linker=hoisted`, required for RN/Metro module resolution generally).
- Env vars: `EXPO_PUBLIC_*` are inlined into the client bundle at BUILD time, not runtime — Docker needs them as `ARG`+`ENV`, not just `env_file`. Validated via Zod in `src/shared/config/env.ts`, fails loudly at boot on misconfig.
- `expo-secure-store` throws on web — `src/shared/lib/secure-storage.ts` branches on `Platform.OS === 'web'` to `localStorage`.

## Login-broken incident (2026-08-01) — root causes and fixes

Three independent bugs stacked on top of each other; all now fixed:

1. **Double `/api/v1` prefix** — `auth.api.ts` called `/api/v1/auth/google` on top of an `EXPO_PUBLIC_API_URL` that already ends in `/api/v1` → 404. Was actually already fixed in a prior commit; the *running* Docker containers were just stale (built before the fix). Lesson: `EXPO_PUBLIC_*` is baked at image build time, so `docker compose up` without `--build` after a frontend source change silently keeps serving the old bundle — always `--build` (or `--force-recreate` won't help, only `--build` does).

2. **React error #419 ("Suspense boundary couldn't finish server rendering, switched to client rendering")** — only appeared after completing the OAuth round-trip, never on a cold `/login` load, which made it look auth-related. Actual cause: `app.json` had `web.output: "static"`, which prerenders every route through a **second, server-side (Node) bundle graph** via `@expo/router-server`. That server bundle graph is a separate module instance from the client bundle, and Tamagui's `createTamagui()` singleton config isn't visible across that boundary — logged as `[@tamagui] ... duplicate tamagui instances ... Falling back to default theme from config`. For `/login` specifically this left one Tamagui-internal text/Suspense boundary dehydrated in the prebuilt HTML (verified directly: `grep -o '<!--\$!-->' dist/login.html` — the `!` marks an unresolved boundary; `/`, `/(app)` didn't have one). React then throws #419 hydrating that leftover marker.
   - **Fix**: this app is 100% auth-gated (no public/SEO content), so it doesn't need per-route static prerendering at all. Changed `app.json`'s `web.output` from `"static"` to `"single"` — pure SPA, one `index.html` shell, `<div id="root"></div>` empty, no server render pass, no hydration, so the whole bug class is structurally impossible now.
   - Also hardened `tamagui-provider.tsx`'s `useFonts` gate to only block rendering on native (`Platform.OS !== 'web' && !fontsLoaded`) — web fonts ship via the CSS `@tamagui/metro-plugin` emits, not this JS loader. Didn't fix #419 by itself but is correct regardless and worth keeping.
   - Debugging note: the *minified* prod error only ever shows "Minified React error #419, see react.dev/errors/419" — to get the real message you need the **unminified dev server**, and it must run on a port Google OAuth's registered redirect URI actually allows (here, `:3000`) since the bug only reproduces after a real OAuth round-trip completes.

3. **Uncaught `sw.js:34` fetch rejection** — the service worker's fetch handler had no `.catch()`, so a rejected `fetch()` (offline, or the browser cancelling the SW's shadow fetch in favor of the real navigation's own fetch — a known race on a page's very first load right after the SW installs) surfaced as an unhandled promise rejection in console. Added a `.catch()` that falls back to cache or re-rejects normally. **Gotcha when verifying this class of fix**: an already-registered service worker keeps controlling every tab in that browser profile/incognito session until *all* windows of that session are closed — a plain reload or even a new tab is not enough to pick up a new `sw.js`, which makes an old bug look unfixed. Always fully close all incognito windows (not just the tab) before retesting SW changes.

## Wii Party / Aero / iOS-gloss redesign (2026-08-04)

Full visual restyle — design system + all screens + new nav shell. Full writeup, palette, and the Tamagui v4 gotchas hit (missing `color` token group, `animation`→`transition` rename, shorthand-only quirks, `+html.tsx` inert in `output: "single"`) live in [[frontend-responsive-frutiger-aero]]. Two pre-existing bugs fixed as part of this pass:

- `$standPurple`/`$strawHatRed` were referenced in 6 places but never rendered — `@tamagui/config/v4`'s `defaultConfig` ships no `color` token group at all, so there was nothing to register them into. Fixed by adding `tokens.color` in `tamagui.config.ts`.
- PWA manifest was never linked from the HTML shell (no installability) — fixed via a `public/index.html` override, not `+html.tsx` (which is inert under `web.output: "single"`).

New deps added via CLI: `@tamagui/lucide-icons-2` (not the deprecated `@tamagui/lucide-icons`) + its undeclared peer `react-native-svg`, `@expo-google-fonts/fredoka`, `@expo-google-fonts/nunito`. `tsconfig.json` needed `moduleSuffixes: [".web", ".native", ""]` added for `tsc` to resolve the new platform-split bubble-field files the same way Metro does.

## Frontend test infra (2026-08-04) — jest-expo, first tests ever added

Before this pass the frontend had **zero tests** and CI's `Frontend` job only ran typecheck+lint —
that's exactly why the layout bugs in [[frontend-responsive-frutiger-aero]] (dead `gap`, inverted
`zIndex`, over-constrained absolute) shipped unnoticed for a while: they all pass `tsc` fine.

- Deps: `jest`, `jest-expo`, `@testing-library/react-native`, `@types/jest`, `@types/node`,
  `@react-native/jest-preset` (peer dep `jest-expo` needs explicitly under pnpm). **Pin `jest@^29`,
  not the installed-by-default `^30`** — `jest-expo@57` depends on `jest-mock@^29.2.1` etc. internally;
  mixing majors throws `this._moduleMocker.clearMocksOnScope is not a function` at the very first test
  run. `tsconfig.json` needs `"types": ["jest", "node"]` or every ambient ts-jest ambient global
  (`describe`, `require`, `__dirname`...) disappears at once — Expo's base tsconfig has no explicit
  `types` array, so adding one for jest silently drops the implicit "all installed @types" default.
- `jest.config.js` splits into **two projects**, not one:
  - `logic` (jsdom, `jest-expo/web` preset): pure, non-rendering checks only —
    `tamagui.config.ts`'s real `zIndex` token order, `layout.ts`'s pure clearance/width math, and a
    static-analysis guard that fails if a `—`/`–` shows up in any user-facing string literal
    (`src/test/__tests__/copy.test.ts`). `testMatch` is scoped to exactly those two directories.
  - `native` (`jest-expo` default preset, react-test-renderer): everything that renders a component.
    `Platform.OS` here is a real native value (not react-native-web's fixed `'web'`), so
    `a11yProps()`'s `accessibilityLabel`/`accessibilityRole` path is what's actually exercised —
    querying by `aria-label` would need the `logic`/jsdom project instead, which doesn't render
    Tamagui components at all. Picking the right project for a new test is a judgment call, not a
    convention to encode further — see which one already covers a similar file first.
  - Both need `transformIgnorePatterns: []` **and** an extended `transform` matching `\.mjs$` too —
    jest-expo's own default only transforms `.[jt]sx?`, and pnpm nests every package two levels deep
    (`node_modules/.pnpm/<pkg>/node_modules/<realpkg>`), which breaks the *default* whitelist regex
    (its negative lookahead matches at the outer `node_modules/.pnpm/` segment, before ever reaching
    the real package name) — `@tamagui/*`'s real `.mjs` ESM files hit "Cannot use import statement
    outside a module" without both fixes.
  - `jest.setup.ts` mocks: `react-native-safe-area-context` (fixed zero insets — deliberately, so
    `layout.ts`'s clearance numbers under test are exactly the named constants, no device noise),
    `@tamagui/linear-gradient` (`LinearGradient` → plain `View`; detect it in a rendered tree via its
    `colors` prop, not `type`, since the type name is gone after mocking), `expo-image-picker`,
    `expo-secure-store`, and a **hand-rolled `react-native-reanimated` mock** — the package's own
    `react-native-reanimated/mock` still pulls in `react-native-worklets`' native init chain
    (`loadUnpackers`), which throws outside a real device/simulator. The hand-rolled version only
    covers what `bubble-field.native.tsx`/`use-reduced-motion.ts` actually call
    (`useSharedValue`/`useAnimatedStyle`/`withTiming`/`withRepeat`/`withDelay`/`interpolate`/
    `useReducedMotion`/`runOnJS`) — cheap and enough, don't reach for the official mock again.
  - `window.matchMedia` needs a manual jsdom polyfill — Tamagui's web build (`@tamagui/select`) calls
    it eagerly at *module load*, before any component mounts, so this has to live in `jest.setup.ts`,
    not a per-test `beforeEach`.
- **`render()` from this `@testing-library/react-native` version is `async`.** Forgetting to `await`
  it doesn't throw — `screen.getByText(...)` right after just fails with `` `render` function has not
  been called `` even though the component renders fine a tick later, because `setRenderResult` only
  fires once the awaited `act()` resolves. Cost real time to track down; `src/test/render.tsx` has a
  comment on `renderWithProviders` now so it isn't re-derived.
- **Don't try to globally fake-time animations across a whole file.** The instinct to wrap every test
  in `jest.useFakeTimers()`/`runOnlyPendingTimers()` to settle Tamagui's `transition="bouncy"` springs
  before teardown (they otherwise fire a scheduled frame *after* Jest tears the test down, logging
  "environment has been torn down") backfires: it made `react-test-renderer`'s *second* `render()`
  call in a file return a **silently empty tree** with no error at all — worse than the noisy-but-
  harmless `console.error("...not wrapped in act(...)")` warnings it was meant to prevent. Left
  disabled; the act-warning noise is cosmetic, doesn't fail CI.
- A real RN `<Modal>` (used by `ConfirmSheet`, not mocked — see the fix in
  [[frontend-responsive-frutiger-aero]]) drives its own fade with real `Animated`/real timers; several
  renders of it in **one file** leave an earlier test's still-settling animation corrupting a later
  one. Splitting the `isConfirming` case into its own file
  (`confirm-sheet-confirming.test.tsx`) sidesteps it — different test files get separate module/timer
  state, so isolate rather than trying to force-settle animations.
- Disabled Tamagui `<Button disabled>` renders `aria-disabled` (not RN's `accessibilityState`) even
  under the native project's real `Platform.OS`, and RNTL's `getByLabelText`/`getByRole` queries
  exclude elements it considers inert. Query the still-visible text instead and check
  `.parent.props['aria-disabled']` directly when asserting a disabled state.
- Axios interceptors have no public single-shot invocation API — tests reach into
  `client.interceptors.request/response`'s internal `.handlers[0]` array (`{fulfilled, rejected}`) to
  call a registered interceptor without a real HTTP round trip. Undocumented but stable across axios
  1.x; if it ever breaks, that's the thing to re-derive.
- CI: `.github/workflows/ci.yml`'s `Frontend` job gained a `pnpm test:ci` (`jest --ci`) step after
  Lint — job renamed to "Frontend (typecheck + lint + test)".

Related: [[docker-setup]], [[backend-contract]]
