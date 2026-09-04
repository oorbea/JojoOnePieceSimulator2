import { useEffect, useState } from 'react'
import { Platform } from 'react-native'
import { useReducedMotion as useReanimatedReducedMotion } from 'react-native-reanimated'

function getInitialReducedMotion(): boolean {
  if (typeof window === 'undefined' || !window.matchMedia) return false
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function useWebReducedMotion(): boolean {
  const [reduced, setReduced] = useState(getInitialReducedMotion)

  useEffect(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return
    const query = window.matchMedia('(prefers-reduced-motion: reduce)')
    // Guard the modern listener API - this hook is now called
    // unconditionally by every AppShell render (ChannelBarIndicator), so
    // it also runs under RN's own test environment where a `window` global
    // exists with a matchMedia stub that doesn't implement
    // addEventListener/removeEventListener. Real browsers always have
    // both; this only skips live-updating the value where they're absent.
    if (typeof query.addEventListener !== 'function') return
    const listener = (event: MediaQueryListEvent) => setReduced(event.matches)
    query.addEventListener('change', listener)
    return () => query.removeEventListener('change', listener)
  }, [])

  return reduced
}

// Web reads the OS/browser media query directly; native defers to
// Reanimated's own hook, which already tracks the platform accessibility
// setting. Used to freeze the bubble field, the lens flare's breathing
// animation, to downgrade `animation="bouncy"` to `"quick"`, and to
// collapse ChannelBarIndicator's slide to an instant snap.
export function useReducedMotion(): boolean {
  const webReduced = useWebReducedMotion()
  const nativeReduced = useReanimatedReducedMotion()
  return Platform.OS === 'web' ? webReduced : nativeReduced
}
