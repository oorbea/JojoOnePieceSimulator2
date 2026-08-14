import { Lock, Swords, Users } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { WiiCard } from '@/shared/components/presentational/wii-card'
import type { PublicLobby } from '@/features/game/types/game.types'

type Props = {
  lobby: PublicLobby
  onJoin: () => void
}

export function PublicLobbyCard({ lobby, onJoin }: Props) {
  const { t } = useTranslation()
  const full = lobby.playerCount >= lobby.maxPlayers

  return (
    <WiiCard width={300} padded gap="$3">
      <XStack items="center" gap="$2">
        <Swords size={18} color="$panelTextSoft" />
        <GlowText level="heading">{t(`enums.gameMode.${lobby.mode}`)}</GlowText>
        {lobby.locked ? <Lock size={14} color="$panelTextSoft" /> : null}
      </XStack>

      <GlowText level="label">{t('game.browse.hostedBy', { name: lobby.hostDisplayName })}</GlowText>

      <XStack items="center" gap="$1.5">
        <Users size={14} color="$panelTextSoft" />
        <GlowText level="label">
          {lobby.playerCount}/{lobby.maxPlayers}
        </GlowText>
      </XStack>

      <XStack flexWrap="wrap" gap="$1.5">
        {lobby.mangas.map((m) => (
          <GlowText key={m} level="label">
            {t(`enums.manga.${m}`)}
          </GlowText>
        ))}
      </XStack>

      <GlossButton
        tone="blue"
        btnSize="sm"
        disabled={full || lobby.locked}
        onPress={onJoin}
        accessibilityLabel={t('game.browse.join')}
      >
        {full ? t('game.lobby.teamFull') : t('game.browse.join')}
      </GlossButton>
    </WiiCard>
  )
}
