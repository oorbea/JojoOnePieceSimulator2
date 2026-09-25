import { Redirect } from 'expo-router'

import { DevLoginContainer } from '@/features/auth'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { env } from '@/shared/config/env'
import { useSessionStore } from '@/shared/stores/session.store'

// Hidden local-only route - not linked from anywhere in the app, reached by
// typing the URL directly. Redirects to /login (never renders, not even
// briefly) unless EXPO_PUBLIC_DEV_AUTH is set, which only ever happens via
// docker-compose.dev.yml's build args. The backend route this screen calls
// (POST /auth/dev-login) has its own, independent defenses - see
// AuthEndpoints.Routes/config.Config.DevAuthBypass - so this flag only ever
// controls whether the UI exists, never the real security boundary.
export default function DevLoginRoute() {
  const session = useSessionStore((state) => state.session)
  const isHydrated = useSessionStore((state) => state.isHydrated)

  if (!env.EXPO_PUBLIC_DEV_AUTH) return <Redirect href="/login" />
  if (!isHydrated) return <LoadingScreen />
  if (session) return <Redirect href="/" />

  return <DevLoginContainer />
}
