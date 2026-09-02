# Providers

App-wide React providers, composed once in `app-providers.tsx` and mounted
by `app/_layout.tsx`. Composition order (outer to inner):

```
I18nextProvider > SafeAreaProvider > TamaguiProvider > QueryProvider
  > ErrorBoundary(children) + ToasterMount + PictureEventsBridge
```

Add new app-wide providers to `AppProviders`, not to the layout file.

- **`tamagui-provider.tsx`** — wraps Tamagui's `TamaguiProvider` with the
  project's `tamagui.config`, loads the Fredoka/Nunito fonts (native only —
  web gets fonts via CSS from `@tamagui/metro-plugin`), and resolves
  light/dark/system theme from `shared/stores/theme.store`.
- **`query-provider.tsx`** — creates the TanStack Query `QueryClient` and
  wraps it in `PersistQueryClientProvider`, persisting the cache to
  `AsyncStorage` (24h max age, busted per build via `EXPO_PUBLIC_BUILD_ID`).
  Live game/lobby queries (`queryKey[1] === 'games'`) are excluded from
  persistence — that state belongs to the WebSocket store. Also wires a
  global `MutationCache.onError` so every mutation gets an error toast for
  free.
- **`picture-events-bridge.tsx`** — renderless component, web-only. Opens an
  `EventSource` to the backend's `/events` SSE stream while a session
  exists, and invalidates the relevant TanStack Query keys (stands, devil
  fruits, stages, profile) as picture-pipeline events arrive. Reconnects
  with exponential backoff (2s → 30s cap) and does a full invalidation on
  reconnect, since the stream has no replay log. React Native has no
  built-in `EventSource`, so native falls back to polling in the relevant
  feature hooks instead.
- **`toaster-mount.web.tsx`** / **`toaster-mount.native.tsx`** — platform
  split via Metro's `.web`/`.native` resolution. Web mounts `burnt`'s
  sonner-backed `<Toaster/>` (required once at the root or `toast()` calls
  fail silently); native is a no-op since `burnt` renders through the
  platform's native overlay there.
