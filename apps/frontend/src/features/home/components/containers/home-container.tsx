import { useRouter } from 'expo-router'

import { HomeScreen } from '@/features/home/components/presentational/home-screen'
import { useSessionStore } from '@/shared/stores/session.store'

export function HomeContainer() {
  const router = useRouter()
  const session = useSessionStore((state) => state.session)
  const clearSession = useSessionStore((state) => state.clearSession)

  if (!session) return null

  return (
    <HomeScreen
      user={session.user}
      onLogout={() => void clearSession()}
      onOpenProfile={() => router.navigate('/profile')}
    />
  )
}
