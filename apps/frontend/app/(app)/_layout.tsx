import { Redirect, Slot } from 'expo-router'

import { ManualOverlayProvider } from '@/features/manual'
import { useNotificationNavigation, usePushBootstrap } from '@/features/pwa'
import { AppShellContainer } from '@/shared/components/containers/app-shell-container'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { useSessionStore } from '@/shared/stores/session.store'

// Auth guard for every route under this group — unauthenticated users never
// see the tree below, they're redirected before Slot renders anything. The
// nav shell mounts HERE (not in the root layout) so it only ever renders
// for a signed-in user and never on /login or +not-found.
export default function AppGroupLayout() {
  const session = useSessionStore((state) => state.session)
  const isHydrated = useSessionStore((state) => state.isHydrated)

  // Push notifications: load availability once signed in (and drop this
  // device's subscription at sign-out), and route a tapped notification to
  // its game inside the already-open app.
  usePushBootstrap(!!session)
  useNotificationNavigation()

  if (!isHydrated) {
    return <LoadingScreen />
  }

  if (!session) return <Redirect href="/login" />

  return (
    <AppShellContainer>
      <ManualOverlayProvider>
        <Slot />
      </ManualOverlayProvider>
    </AppShellContainer>
  )
}
