import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { LoadoutCard } from '@/features/game/components/presentational/match/loadout-card'
import { teamTone, teamToneColor } from '@/features/game/lib/lobby-rules'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlowText } from '@/shared/components/presentational/glow-text'
import type { GameSnapshot } from '@/features/game/types/game.types'

type Props = {
  snapshot: GameSnapshot
  selfId: string
}

// VERSUS: two tone-colored columns (reusing the lobby's team tone mapping).
// GAUNTLET: one wrapped row of cards. The caller's own loadout renders
// inline, highlighted by LoadoutCard's isSelf prop, rather than in a
// separate section. Only rendered once the sorteo overlay (RevealStage) has
// finished, so every card here already has its full loadout - no reveal
// state to thread through anymore.
export function MatchRoster({ snapshot, selfId }: Props) {
  const { t } = useTranslation()
  const mangas = snapshot.config.mangas

  if (snapshot.mode === 'VERSUS') {
    return (
      <YStack width="100%" gap="$3" $md={{ flexDirection: 'row' }}>
        {snapshot.teams.map((team, index) => {
          const tone = teamTone(index)
          const participants = snapshot.participants.filter((p) => p.teamId === team.id)
          return (
            <GlassPanel key={team.id} tone="strong" flex={1} p="$3" gap="$2.5" borderColor={teamToneColor(tone) as never}>
              <GlowText level="heading" color={teamToneColor(tone) as never}>
                {team.name}
              </GlowText>
              <XStack flexWrap="wrap" gap="$2.5">
                {participants.map((p) => (
                  <LoadoutCard key={p.id} participant={p} isSelf={p.id === selfId} mangas={mangas} />
                ))}
              </XStack>
            </GlassPanel>
          )
        })}
      </YStack>
    )
  }

  return (
    <GlassPanel tone="strong" width="100%" p="$3" gap="$2.5">
      <GlowText level="heading">{t('game.match.title')}</GlowText>
      <XStack flexWrap="wrap" gap="$2.5">
        {snapshot.participants.map((p) => (
          <LoadoutCard key={p.id} participant={p} isSelf={p.id === selfId} mangas={mangas} />
        ))}
      </XStack>
    </GlassPanel>
  )
}
