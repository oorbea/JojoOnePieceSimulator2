import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister'
import { MutationCache, QueryClient } from '@tanstack/react-query'

import { toAppError } from '@/shared/api/errors'
import { AsyncStorage } from '@/shared/lib/async-storage'
import { showErrorToast } from '@/shared/lib/toast'
import { registerQueryCachePurge } from '@/shared/stores/query-cache-purge'

// Module-level singletons (not created inside QueryProvider's useState)
// specifically so clearPersistedQueryCache() below can reach them from
// outside React - session.store.ts's clearSession() calls it on logout, and
// a zustand store action isn't a component that could read state/context.
// There is still only ever one QueryClient/persister for the whole app;
// QueryProvider just renders PersistQueryClientProvider around these instead
// of constructing its own.
export const queryClient = new QueryClient({
  // Global mutation error toast so every feature gets user-friendly error
  // display for free — no per-mutation onError boilerplate needed.
  mutationCache: new MutationCache({
    onError: (error) => showErrorToast(toAppError(error)),
  }),
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      retry: 1,
    },
  },
})

export const persister = createAsyncStoragePersister({ storage: AsyncStorage })

// Registers the real purge with shared/stores/query-cache-purge.ts, which
// session.store.ts's clearSession() calls on logout - see that module's doc
// for why this indirection exists instead of session.store.ts importing this
// module directly. On logout, another user's profile/lobby data must not
// linger either in memory or in the persisted AsyncStorage/localStorage
// snapshot (see query-provider.tsx's PERSIST_MAX_AGE_MS doc for why that
// snapshot exists at all): queryClient.clear() drops the in-memory cache,
// persister.removeClient() deletes the on-disk snapshot so a fresh login
// never rehydrates the previous user's data.
registerQueryCachePurge(async () => {
  queryClient.clear()
  await persister.removeClient()
})
