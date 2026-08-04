import { useQuery, type Query } from '@tanstack/react-query'
import { useEffect } from 'react'

import { getMe } from '@/features/profile/api/profile.api'
import { profileKeys } from '@/features/profile/api/profile.keys'
import type { ProfileUser } from '@/features/profile/types/profile.types'
import { useSessionStore } from '@/shared/stores/session.store'

// The avatar transcode is a fire-and-forget background job on the backend
// (no websocket exists in this project — see ObsidianVault/backend-contract.md)
// so a PENDING avatar is only ever observed by polling this query. Backs off
// 2s -> 4s -> 8s ... capped at 30s, and gives up after 8 attempts so a stuck
// worker doesn't poll forever in the background.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

function pollInterval(query: Query<ProfileUser>): number | false {
  if (query.state.data?.avatarStatus !== 'PENDING') return false
  if (query.state.dataUpdateCount >= MAX_POLL_ATTEMPTS) return false
  return Math.min(BASE_POLL_INTERVAL_MS * 2 ** query.state.dataUpdateCount, MAX_POLL_INTERVAL_MS)
}

export function useProfile() {
  const query = useQuery({
    queryKey: profileKeys.me,
    queryFn: getMe,
    refetchInterval: pollInterval,
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
