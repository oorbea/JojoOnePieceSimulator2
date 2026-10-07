import { useEffect, useRef } from 'react'

import { cueForStateChange } from '@/features/game/lib/haptic-cues'
import type { GameState } from '@/shared/contracts'
import { vibrate } from '@/shared/lib/haptics'

type Params = {
  state: GameState | undefined
  /** Distinct, non-negative key each time one of YOUR OWN powers lands in
   * the sorteo (the slot index); -1 whenever it isn't landing. */
  ownLandingKey: number
}

// Vibrates on the moments worth looking up for: your power landing, voting
// opening, the round resolving, the game ending. See ../lib/haptic-cues.ts.
export function useGameHaptics({ state, ownLandingKey }: Params) {
  const previousState = useRef<GameState | undefined>(undefined)
  useEffect(() => {
    const cue = cueForStateChange(previousState.current, state)
    previousState.current = state
    if (cue) vibrate(cue)
  }, [state])

  useEffect(() => {
    if (ownLandingKey >= 0) vibrate('powerLand')
  }, [ownLandingKey])
}
