import { useEffect, useRef } from 'react'
import { Platform } from 'react-native'

import { installBackGuard } from '@/shared/lib/back-guard'
import { isStandaloneDisplay } from '@/shared/lib/pwa-display'

// While `enabled`, a system back press in the installed app calls
// `onBackAttempt` (to ask for confirmation) instead of navigating away. Only
// active in the standalone PWA: in a normal browser tab the back button is
// the user's escape hatch and an extra history entry would just be annoying.
export function useBackButtonGuard(enabled: boolean, onBackAttempt: () => void) {
  // Latest callback without re-installing the guard (and pushing another
  // history entry) every render.
  const callback = useRef(onBackAttempt)
  useEffect(() => {
    callback.current = onBackAttempt
  })

  useEffect(() => {
    if (!enabled || Platform.OS !== 'web' || !isStandaloneDisplay()) return
    return installBackGuard(window, () => callback.current())
  }, [enabled])
}
