# Frontend — JoJo x One Piece Simulator

Progressive Web App built with Expo + React Native Web. Ships to web today;
native (iOS/Android) is configured and ready without restructuring.

## Stack

- **TypeScript**
- **Expo SDK 57** + **React Native** + **React Native Web**
- **Expo Router** (file-based routing, `app/`)
- **TanStack Query** (server state) + persisted cache via AsyncStorage
- **Zustand** (client state — currently just the auth session)
- **Expo SecureStore** (native) / `localStorage` (web) for token storage
- **AsyncStorage** (query cache persistence)
- **React Hook Form** + **Zod** (forms + validation)
- **Tamagui** (UI kit, universal styling, compiled)
- **Axios** (HTTP client)
- **socket.io-client** (installed, unwired — the backend has no websocket
  endpoint yet)
- **pnpm** (package manager)
- **Expo PWA** (manifest + service worker in `public/`)

## Getting started

```bash
pnpm install
cp .env.example .env   # fill in EXPO_PUBLIC_API_URL at minimum
pnpm web                # start the web dev server
pnpm start              # start the Expo dev server (scan QR with Expo Go)
```

## Scripts

| Script | What it does |
|---|---|
| `pnpm web` / `start` / `android` / `ios` | Expo dev server |
| `pnpm build:web` | `expo export -p web` → static PWA in `dist/` |
| `pnpm lint` | ESLint (flat config, `eslint.config.js`) |
| `pnpm format` | Prettier, write mode |
| `pnpm typecheck` | `tsc --noEmit` |

## Environment variables

All client-visible config goes through `EXPO_PUBLIC_*` vars (inlined into the
bundle at build time — never put secrets here). See `.env.example`:

- `EXPO_PUBLIC_API_URL` — backend base URL, e.g. `http://localhost:8080/api/v1`
- `EXPO_PUBLIC_GOOGLE_*_CLIENT_ID` — Google OAuth client IDs (web/iOS/Android)
- `EXPO_PUBLIC_SOCKET_URL` — reserved for when the backend gets a websocket

`src/shared/config/env.ts` validates these with Zod at boot — a missing or
malformed var fails loudly instead of surfacing as a mystery `undefined`
request URL later.

## Docker

Two images, both defined relative to the repo root (matching the backend's
convention):

- **Dev**: `apps/frontend/Dockerfile.dev` — runs the Expo web dev server
  with hot reload; source is meant to be volume-mounted.
- **Prod**: `deployments/docker/Dockerfile.frontend` — multi-stage build,
  `expo export -p web` → static files served by nginx (`nginx.frontend.conf`
  adds SPA fallback + correct caching for the service worker/manifest).

Run everything together:

```bash
docker compose -f ../../deployments/docker-compose.yml up --build frontend
```

The compose file also sets `CORS_ALLOWED_ORIGINS` on the `backend` service to
the frontend's origin — without it the backend denies every browser request.

Note: `pnpm-workspace.yaml` here is **not** a monorepo workspace root, it
only holds this package's build-script allowlist (pnpm requires it for
native postinstall scripts like `esbuild`). This project is standalone.

## Architecture

Feature-Driven, combined with Container/Presentational components. Full
convention is documented in `src/features/README.md` — read that before
adding a feature. Summary:

- `app/` — routes only, thin shims that render a container.
- `src/features/<feature>/` — everything one capability owns (api, hooks,
  stores, types, components). Cross-feature imports go through the feature's
  `index.ts` barrel only. Copy `src/features/_template/` to start a new one.
- `src/features/<feature>/components/{containers,presentational}` —
  containers wire data/navigation/forms; presentational components are pure
  Tamagui UI, props in JSX out.
- `src/shared/` — the same container/presentational split, plus the axios
  client, ETag/auth interceptors, Zod schemas mirroring backend enums,
  storage wrappers, and the session store.
- `src/providers/` — app-wide providers (TanStack Query, Tamagui, safe area)
  composed once in `AppProviders`.

## Talking to the backend

- Base path: `/api/v1`. Auth is `POST /api/v1/auth/google` with a Google ID
  token; the backend returns a plain bearer JWT (24h, **no refresh token**).
  A 401 clears the local session — there is no silent refresh, the user has
  to sign in again.
- Every other endpoint requires `Authorization: Bearer <token>`; writes also
  require an `ADMIN` role.
- Reads are conditional-GET aware: the backend sends `ETag`, the axios client
  (`src/shared/api/interceptors.ts` + `etag.ts`) remembers it per URL and
  sends `If-None-Match` automatically.
- Errors come back as `{ error, details?[] }` and are normalized into
  `AppError` (`src/shared/api/errors.ts`).
