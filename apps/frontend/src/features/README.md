# Features

This app follows **Feature-Driven Architecture** combined with **Container /
Presentational** components.

## Feature-Driven

One folder per business capability, e.g. `src/features/stands/`,
`src/features/devil-fruits/`, `src/features/auth/`. Copy `_template/` to
start a new one.

A feature owns everything it needs:

```
features/<feature>/
  api/                 # axios calls + TanStack Query hooks for this feature
  components/
    containers/        # data-wired components (see below)
    presentational/     # pure UI components (see below)
  hooks/               # feature-local hooks that aren't containers
  stores/              # zustand slices scoped to this feature, if any
  types/               # DTOs and view models for this feature
  index.ts             # the feature's public barrel
```

**Rule:** cross-feature imports go only through the other feature's
`index.ts`. Never `import X from '@/features/stands/api/...'` outside
`features/stands` itself. If two features need the same thing, promote it to
`src/shared/`.

## Container / Presentational

Inside `components/`:

- **`presentational/`** — pure components. Props in, JSX out. No
  `useQuery`/`useMutation`, no zustand, no `useRouter`. Built with Tamagui.
  Trivial to reuse and to test in isolation.
- **`containers/`** — the wiring. Call this feature's query/mutation hooks,
  read/write its zustand store, own `react-hook-form` state and navigation,
  then render a presentational component with plain props. No styling
  decisions live here.

Routes under `app/` import **containers only**, never presentational
components directly — `app/` files stay thin route shims.

## Shared

`src/shared/` mirrors this same split (`shared/components/{containers,presentational}`)
for cross-feature primitives, plus:

- `shared/api/` — the axios client, auth/ETag interceptors, error
  normalization, the root query-key factory.
- `shared/lib/` — platform-agnostic wrappers (`secure-storage`,
  `async-storage`) and small hand-written schemas that compose generated
  contracts (e.g. the translation-form superRefines) — never a hand-mirror
  of a backend enum.
- `shared/stores/` — app-wide zustand state (currently: session).
- `shared/config/` — validated environment access.
- `shared/contracts/` — **generated, do not edit.** The single source of
  every wire DTO, enum, and WS command/frame shape, produced from Go by
  `apps/backend/cmd/typegen`. If you need a type or schema for something
  the backend sends or receives, import it from here (or from a feature's
  `types/` file that re-exports it) — never redeclare it. See
  `ObsidianVault/contratos-tipos-generados.md`.
