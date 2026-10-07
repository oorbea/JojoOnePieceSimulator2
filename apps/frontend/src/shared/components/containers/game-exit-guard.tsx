import { useEffect, useRef } from 'react'
import { Platform } from 'react-native'

import { installBackGuard, type BackGuard } from '@/shared/lib/back-guard'
import { isStandaloneDisplay } from '@/shared/lib/pwa-display'
import { useExitGuardStore } from '@/shared/stores/exit-guard.store'

type Props = {
  /** A game is in progress: leaving needs the player's OK. */
  enabled: boolean
  /** Leaves the game for good (LEAVE + clearing client state), no navigation. */
  onLeave: () => void
  /** Where the system back button / swipe goes once the player confirmed. */
  onBackConfirmed: () => void
}

// Renders nothing. While `enabled` it protects a game in progress:
//  - registers with the exit-guard store, so the app shell's navigation and
//    logout ask before leaving (they have no business knowing about games);
//  - in the installed app, also catches the system back button / edge swipe
//    (see lib/back-guard.ts) - in a normal browser tab back is the user's
//    escape hatch and extra history entries would just be in the way.
// It is a component, not a hook call, so a screen can mount it from JSX placed
// after its own early returns, where the handlers it needs are defined.
export function GameExitGuard({ enabled, onLeave, onBackConfirmed }: Props) {
  const register = useExitGuardStore((state) => state.register)
  const request = useExitGuardStore((state) => state.request)

  // Latest callbacks without re-installing the guard (and pushing history
  // entries again) every render.
  const callbacks = useRef({ onLeave, onBackConfirmed })
  useEffect(() => {
    callbacks.current = { onLeave, onBackConfirmed }
  })

  useEffect(() => {
    if (!enabled) return

    let guard: BackGuard | null = null
    if (Platform.OS === 'web' && isStandaloneDisplay()) {
      guard = installBackGuard(window, () => request(() => callbacks.current.onBackConfirmed()))
    }
    register({
      leaveGame: () => callbacks.current.onLeave(),
      releaseHistory: () => (guard ? guard.release() : Promise.resolve()),
    })

    return () => {
      register(null)
      void guard?.release()
    }
  }, [enabled, register, request])

  return null
}
