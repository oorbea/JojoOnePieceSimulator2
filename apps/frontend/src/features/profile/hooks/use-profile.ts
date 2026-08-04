import { useQuery, type Query } from '@tanstack/react-query'

import { getMe } from '@/features/profile/api/profile.api'
import { profileKeys } from '@/features/profile/api/profile.keys'
import type { ProfileUser } from '@/features/profile/types/profile.types'

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
  return useQuery({
    queryKey: profileKeys.me,
    queryFn: getMe,
    refetchInterval: pollInterval,
  })
}
