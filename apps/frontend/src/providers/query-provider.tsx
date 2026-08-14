import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister'
import { MutationCache, QueryClient } from '@tanstack/react-query'
import { PersistQueryClientProvider } from '@tanstack/react-query-persist-client'
import { useState, type ReactNode } from 'react'

import { toAppError } from '@/shared/api/errors'
import { env } from '@/shared/config/env'
import { AsyncStorage } from '@/shared/lib/async-storage'
import { showErrorToast } from '@/shared/lib/toast'

// A persisted cache older than this is discarded outright rather than
// rehydrated - bounds how stale a reopened tab's data can be independent of
// any per-query staleTime.
const PERSIST_MAX_AGE_MS = 24 * 60 * 60 * 1000

// staleTime > 0 avoids refetch-on-mount storms; the backend already does the
// heavy lifting via ETag/304 (see shared/api/etag.ts) so a stale cache is
// cheap to revalidate anyway.
function createQueryClient(): QueryClient {
  return new QueryClient({
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
}

export function QueryProvider({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createQueryClient)
  const [persister] = useState(() => createAsyncStoragePersister({ storage: AsyncStorage }))

  return (
    <PersistQueryClientProvider
      client={queryClient}
      persistOptions={{
        persister,
        maxAge: PERSIST_MAX_AGE_MS,
        // Distinct per deploy (see Dockerfile.frontend/docker-compose.yml)
        // so a new build's persisted cache never rehydrates data shaped for
        // an older build's query keys/response schema - a mismatch here is
        // exactly the "stale until you clear the cache" symptom this exists
        // to prevent, just for the persisted layer instead of the service
        // worker (see public/sw.js for that half).
        buster: env.EXPO_PUBLIC_BUILD_ID,
        dehydrateOptions: {
          // A live game/lobby snapshot is realtime state the WebSocket
          // store owns (see features/game/api/game.keys.ts) - persisting it
          // to AsyncStorage would rehydrate yesterday's roster/state on next
          // launch, the exact "stale until you clear the cache" class the
          // buster above guards against for the schema dimension instead.
          shouldDehydrateQuery: (query) =>
            query.queryKey[1] !== 'games' && query.state.status === 'success',
        },
      }}
    >
      {children}
    </PersistQueryClientProvider>
  )
}
