import { PersistQueryClientProvider } from '@tanstack/react-query-persist-client'
import type { ReactNode } from 'react'

import { env } from '@/shared/config/env'
import { persister, queryClient } from '@/providers/query-client'
import { useSessionStore } from '@/shared/stores/session.store'

// A persisted cache older than this is discarded outright rather than
// rehydrated - bounds how stale a reopened tab's data can be independent of
// any per-query staleTime.
const PERSIST_MAX_AGE_MS = 24 * 60 * 60 * 1000

// Mirrors AuthService.DevEmailDomain (apps/backend) - a dev-login account's
// email always ends in this reserved, never-resolvable domain. Duplicated
// here rather than imported (there's no generated contract for it: it's not
// wire data, just a naming convention) the same way profile.types.ts already
// mirrors the backend's username sanitizer regex client-side.
const DEV_EMAIL_DOMAIN = '@dev.invalid'

// queryClient/persister are module-level singletons (see query-client.ts) so
// session.store.ts's clearSession() can purge them from outside React on
// logout - this provider just wires them into PersistQueryClientProvider,
// it no longer constructs its own per-mount instances.
export function QueryProvider({ children }: { children: ReactNode }) {
  const userID = useSessionStore((state) => state.session?.user.id)
  const userEmail = useSessionStore((state) => state.session?.user.email)
  // A dev-login tab's persisted snapshot must never rehydrate into a
  // different tab logged in as a different dev account - unlike a real
  // account (one browser, one cookie, one user), several dev sessions share
  // this same browser/localStorage at once (see dev-refresh-token.ts). Only
  // a dev session's buster gets this extra per-user suffix; a real session's
  // buster is untouched, so normal cross-tab same-user behavior is unchanged.
  const buster =
    userEmail?.endsWith(DEV_EMAIL_DOMAIN) && userID
      ? `${env.EXPO_PUBLIC_BUILD_ID}:${userID}`
      : env.EXPO_PUBLIC_BUILD_ID

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
        buster,
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
