import { useEffect, useState } from 'react'
import { Animated } from 'react-native'
import { YStack } from 'tamagui'

import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  /** The line to show right now - changes on every narrator/land beat.
   * Empty string renders nothing (kept mounted so the fade doesn't pop). */
  line: string
  reducedMotion: boolean
}

// The sorteo's own narrator band, sitting above the roulette: V1's printed
// "before"/"after" lines (see i18n's game.match.reveal.narrator.* block),
// one at a time. Fades the text in on every change instead of an instant
// swap - purely decorative, so it collapses to an instant swap under
// useReducedMotion() rather than skip the update outright.
export function RevealNarrator({ line, reducedMotion }: Props) {
  const [opacity] = useState(() => new Animated.Value(1))

  useEffect(() => {
    if (reducedMotion) {
      opacity.setValue(1)
      return
    }
    opacity.setValue(0)
    Animated.timing(opacity, { toValue: 1, duration: 220, useNativeDriver: true }).start()
    // opacity is a stable ref (useState initializer), safe to omit.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [line, reducedMotion])

  return (
    <YStack width="100%" items="center" minH={28} justify="center">
      <Animated.View style={{ opacity }}>
        <GlowText level="heading" tone="soft">
          {line}
        </GlowText>
      </Animated.View>
    </YStack>
  )
}
