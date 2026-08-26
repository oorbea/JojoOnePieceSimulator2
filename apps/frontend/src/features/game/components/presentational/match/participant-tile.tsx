import { YStack } from 'tamagui'

import { ParticipantAvatar } from '@/features/game/components/presentational/match/participant-avatar'
import type { GameParticipant } from '@/features/game/types/game.types'
import type { RovingItemProps } from '@/shared/hooks/use-roving-group'
import { a11yProps } from '@/shared/lib/a11y'
import { GlowText } from '@/shared/components/presentational/glow-text'

const AVATAR_SIZE = 56

type Props = {
  participant: GameParticipant
  isSelf: boolean
  onPress: () => void
  /** Long-press (native) / hover (web, after the caller's delay) trigger
   * props and ref from useHoverTrigger - ParticipantTile only renders the
   * avatar+name, the hover card itself is the caller's job (MatchRoster),
   * since the card's content (a full LoadoutCard) needs `mangas`, which
   * this tile has no reason to know about. */
  triggerRef: React.Ref<unknown>
  triggerProps: Record<string, unknown>
  /** Roving-tabindex props (see roster-participant.tsx's doc). Merged with
   * triggerProps below rather than spread separately, since both define
   * onFocus (useHoverTrigger's opens the hover card, the roving group's
   * tracks which tile is active) and both must fire. */
  itemProps: RovingItemProps
  viewA11yLabel: string
}

// The voting-round roster's tile: avatar + username, nothing else - the
// full power breakdown lives one interaction away (hover card / tap
// modal), not on the roster itself. See match-roster.tsx for why: showing
// every stat for every participant at once was too much information at a
// glance once the roster is what you actually look at during voting.
export function ParticipantTile({
  participant,
  isSelf,
  onPress,
  triggerRef,
  triggerProps,
  itemProps,
  viewA11yLabel,
}: Props) {
  return (
    <YStack
      ref={triggerRef as never}
      {...triggerProps}
      onFocus={() => {
        ;(triggerProps.onFocus as (() => void) | undefined)?.()
        itemProps.onFocus?.()
      }}
      // Enter/Space is handled entirely by itemProps.onKeyDown (it calls
      // MatchRoster's onActivate, which opens the modal for this same
      // participant via useRovingGroup's index) - not duplicated here.
      onKeyDown={itemProps.onKeyDown}
      tabIndex={itemProps.tabIndex}
      id={itemProps.id}
      onPress={onPress}
      items="center"
      gap="$1.5"
      width={88}
      py="$1.5"
      rounded="$card"
      cursor="pointer"
      transition="bouncy"
      hoverStyle={{ scale: 1.05, y: -2 }}
      pressStyle={{ scale: 0.94 }}
      focusStyle={{ outlineColor: '$channelActive', outlineWidth: 3, outlineStyle: 'solid' }}
      {...a11yProps(viewA11yLabel, 'button')}
    >
      <ParticipantAvatar participant={participant} size={AVATAR_SIZE} isSelf={isSelf} />
      <GlowText level="label" fontSize="$1" numberOfLines={1}>
        {participant.displayName}
      </GlowText>
    </YStack>
  )
}
