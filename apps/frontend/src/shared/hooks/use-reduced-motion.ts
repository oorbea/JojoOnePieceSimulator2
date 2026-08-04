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
    const listener = (event: MediaQueryListEvent) => setReduced(event.matches)
    query.addEventListener('change', listener)
    return () => query.removeEventListener('change', listener)
  }, [])

  return reduced
}

// Web reads the OS/browser media query directly; native defers to
// Reanimated's own hook, which already tracks the platform accessibility
// setting. Used to freeze the bubble field, the lens flare's breathing
// animation, and to downgrade `animation="bouncy"` to `"quick"`.
export function useReducedMotion(): boolean {
  const webReduced = useWebReducedMotion()
  const nativeReduced = useReanimatedReducedMotion()
  return Platform.OS === 'web' ? webReduced : nativeReduced
}
