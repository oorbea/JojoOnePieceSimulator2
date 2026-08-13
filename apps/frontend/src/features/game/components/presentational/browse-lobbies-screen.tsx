import { ChevronLeft, RefreshCw, TriangleAlert } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { ActivityIndicator } from 'react-native'
import { XStack } from 'tamagui'

import { PublicLobbyCard } from '@/features/game/components/presentational/public-lobby-card'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { SpeechBubble } from '@/shared/components/presentational/speech-bubble'
import type { PublicLobby } from '@/features/game/types/game.types'

type Props = {
  onBack: () => void
  onCreate: () => void
  lobbies: PublicLobby[]
  isLoading: boolean
  isFetching: boolean
  isError: boolean
  onRefresh: () => void
  onJoin: (gameId: string) => void
}

export function BrowseLobbiesScreen({ onBack, onCreate, lobbies, isLoading, isFetching, isError, onRefresh, onJoin }: Props) {
  const { t } = useTranslation()

  return (
    <PageShell align="top" scroll maxWidth={960}>
      <XStack width="100%" items="center" justify="space-between">
        <XStack items="center" gap="$3">
          <GlossButton tone="glass" btnSize="sm" shape="circle" onPress={onBack} accessibilityLabel={t('common.cancel')}>
            <ChevronLeft size={18} color="$panelText" />
          </GlossButton>
          <GlowText level="title">{t('game.browse.title')}</GlowText>
        </XStack>
        <GlossButton tone="glass" btnSize="sm" shape="circle" onPress={onRefresh} accessibilityLabel={t('game.browse.refresh')}>
          {isFetching ? <ActivityIndicator /> : <RefreshCw size={16} color="$panelText" />}
        </GlossButton>
      </XStack>

      {isLoading ? <ActivityIndicator /> : null}

      {isError ? (
        <SpeechBubble maxW={420} self="center" items="center">
          <XStack items="center" gap="$2">
            <TriangleAlert size={16} color="$strawHatRedDeep" />
            <GlowText level="label" align="center">{t('game.browse.error')}</GlowText>
          </XStack>
          <GlossButton tone="glass" btnSize="sm" onPress={onRefresh} accessibilityLabel={t('game.browse.retry')}>
            {t('game.browse.retry')}
          </GlossButton>
        </SpeechBubble>
      ) : null}

      {!isLoading && !isError && lobbies.length === 0 ? (
        <SpeechBubble maxW={420} self="center" items="center">
          <GlowText level="label" align="center">{t('game.browse.empty')}</GlowText>
          <GlossButton tone="green" btnSize="sm" onPress={onCreate} accessibilityLabel={t('game.browse.emptyCta')}>
            {t('game.browse.emptyCta')}
          </GlossButton>
        </SpeechBubble>
      ) : null}

      <XStack flexWrap="wrap" gap="$4" justify="center">
        {lobbies.map((lobby) => (
          <PublicLobbyCard key={lobby.gameId} lobby={lobby} onJoin={() => onJoin(lobby.gameId)} />
        ))}
      </XStack>
    </PageShell>
  )
}
