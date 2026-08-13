import { useRouter, useLocalSearchParams } from 'expo-router'
import { useState } from 'react'
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

  // Adjusts state when the deep-link param changes, during render rather
  // than in an effect (React's own recommended pattern for this) - avoids
  // both the extra commit-then-re-render an effect would cost and the
  // `react-hooks/set-state-in-effect` lint rule, which flags an
  // unconditional `setState` inside `useEffect` as a cascading-render risk.
  const [seededParamCode, setSeededParamCode] = useState(params.code)
  if (params.code !== seededParamCode) {
    setSeededParamCode(params.code)
    if (params.code) setCode(normalizeCode(params.code))
  }

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
