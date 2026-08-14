import { useRouter } from 'expo-router'

import { BrowseLobbiesScreen } from '@/features/game/components/presentational/browse-lobbies-screen'
import { useJoinGameById } from '@/features/game/hooks/use-join-game'
import { usePublicLobbies } from '@/features/game/hooks/use-public-lobbies'

export function BrowseLobbiesContainer() {
  const router = useRouter()
  const query = usePublicLobbies()
  const joinGame = useJoinGameById()

  return (
    <BrowseLobbiesScreen
      onBack={() => router.back()}
      onCreate={() => router.replace('/play/create' as never)}
      lobbies={query.data?.items ?? []}
      isLoading={query.isLoading}
      isFetching={query.isFetching}
      isError={query.isError}
      onRefresh={() => void query.refetch()}
      onJoin={(gameId) =>
        joinGame.mutate(gameId, {
          onSuccess: (data) => router.replace(`/play/${data.game.id}` as never),
        })
      }
    />
  )
}
