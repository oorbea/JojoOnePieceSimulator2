---
title: "Feature: Google login + global error handling"
tags:
  - project
  - jojo-onepiece-simulator
  - frontend
  - auth
  - decision
---

# Feature — Google Login + Global Error Handling

## What was built (2026-07-31)

First feature code in the frontend (previously skeleton-only):

- **`src/features/auth/`** — Google OAuth login via `expo-auth-session`, Frutiger Aero styled login screen, wired to existing session store.
- **`src/features/home/`** — placeholder post-login screen (avatar, name, email, role, logout).
- **Auth-gated routing** — `app/login.tsx` (public) + `app/(app)/_layout.tsx` (guard, redirects to `/login` if no session) + `app/(app)/index.tsx` (home).
- **Global error handling**: `ErrorBoundary` (class component) + `ErrorFallback` UI wrapping `AppProviders`; `MutationCache.onError` in `query-provider.tsx` auto-toasts failed mutations via `burnt`.
- New deps: `@tamagui/linear-gradient`, `burnt`.

## Gotchas hit (don't re-derive)

- **Tamagui v4 shorthands are a fixed, small list** — no `br`, `bw`, `bc`. Actual set: `text, b, bg, content, grow, items, justify, l, m, maxH, maxW, mb, minH, minW, ml, mr, mt, mx, my, p, pb, pl, pr, pt, px, py, r, rounded, select, self, shrink, t, z`. Use `rounded` for borderRadius, full `borderWidth`/`borderColor` — those two have no shorthand at all in v4.
- **`@tamagui/linear-gradient`'s `LinearGradient` component does NOT accept Tamagui stack shorthand props** (`items`, `justify`, etc.) directly — only `flex`, `colors`, `start`, `end` and similar. Wrap a plain `YStack` *inside* it for centering/padding instead of passing those props to `LinearGradient` itself.
- **react-compiler ESLint rule** flags any `setState` called synchronously in the body of a `useEffect` (even in an early-return branch) as "can trigger cascading renders." Fix: wrap the whole effect body in an async IIFE so all `setState` calls happen inside promise-chain callbacks, not synchronously in the effect function body.
- **`pnpm` is not on PATH** on this machine — only `corepack pnpm` works (`corepack enable` fails with EPERM on an unrelated yarn shim in nvm4w's node dir). Always invoke as `corepack pnpm <cmd>` here, not bare `pnpm`.
- **Chrome browser automation (claude-in-chrome) cannot reach `localhost:8081`** (Expo dev server) from this sandboxed session — network isolation between the shell/task execution context and the automated browser. Verification for this feature relied on `tsc --noEmit` + `eslint` only; live browser testing of the OAuth flow needs to be done manually by the user (also needs real Google Client ID — `.env` is gitignored, uses a placeholder locally).

Related: [[frontend-stack]], [[backend-contract]], [[frontend-responsive-frutiger-aero]]
