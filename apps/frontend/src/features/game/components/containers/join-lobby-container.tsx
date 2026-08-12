import { useRouter, useLocalSearchParams } from 'expo-router'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { JoinLobbyScreen } from '@/features/game/components/presentational/join-lobby-screen'
import { normalizeCode } from '@/features/game/lib/game-code'
import { useJoinGameByCode } from '@/features/game/hooks/use-join-game'
import { useLobbyPreview } from '@/features/game/hooks/use-lobby-preview'
import { toAppError } from '@/shared/api/errors'

export function JoinLobbyContainer() {
  const router = useRouter()
  const { t } = useTranslation()
  const params = useLocalSearchParams<{ code?: string }>()
  const [code, setCode] = useState(() => normalizeCode(params.code ?? ''))

  useEffect(() => {
    if (params.code) setCode(normalizeCode(params.code))
    // eslint-disable-next-line react-hooks/exhaustive-deps -- only react to the deep-link param changing
  }, [params.code])

  const preview = useLobbyPreview(code)
  const joinGame = useJoinGameByCode()

  const previewError = preview.isError
    ? t(`errors.${toAppError(preview.error).code}`, { defaultValue: t('game.join.notFound') })
    : undefined

  return (
    <JoinLobbyScreen
      onBack={() => router.back()}
      code={code}
      onChangeCode={(raw) => setCode(normalizeCode(raw))}
      preview={preview.data}
      previewLoading={preview.isFetching}
      previewError={previewError}
      joining={joinGame.isPending}
      onSubmit={() =>
        joinGame.mutate(code, {
          onSuccess: (data) => router.replace(`/play/${data.game.id}` as never),
        })
      }
    />
  )
}
