import { YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { asToken } from '@/shared/lib/tamagui-token'

export type MeterBarTone = 'blue' | 'green' | 'red' | 'yellow'

const TONE_COLORS: Record<MeterBarTone, string> = {
  blue: '$wiiBlue',
  green: '$meadowGreen',
  red: '$strawHatRed',
  yellow: '$sunYellow',
}

type Props = {
  /** Clamped to [0, 1] internally - callers don't need to clamp first. */
  value: number
  tone?: MeterBarTone
  heightPx?: number
  /** Announced to assistive tech as a progressbar - pass a label describing
   * what's draining (e.g. "time left to vote"). */
  a11yLabel?: string
}

// A dumb determinate progress bar - no animation library, no layout
// measuring, just a width percentage on the fill. Shared rather than local
// to the vote bar because the reveal overlay's plain-text
// game.match.reveal.progress is the obvious second caller later. See
// dataviz's meter/progress guidance for the visual recipe this follows.
export function MeterBar({ value, tone = 'blue', heightPx = 8, a11yLabel }: Props) {
  const clamped = Math.max(0, Math.min(1, value))
  return (
    <YStack
      width="100%"
      height={heightPx}
      rounded="$pill"
      bg="$glassFillStrong"
      overflow="hidden"
      {...a11yProps(a11yLabel, 'progressbar', undefined)}
    >
      <YStack
        width={`${Math.round(clamped * 100)}%`}
        height="100%"
        rounded="$pill"
        bg={asToken(TONE_COLORS[tone])}
        transition="quick"
      />
    </YStack>
  )
}
