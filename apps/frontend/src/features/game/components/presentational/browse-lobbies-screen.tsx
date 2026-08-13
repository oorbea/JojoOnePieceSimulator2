import { ChevronLeft, Compass, RefreshCw, TriangleAlert } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { ActivityIndicator } from 'react-native'
import { XStack } from 'tamagui'

import { PublicLobbyCard } from '@/features/game/components/presentational/public-lobby-card'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
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
        // Same empty/error-state recipe as the Stands admin screen (a flat
        // GlassPanel card, not a SpeechBubble) - a plain informational card
        // reads clearer here than a dialogue bubble with a tail pointing at
        // nothing in particular.
        <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
          <TriangleAlert size={28} color="$strawHatRed" />
          <GlowText level="label" align="center">{t('game.browse.error')}</GlowText>
          <GlossButton tone="blue" btnSize="sm" onPress={onRefresh} accessibilityLabel={t('game.browse.retry')}>
            {t('game.browse.retry')}
          </GlossButton>
        </GlassPanel>
      ) : null}

      {!isLoading && !isError && lobbies.length === 0 ? (
        <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
          <Compass size={28} color="$grapeSoda" />
          <GlowText level="label" align="center">{t('game.browse.empty')}</GlowText>
          <GlossButton tone="green" btnSize="sm" onPress={onCreate} accessibilityLabel={t('game.browse.emptyCta')}>
            {t('game.browse.emptyCta')}
          </GlossButton>
        </GlassPanel>
      ) : null}

      <XStack flexWrap="wrap" gap="$4" justify="center">
        {lobbies.map((lobby) => (
          <PublicLobbyCard key={lobby.gameId} lobby={lobby} onJoin={() => onJoin(lobby.gameId)} />
        ))}
      </XStack>
    </PageShell>
  )
}
