import { useEffect } from 'react'
import { Platform } from 'react-native'

import { useGameActivityStore } from '@/shared/stores/game-activity.store'

import { createWakeLock, type WakeLockApi } from '../lib/wake-lock'

// Holds the screen awake while the player is in a lobby or a game (see
// ../lib/wake-lock.ts). No-op off web and where the Wake Lock API is missing
// (older browsers, iOS < 16.4).
export function useWakeLock() {
  const wanted = useGameActivityStore((state) => state.activity !== 'none')

  useEffect(() => {
    if (!wanted || Platform.OS !== 'web') return
    if (typeof navigator === 'undefined' || !('wakeLock' in navigator)) return

    const lock = createWakeLock(
      (navigator as Navigator & { wakeLock: WakeLockApi }).wakeLock,
      document
    )
    lock.start()
    return () => lock.stop()
  }, [wanted])
}
