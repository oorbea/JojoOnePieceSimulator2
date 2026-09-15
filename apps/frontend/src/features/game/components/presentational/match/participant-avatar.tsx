import { Bot } from '@tamagui/lucide-icons-2'
import { YStack } from 'tamagui'

import { teamTone, teamToneColor } from '@/features/game/lib/lobby-rules'
import type { GameParticipant } from '@/features/game/types/game.types'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { LazyImage } from '@/shared/components/presentational/lazy-image'
import { focalPosition } from '@/shared/lib/picture-source'

// Deterministic (never random) tone for a participant with no picture at
// all - the same id always gets the same colour, reusing the four tones
// TeamColumn/MatchRoster already use rather than inventing a fifth palette.
function toneFor(id: string): string {
  let sum = 0
  for (let i = 0; i < id.length; i++) sum += id.charCodeAt(i)
  return teamToneColor(teamTone(sum))
}

function initialFor(displayName: string): string {
  return displayName.trim().charAt(0).toUpperCase() || '?'
}

type Props = {
  participant: GameParticipant
  size: number
  isSelf?: boolean
}

// A participant's picture: their own avatar/Google picture if
// `avatarThumb` resolved to one, a robot icon for a bot, or - the common
// case, most players never upload an avatar - an initial in a colour
// circle derived from their id. Shared by ParticipantTile (the roster) and
// LoadoutModal (the breakdown header) so the two never drift.
//
// A seat that has dropped out (see GameService's grace timers,
// ObsidianVault) reads visually here too, without any extra prop - both
// `connected`/`abandoned` already ride on `participant` itself:
//   - disconnected (still within the grace window): the whole avatar dims
//     and desaturates - RN has no CSS filter, so a semi-opaque grey scrim
//     over the picture does the job instead of a true grayscale filter.
//   - abandoned (grace elapsed, the seat now auto-plays/stays only for its
//     loadout - see IGameMode.AutoVotesForAbandoned): same dimmed look plus
//     a small bot badge, the same visual cue a real Bot participant already
//     gets, since the seat is now effectively bot-controlled.
export function ParticipantAvatar({ participant, size, isSelf = false }: Props) {
  const isBot = participant.kind === 'BOT'
  const isAbandoned = participant.abandoned
  const isInactive = !participant.connected || isAbandoned
  const badgeSize = Math.max(14, size * 0.4)

  return (
    <YStack
      width={size}
      height={size}
      rounded="$circle"
      overflow="visible"
      opacity={isInactive ? 0.55 : 1}
    >
      <YStack
        width={size}
        height={size}
        rounded="$circle"
        overflow="hidden"
        borderWidth={isSelf ? 2.5 : 1.5}
        borderColor={isSelf ? ('$wiiBlue' as never) : '$glassEdge'}
      >
        <LazyImage
          uri={participant.avatarThumb || null}
          height={size}
          rounded="$circle"
          contentPosition={focalPosition({
            focalX: participant.avatarFocalX,
            focalY: participant.avatarFocalY,
          })}
          fallback={
            <YStack
              flex={1}
              width="100%"
              items="center"
              justify="center"
              bg={(isBot ? '$plasticEdge' : toneFor(participant.id)) as never}
            >
              {isBot ? (
                <Bot size={size * 0.46} color="$panelTextSoft" />
              ) : (
                <GlowText level="heading" tone="onColor" fontSize={size * 0.42}>
                  {initialFor(participant.displayName)}
                </GlowText>
              )}
            </YStack>
          }
        />
        {isInactive ? (
          <YStack position="absolute" inset={0} bg="rgba(90,100,115,0.45)" />
        ) : null}
      </YStack>
      {isAbandoned ? (
        <YStack
          position="absolute"
          b={-2}
          r={-2}
          width={badgeSize}
          height={badgeSize}
          rounded="$circle"
          items="center"
          justify="center"
          bg="$tangerine"
          borderWidth={1.5}
          borderColor="$glassEdge"
        >
          <Bot size={badgeSize * 0.62} color="$plasticWhite" />
        </YStack>
      ) : null}
    </YStack>
  )
}
