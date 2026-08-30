import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { LoadoutCard } from '@/features/game/components/presentational/match/loadout-card'
import { teamTone, teamToneColor } from '@/features/game/lib/lobby-rules'
import { secondsUntil } from '@/features/game/lib/match-rules'
import type { GameSnapshot } from '@/features/game/types/game.types'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { useNow } from '@/shared/hooks/use-now'

type Props = {
  snapshot: GameSnapshot
  selfId: string
  summaryEndsAt: number | null
  /** SUMMARY_READY_CHANGED's own aggregate - how many of how many connected
   * humans have already marked themselves ready to skip. null before the
   * first frame for this SUMMARY window arrives. */
  readyCount: number | null
  readyTotal: number | null
  onSkip: () => void
}

// The loadout-summary screen (owner request, 2026-08-30): once a reassigning
// round's sorteo finishes (or is skipped), every participant's freshly
// assigned loadout gets one more beat on screen - full LoadoutCard tiles,
// grouped by team in VERSUS same as MatchRoster, one row in GAUNTLET -
// before voting opens. Synchronized skip mirrors the sorteo's own
// MarkRevealReady exactly, just scoped to the SUMMARY window.
export function LoadoutSummaryStage({ snapshot, selfId, summaryEndsAt, readyCount, readyTotal, onSkip }: Props) {
  const { t } = useTranslation()
  const mangas = snapshot.config.powerMangas
  const now = useNow(1000, summaryEndsAt !== null)
  const seconds = secondsUntil(summaryEndsAt, now)

  const grid =
    snapshot.mode === 'VERSUS' ? (
      <YStack width="100%" gap="$3" $md={{ flexDirection: 'row' }}>
        {snapshot.teams.map((team, index) => {
          const tone = teamTone(index)
          const participants = snapshot.participants.filter((p) => p.teamId === team.id)
          return (
            <GlassPanel
              key={team.id}
              tone="strong"
              flex={1}
              p="$3"
              gap="$2.5"
              borderColor={teamToneColor(tone) as never}
            >
              <GlowText level="heading" color={teamToneColor(tone) as never}>
                {team.name}
              </GlowText>
              <XStack flexWrap="wrap" gap="$2.5" justify="center">
                {participants.map((p) => (
                  <LoadoutCard key={p.id} participant={p} isSelf={p.id === selfId} mangas={mangas} />
                ))}
              </XStack>
            </GlassPanel>
          )
        })}
      </YStack>
    ) : (
      <XStack flexWrap="wrap" gap="$2.5" justify="center" width="100%">
        {snapshot.participants.map((p) => (
          <LoadoutCard key={p.id} participant={p} isSelf={p.id === selfId} mangas={mangas} />
        ))}
      </XStack>
    )

  return (
    <GlassPanel tone="strong" width="100%" p="$4" gap="$3" items="center">
      <GlowText level="heading">{t('game.match.summary.title')}</GlowText>
      {seconds !== null ? (
        <GlowText level="label" tone="soft">
          {t('game.match.summary.votingIn', { seconds })}
        </GlowText>
      ) : null}

      {grid}

      <GlossButton
        tone="glass"
        btnSize="sm"
        onPress={onSkip}
        accessibilityLabel={t('game.match.summary.skipA11y')}
        tooltip={t('game.match.summary.skipA11y')}
      >
        {readyTotal
          ? t('game.match.summary.readyCount', { ready: readyCount ?? 0, total: readyTotal })
          : t('game.match.summary.skip')}
      </GlossButton>
    </GlassPanel>
  )
}
