import { createAsyncStoragePersister } from '@tanstack/query-async-storage-persister'
import { QueryClient } from '@tanstack/react-query'
import { PersistQueryClientProvider } from '@tanstack/react-query-persist-client'
import { useState, type ReactNode } from 'react'

import { AsyncStorage } from '@/shared/lib/async-storage'

// staleTime > 0 avoids refetch-on-mount storms; the backend already does the
// heavy lifting via ETag/304 (see shared/api/etag.ts) so a stale cache is
// cheap to revalidate anyway.
function createQueryClient(): QueryClient {
  return new QueryClient({
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
