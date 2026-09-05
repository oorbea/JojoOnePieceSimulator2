import { Redirect } from 'expo-router'

import { LoginContainer } from '@/features/auth'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { useSessionStore } from '@/shared/stores/session.store'

export default function LoginRoute() {
  const session = useSessionStore((state) => state.session)
  const isHydrated = useSessionStore((state) => state.isHydrated)

  // Hydration now involves an async silent-refresh round trip (see
  // session.store.ts), so it can no longer be assumed instantaneous - render
  // the loading screen instead of flashing the login UI while it's pending.
  if (!isHydrated) return <LoadingScreen />
  if (session) return <Redirect href="/" />

  return <LoginContainer />
}
