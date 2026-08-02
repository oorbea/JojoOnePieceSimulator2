import { Redirect, Slot } from 'expo-router'
import { YStack, Spinner } from 'tamagui'

import { useSessionStore } from '@/shared/stores/session.store'

// Auth guard for every route under this group — unauthenticated users never
// see the tree below, they're redirected before Slot renders anything.
export default function AppGroupLayout() {
  const session = useSessionStore((state) => state.session)
  const isHydrated = useSessionStore((state) => state.isHydrated)

  if (!isHydrated) {
    return (
      <YStack flex={1} items="center" justify="center">
        <Spinner size="large" />
      </YStack>
    )
  }

  if (!session) return <Redirect href="/login" />

  return <Slot />
}
