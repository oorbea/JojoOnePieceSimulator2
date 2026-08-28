import { YStack } from 'tamagui'

import { ParticipantAvatar } from '@/features/game/components/presentational/match/participant-avatar'
import type { GameParticipant } from '@/features/game/types/game.types'
import { TooltipBubble, useTooltipTrigger } from '@/shared/components/presentational/tooltip'
import { a11yProps } from '@/shared/lib/a11y'

const AVATAR_SIZE = 32

type Props = {
  participant: GameParticipant
  isSelf: boolean
}

// One voter's avatar inside RoundResultPanel's per-option breakdown - just
// the picture (no username label, per the owner's "solo avatares" call,
// 2026-08-28), with the display name available as a tooltip (hover/focus
// web, long-press native) via GlossButton's same useTooltipTrigger/
// TooltipBubble pattern, since this isn't a GlossButton itself. Reuses
// ParticipantAvatar so a bot/initial/real picture render identically to the
// roster tile.
export function VoterAvatar({ participant, isSelf }: Props) {
  const { visible, anchor, triggerRef, triggerProps } = useTooltipTrigger(participant.displayName)

  return (
    <>
      <TooltipBubble visible={visible} label={participant.displayName} anchor={anchor} />
      <YStack
        ref={triggerRef as never}
        {...triggerProps}
        {...a11yProps(participant.displayName, 'image')}
      >
        <ParticipantAvatar participant={participant} size={AVATAR_SIZE} isSelf={isSelf} />
      </YStack>
    </>
  )
}
