import { useTranslation } from 'react-i18next'

import { LoadoutCard } from '@/features/game/components/presentational/match/loadout-card'
import { ParticipantTile } from '@/features/game/components/presentational/match/participant-tile'
import type { GameParticipant } from '@/features/game/types/game.types'
import type { RovingItemProps } from '@/shared/hooks/use-roving-group'
import type { Manga } from '@/shared/lib/zod'
import { TooltipCard, useHoverTrigger } from '@/shared/components/presentational/tooltip'

const HOVER_CARD_DELAY_MS = 500

type Props = {
  participant: GameParticipant
  isSelf: boolean
  mangas: Manga[]
  onOpenModal: (participant: GameParticipant) => void
  /** Roving-tabindex props from MatchRoster's shared useRovingGroup - Tab
   * reaches one tile at a time, arrows move between them, Enter/Space opens
   * the modal (see use-roving-group.ts's doc on why it focuses via DOM id
   * rather than a ref: ParticipantTile's own triggerRef below is already
   * spoken for by the hover card). */
  itemProps: RovingItemProps
}

// One roster tile plus its own hover-card trigger - a per-participant hook
// instance, which is why this is its own component rather than inline in
// MatchRoster's .map (hooks can't live inside a loop body). Hover (web,
// after HOVER_CARD_DELAY_MS, 0.5s) or long-press (native) reveals the full
// LoadoutCard as a floating card; a tap/click opens the bigger modal
// instead - see useHoverTrigger's `nativeAutoHideMs: null` for why native
// dismisses the card on release rather than on a timer.
export function RosterParticipant({ participant, isSelf, mangas, onOpenModal, itemProps }: Props) {
  const { t } = useTranslation()
  const { visible, anchor, triggerRef, triggerProps } = useHoverTrigger({
    delayMs: HOVER_CARD_DELAY_MS,
    nativeAutoHideMs: null,
  })

  return (
    <>
      <ParticipantTile
        participant={participant}
        isSelf={isSelf}
        onPress={() => onOpenModal(participant)}
        triggerRef={triggerRef}
        triggerProps={triggerProps}
        itemProps={itemProps}
        viewA11yLabel={t('game.match.loadout.viewA11y', { name: participant.displayName })}
      />
      <TooltipCard visible={visible} anchor={anchor}>
        <LoadoutCard participant={participant} isSelf={isSelf} mangas={mangas} />
      </TooltipCard>
    </>
  )
}
