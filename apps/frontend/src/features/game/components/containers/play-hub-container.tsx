import { useRouter } from 'expo-router'

import { PlayHubScreen } from '@/features/game/components/presentational/play-hub-screen'

export function PlayHubContainer() {
  const router = useRouter()

  return (
    <PlayHubScreen
      onCreate={() => router.navigate('/play/create' as never)}
      onJoin={() => router.navigate('/play/join' as never)}
      onBrowse={() => router.navigate('/play/browse' as never)}
    />
  )
}
