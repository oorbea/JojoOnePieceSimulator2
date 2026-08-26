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
  /** The server's own sorteo deadline (live.revealEndsAt) - still set for a
   * beat after the player has locally skipped the animation, since the
   * server doesn't actually open voting until this passes. */
  revealEndsAt: number | null
  gameState: string
}

// Three states: while RevealStage is showing (isRevealing, handled by the
// caller instead of here now - see MatchScreen), this bar isn't rendered at
// all. Once the player has locally skipped/finished but the server's own
// sorteo deadline hasn't passed yet, a "voting opens in Xs" countdown -
// nobody can vote early just because their own animation finished first.
// Once voting has genuinely opened: the current game state label, with the
// real countdown.
export function VotingStatusBar({ isRevealing, onSkip, tiebreak, votingClosesAt, revealEndsAt, gameState }: Props) {
  const { t } = useTranslation()
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (votingClosesAt === null && revealEndsAt === null) return
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [votingClosesAt, revealEndsAt])

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

  if (votingClosesAt === null && revealEndsAt !== null) {
    const seconds = secondsUntil(revealEndsAt, now)
    return (
      <GlassPanel tone="strong" rounded="$pill" px="$4" py="$2.5" width="100%">
        <GlowText level="label" tone="soft">
          {seconds !== null ? t('game.match.reveal.votingIn', { seconds }) : t('game.match.reveal.title')}
        </GlowText>
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
