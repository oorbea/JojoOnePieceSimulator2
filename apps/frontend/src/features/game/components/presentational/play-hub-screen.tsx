import { Compass, KeyRound, Swords } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { ChannelTile } from '@/shared/components/presentational/channel-tile'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import { SpeechBubble } from '@/shared/components/presentational/speech-bubble'

type Props = {
  onCreate: () => void
  onJoin: () => void
  onBrowse: () => void
}

export function PlayHubScreen({ onCreate, onJoin, onBrowse }: Props) {
  const { t } = useTranslation()

  return (
    <PageShell align="top" maxWidth={720}>
      <GlowText level="title">{t('game.hub.title')}</GlowText>
      <SpeechBubble tailSide="top">
        <GlowText level="label">{t('game.hub.subtitle')}</GlowText>
      </SpeechBubble>

      <XStack flexWrap="wrap" gap="$4" justify="center" pt="$4">
        <ChannelTile label={t('game.hub.create')} tone="green" icon={Swords} onPress={onCreate} />
        <ChannelTile label={t('game.hub.join')} tone="blue" icon={KeyRound} onPress={onJoin} />
        <ChannelTile label={t('game.hub.browse')} tone="grape" icon={Compass} onPress={onBrowse} />
      </XStack>
    </PageShell>
  )
}
