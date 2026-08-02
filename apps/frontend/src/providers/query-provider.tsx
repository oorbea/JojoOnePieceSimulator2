import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister'
import { MutationCache, QueryClient } from '@tanstack/react-query'
import { PersistQueryClientProvider } from '@tanstack/react-query-persist-client'
import { useState, type ReactNode } from 'react'

import { toAppError } from '@/shared/api/errors'
import { AsyncStorage } from '@/shared/lib/async-storage'
import { showErrorToast } from '@/shared/lib/toast'

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
    <PersistQueryClientProvider client={queryClient} persistOptions={{ persister }}>
      {children}
    </PersistQueryClientProvider>
  )
}
