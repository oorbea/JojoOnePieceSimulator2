import { Lock, Unlock } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  locked: boolean
  isHost: boolean
  onToggle: () => void
}

export function LobbyLockRow({ locked, isHost, onToggle }: Props) {
  const { t } = useTranslation()

  if (!isHost) {
    return locked ? (
      <GlassPanel tone="plastic" px="$3" py="$2" rounded="$pill" elevate={0}>
        <XStack items="center" gap="$2">
          <Lock size={14} color="$panelTextSoft" />
          <GlowText level="label">{t('game.lock.locked')}</GlowText>
        </XStack>
      </GlassPanel>
    ) : null
  }

  return (
    <GlossButton
      tone="glass"
      btnSize="sm"
      onPress={onToggle}
      accessibilityLabel={t(locked ? 'game.lock.unlockAction' : 'game.lock.lockAction')}
    >
      <XStack items="center" gap="$2">
        {locked ? <Unlock size={14} color="$panelText" /> : <Lock size={14} color="$panelText" />}
        <GlowText level="label">{t(locked ? 'game.lock.unlockAction' : 'game.lock.lockAction')}</GlowText>
      </XStack>
    </GlossButton>
  )
}
