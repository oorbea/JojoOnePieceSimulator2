import { LinearGradient } from '@tamagui/linear-gradient'
import { useMedia, YStack } from 'tamagui'

import { BubbleField } from './bubble-field'

// Bubble count is picked once per render from breakpoint, never re-randomized
// per frame — keeps the animated backdrop's cost flat regardless of screen
// size.
function useBubbleCount() {
  const media = useMedia()
  if (media.maxSm) return 7
  if (media.lg) return 16
  return 11
}

type AquaBackgroundProps = {
  /** Static gradient only, no animated bubbles — used by the error fallback
   * so a render error in the backdrop can never cause a retry loop. */
  plain?: boolean
}

// Sky gradient + two static low-alpha caustic blobs + the animated bubble
// field. Absolutely positioned, inset 0, non-interactive — sits behind
// every screen's content.
export function AquaBackground({ plain = false }: AquaBackgroundProps) {
  const count = useBubbleCount()

  return (
    <YStack
      position="absolute"
      t={0}
      l={0}
      r={0}
      b={0}
      z="$backdrop"
      style={{ pointerEvents: 'none' }}
      overflow="hidden"
    >
      <LinearGradient
        flex={1}
        colors={['$pageFrom', '$pageMid', '$pageTo']}
        start={[0, 0]}
        end={[0.2, 1]}
      />
      <YStack
        position="absolute"
        t="-10%"
        l="-15%"
        width="55%"
        height="55%"
        rounded="$circle"
        bg="$bubbleFill"
        opacity={0.35}
        rotate="12deg"
      />
      <YStack
        position="absolute"
        b="-15%"
        r="-10%"
        width="60%"
        height="60%"
        rounded="$circle"
        bg="$bubbleFill"
        opacity={0.35}
        rotate="-8deg"
      />
      {!plain ? <BubbleField count={count} /> : null}
    </YStack>
  )
}
