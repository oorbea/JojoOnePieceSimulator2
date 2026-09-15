import { useRouter } from 'expo-router'

import { useMyGame } from '@/features/game/hooks/use-my-game'
import { HomeScreen } from '@/features/home/components/presentational/home-screen'
import { useSessionStore } from '@/shared/stores/session.store'

export function HomeContainer() {
  const router = useRouter()
  const session = useSessionStore((state) => state.session)
  // Only asked for once there's a session to resume a game for - see
  // GET /games/me's own doc (the other half of the disconnect grace period:
  // a player who closed the tab must be able to find their way back in).
  const { data: myGame } = useMyGame(!!session)

  if (!session) return null

  return (
    <HomeScreen
      user={session.user}
      resumeGameId={myGame?.game.id ?? null}
      onResumeGame={(gameId) => router.navigate(`/play/${gameId}` as never)}
      // Every channel's href is a real, typed route by the time this
      // renders (see /catalog/* under app/(app)/catalog/) - `as never`
      // matches the cast already used elsewhere for routes typedRoutes
      // doesn't fully infer from a plain string.
      onNavigate={(href) => router.navigate(href as never)}
    />
  )
}
