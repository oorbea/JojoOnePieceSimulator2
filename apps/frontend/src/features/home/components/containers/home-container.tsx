import { useRouter } from 'expo-router'

import { HomeScreen } from '@/features/home/components/presentational/home-screen'
import { useSessionStore } from '@/shared/stores/session.store'

export function HomeContainer() {
  const router = useRouter()
  const session = useSessionStore((state) => state.session)

  if (!session) return null

  return (
    <HomeScreen
      user={session.user}
      // Every channel's href is a real, typed route by the time this
      // renders (see /catalog/* under app/(app)/catalog/) - `as never`
      // matches the cast already used elsewhere for routes typedRoutes
      // doesn't fully infer from a plain string.
      onNavigate={(href) => router.navigate(href as never)}
    />
  )
}
