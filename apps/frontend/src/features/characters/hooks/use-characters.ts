import { useQuery, type Query } from '@tanstack/react-query'
import { useRef } from 'react'
import { Platform } from 'react-native'

import { getJojoCharacters, getOnePieceCharacters } from '@/features/characters/api/characters.api'
import { characterKeys } from '@/features/characters/api/characters.keys'
import type {
  JojoCharacterFilters,
  JojoCharacterResponse,
  OnePieceCharacterFilters,
  OnePieceCharacterResponse,
} from '@/features/characters/types/characters.types'

// Same polling rationale as use-stages.ts's useStages: PictureEventsBridge
// covers web via SSE, this is native's fallback only. Inlined per hook
// (rather than factored into a shared helper taking the ref) since handing
// a ref into a function during render trips react-hooks/refs.
const MAX_POLL_ATTEMPTS = 8
const MAX_POLL_INTERVAL_MS = 30_000
const BASE_POLL_INTERVAL_MS = 2_000

// `enabled` lets a manga-exclusive screen (CharactersContainer) mount both
// hooks unconditionally (required - hooks can't be called conditionally)
// while only the active kind actually hits the network; the inactive
// kind's query just sits idle until the toggle flips.
export function useJojoCharacters(filters?: JojoCharacterFilters, enabled = true) {
  const pollAttempts = useRef(0)
  return useQuery({
    queryKey: characterKeys.jojo.list(filters),
    queryFn: () => getJojoCharacters(filters),
    enabled,
    refetchInterval:
      Platform.OS === 'web'
        ? undefined
        : (query: Query<JojoCharacterResponse[]>) => {
            const hasPending = query.state.data?.some((c) => c.pictureStatus === 'PENDING')
            if (!hasPending) {
              pollAttempts.current = 0
              return false
            }
            if (pollAttempts.current >= MAX_POLL_ATTEMPTS) return false
            const interval = Math.min(
              BASE_POLL_INTERVAL_MS * 2 ** pollAttempts.current,
              MAX_POLL_INTERVAL_MS
            )
            pollAttempts.current += 1
            return interval
          },
    refetchIntervalInBackground: true,
  })
}

export function useOnePieceCharacters(filters?: OnePieceCharacterFilters, enabled = true) {
  const pollAttempts = useRef(0)
  return useQuery({
    queryKey: characterKeys.onePiece.list(filters),
    queryFn: () => getOnePieceCharacters(filters),
    enabled,
    refetchInterval:
      Platform.OS === 'web'
        ? undefined
        : (query: Query<OnePieceCharacterResponse[]>) => {
            const hasPending = query.state.data?.some((c) => c.pictureStatus === 'PENDING')
            if (!hasPending) {
              pollAttempts.current = 0
              return false
            }
            if (pollAttempts.current >= MAX_POLL_ATTEMPTS) return false
            const interval = Math.min(
              BASE_POLL_INTERVAL_MS * 2 ** pollAttempts.current,
              MAX_POLL_INTERVAL_MS
            )
            pollAttempts.current += 1
            return interval
          },
    refetchIntervalInBackground: true,
  })
}
