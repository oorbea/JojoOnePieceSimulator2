import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import { secondsUntil } from '@/features/game/lib/match-rules'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  isRevealing: boolean
  onSkip: () => void
  tiebreak: boolean
  votingClosesAt: number | null
  gameState: string
}

// While the loadout reveal is playing: a "revealing" label plus a Skip
// button. Once revealed: the current game state label (voting/tiebreak),
// with an optional countdown once the server has told us when it closes.
export function VotingStatusBar({ isRevealing, onSkip, tiebreak, votingClosesAt, gameState }: Props) {
  const { t } = useTranslation()
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (votingClosesAt === null) return
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [votingClosesAt])

  if (isRevealing) {
    return (
      <GlassPanel tone="strong" rounded="$pill" px="$4" py="$2.5" width="100%">
        <XStack items="center" justify="space-between" gap="$2" flexWrap="wrap">
          <GlowText level="label">{t('game.match.revealing')}</GlowText>
          <GlossButton
            tone="glass"
            btnSize="sm"
            onPress={onSkip}
            accessibilityLabel={t('game.match.skipRevealA11y')}
            tooltip={t('game.match.skipRevealA11y')}
          >
            {t('game.match.skipReveal')}
          </GlossButton>
        </XStack>
      </GlassPanel>
    )
  }

  const seconds = secondsUntil(votingClosesAt, now)
  const label = tiebreak ? t('game.match.tiebreakOpen') : t(`enums.gameState.${gameState}`)

  return (
    <GlassPanel tone="strong" rounded="$pill" px="$4" py="$2.5" width="100%">
      <XStack items="center" justify="space-between" gap="$2" flexWrap="wrap">
        <GlowText level="label">{label}</GlowText>
        {seconds !== null ? (
          <GlowText level="label" tone="soft">
            {t('game.match.closesIn', { seconds })}
          </GlowText>
        ) : null}
      </XStack>
    </GlassPanel>
  )
}
