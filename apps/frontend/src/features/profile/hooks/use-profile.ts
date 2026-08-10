import { useQuery, type Query } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { Platform } from 'react-native'

import { getMe } from '@/features/profile/api/profile.api'
import { profileKeys } from '@/features/profile/api/profile.keys'
import type { ProfileUser } from '@/features/profile/types/profile.types'
import { useSessionStore } from '@/shared/stores/session.store'

// On web, PictureEventsBridge (src/providers/picture-events-bridge.tsx)
// pushes avatar-pipeline completion via SSE instead of this polling every
// query - see use-stands.ts's useStands for the full reasoning. React
// Native has no EventSource, so native keeps polling as its fallback: 2s ->
// 4s -> 8s ... capped at 30s, giving up after 8 attempts so a stuck worker
// doesn't poll forever in the background.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

export function useProfile() {
  // Attempt count must live outside query.state - see use-stands.ts's
  // useStands for why `dataUpdateCount` (cumulative for this query key's
  // whole lifetime, and persisted across reloads by
  // PersistQueryClientProvider) silently disables polling forever once 8
  // refetches have ever happened, for any reason. A ref scoped to this hook
  // instance resets whenever the avatar isn't PENDING.
  const pollAttempts = useRef(0)

  const query = useQuery({
    queryKey: profileKeys.me,
    queryFn: getMe,
    refetchInterval:
      Platform.OS === 'web'
        ? undefined
        : (query: Query<ProfileUser>) => {
            if (query.state.data?.avatarStatus !== 'PENDING') {
              pollAttempts.current = 0
              return false
            }
            if (pollAttempts.current >= MAX_POLL_ATTEMPTS) return false
            const interval = Math.min(BASE_POLL_INTERVAL_MS * 2 ** pollAttempts.current, MAX_POLL_INTERVAL_MS)
            pollAttempts.current += 1
            return interval
          },
    refetchIntervalInBackground: true,
  })

  const session = useSessionStore((state) => state.session)
  const setSession = useSessionStore((state) => state.setSession)

  // useUploadAvatar's mutation only syncs the session with the 202 response
  // - which still carries the *previous* avatar, since the worker hasn't
  // finished yet (see profile.api.ts's PATCH .../picture doc). Once this
  // query's poll observes the finished (READY/FAILED) avatar, or any other
  // out-of-band change to username/avatar, mirror it into the session store
  // too - otherwise HomeScreen/the nav shell (which read the session, not
  // this query) keep showing the stale value even after a page reload.
  useEffect(() => {
    if (!query.data || !session) return
    const picture = query.data.avatar || null
    if (session.user.username === query.data.username && session.user.picture === picture) return
    void setSession({
      ...session,
      user: { ...session.user, username: query.data.username, picture },
    })
    // Only re-sync when the server data actually changes; session/setSession
    // are stable store references and including them would re-run this on
    // every unrelated session update.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query.data])

  return query
}
