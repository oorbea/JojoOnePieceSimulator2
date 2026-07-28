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

TypeScript, Expo SDK 57, React Native 0.86, React Native Web, Expo Router (file routing), TanStack Query v5 + `@tanstack/query-async-storage-persister` (persisted via AsyncStorage), Zustand v5, React Hook Form + Zod, Axios, Tamagui v2.5.1 (UI kit — NativeWind dropped, one styling system only), socket.io-client (installed, unwired, no backend WS yet), Expo SecureStore (native) / localStorage (web), Expo PWA (manifest + custom SW).

All deps installed via CLI only, never hand-edited into `package.json` (user hard requirement). Package manager: pnpm.

## Decisions made with user

| Question | Decision |
|---|---|
| UI kit | Tamagui only (dropped Gluestack and NativeWind) |
| Repo layout | Standalone pnpm project, no root workspace |
| Docker | Both dev (hot reload) + prod (static export via nginx) |
| Platforms | Web-first, native-ready (iOS/Android configured, not fully asset-built) |
| Feature scaffolding | Skeleton only, no feature code this pass |
| socket.io | Install only, no wiring |

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

Related: [[docker-setup]], [[backend-contract]]
