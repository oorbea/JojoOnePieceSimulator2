import { Redirect, Slot } from 'expo-router'

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

  if (!isHydrated) {
    return <LoadingScreen />
  }

  if (!session) return <Redirect href="/login" />

  return (
    <AppShellContainer>
      <Slot />
    </AppShellContainer>
  )
}
