import { create } from 'zustand'

import type { GameState } from '@/shared/contracts'

// What the player is doing right now, in the coarse terms app-wide PWA
// behaviour cares about. Lives in shared/ (not in the game feature) so the
// pwa feature can read it without importing across features: the game room
// writes it, the service-worker update flow, the wake lock and the
// back-button guard read it.
//
//   none    - not in a game room (or the game already ended)
//   lobby   - waiting in a lobby; safe to reload, reconnects on its own
//   playing - the game has started and is not finished; never interrupt it
export type GameActivity = 'none' | 'lobby' | 'playing'

export function activityForState(state: GameState | null | undefined): GameActivity {
  switch (state) {
    case 'LOBBY':
      return 'lobby'
    case 'ASSIGNING':
    case 'SUMMARY':
    case 'VOTING':
    case 'TIEBREAK':
    case 'RESOLVING':
      return 'playing'
    default:
      return 'none'
  }
}

type GameActivityState = {
  activity: GameActivity
  setActivity: (activity: GameActivity) => void
}

export const useGameActivityStore = create<GameActivityState>((set) => ({
  activity: 'none',
  setActivity: (activity) => set({ activity }),
}))
