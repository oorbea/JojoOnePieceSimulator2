import { LinearGradient } from '@tamagui/linear-gradient'
import { styled, YStack } from 'tamagui'

import { isWeb } from '@/shared/lib/web-blur'

// Neither @tamagui/linear-gradient nor expo-linear-gradient supports radial
// gradients. Faked with 5 concentric circles whose alphas approximate a
// Gaussian falloff — cheap, works identically on every platform. On web we
// additionally apply a CSS blur to the ring stack, which collapses the
// rings into a genuinely smooth radial glow.
const RING_STOPS = [
  { fraction: 1.0, opacity: 0.05 },
  { fraction: 0.78, opacity: 0.09 },
  { fraction: 0.58, opacity: 0.14 },
  { fraction: 0.4, opacity: 0.22 },
  { fraction: 0.24, opacity: 0.34 },
]

const Ring = styled(YStack, {
  name: 'LensFlareRing',
  position: 'absolute',
  rounded: '$circle',
  bg: '$glowColor',
})

const Streak = styled(LinearGradient, {
  name: 'LensFlareStreak',
  position: 'absolute',
  height: 3,
  width: '140%',
  colors: ['$glossNil', '$glossPeak', '$glossNil'],
  start: [0, 0],
  end: [1, 0],
  opacity: 0.55,
})

const SIZE_PX = { sm: 120, md: 220, lg: 360 } as const

type LensFlareProps = {
  size?: keyof typeof SIZE_PX
  streak?: boolean
}

// A static (never per-frame-animated) glow + optional anamorphic streak,
// meant to sit BEHIND a single primary CTA per screen — never on every
// button, per the "spend boldness in one place" rule. Always
// pointerEvents="none" so it never intercepts touches.
export function LensFlare({ size = 'md', streak = true }: LensFlareProps) {
  const px = SIZE_PX[size]

  return (
    <YStack
      position="absolute"
      t="50%"
      l="50%"
      width={px}
      height={px}
      ml={-px / 2}
      mt={-px / 2}
      items="center"
      justify="center"
      z="$backdrop"
      pointerEvents="none"
      style={isWeb ? ({ filter: 'blur(28px)' } as React.CSSProperties as object) : undefined}
    >
      {RING_STOPS.map((ring) => (
        <Ring
          key={ring.fraction}
          width={px * ring.fraction}
          height={px * ring.fraction}
          opacity={ring.opacity}
        />
      ))}
      {streak ? <Streak rotate="-18deg" /> : null}
    </YStack>
  )
}
