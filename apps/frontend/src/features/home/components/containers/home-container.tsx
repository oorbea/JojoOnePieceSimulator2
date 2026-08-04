import { useRouter } from 'expo-router'

import { HomeScreen } from '@/features/home/components/presentational/home-screen'
import { useSessionStore } from '@/shared/stores/session.store'

export function HomeContainer() {
  const router = useRouter()
  const session = useSessionStore((state) => state.session)

  if (!session) return null

  return <HomeScreen user={session.user} onOpenProfile={() => router.navigate('/profile')} />
}
