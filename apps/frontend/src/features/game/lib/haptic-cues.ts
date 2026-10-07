import type { GameState } from '@/shared/contracts'
import type { HapticCue } from '@/shared/lib/haptics'

// Which vibration (if any) a game-state change deserves. `before` undefined
// means "first state this client has seen for the game" - a page load, a
// reconnect or a resume - which must stay silent: the buzz marks something
// that just happened, not the state you walked into.
export function cueForStateChange(
  before: GameState | undefined,
  after: GameState | undefined
): HapticCue | null {
  if (before === undefined || after === undefined || before === after) return null
  switch (after) {
    case 'VOTING':
    case 'TIEBREAK':
      return 'votingOpened'
    case 'RESOLVING':
      return 'roundResolved'
    case 'FINISHED':
      return 'gameFinished'
    default:
      return null
  }
}
