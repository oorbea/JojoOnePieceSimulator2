import { useEffect, useState } from 'react'
import { AppState, type AppStateStatus } from 'react-native'

// Placeholder shared hook — exposes RN's AppState as reactive state so
// features can react to foreground/background transitions (e.g. pausing
// polling, refetching on resume) without each one wiring its own listener.
export function useAppState(): AppStateStatus {
  const [state, setState] = useState<AppStateStatus>(AppState.currentState)

  useEffect(() => {
    const subscription = AppState.addEventListener('change', setState)
    return () => subscription.remove()
  }, [])

  return state
}
