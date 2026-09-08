import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack, YStack, styled } from 'tamagui'

import { InsetRing, InsetShade, WiiCard } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { useReducedMotion } from '@/shared/hooks/use-reduced-motion'
import { subscribeShimmer } from '@/shared/lib/shimmer-ticker'
import { asToken } from '@/shared/lib/tamagui-token'

const SkeletonBase = styled(YStack, {
  name: 'SkeletonBase',
  bg: '$plasticEdge',
  rounded: '$card',
  overflow: 'hidden',
  position: 'relative',
})

// A wet-plastic tile, not a grey Material bar: the same InsetShade/InsetRing
// gloss vocabulary WiiCard uses, just dimmed and pulsing. Pulses via a
// shared module-level ticker (shimmer-ticker.ts), not Reanimated/CSS
// keyframes - see that file's header for why. Static under reduced motion.
export function Skeleton({
  width,
  height,
  rounded = '$card',
  label,
}: {
  width?: number | string
  height: number | string
  rounded?: string
  label?: string
}) {
  const { t } = useTranslation()
  const reducedMotion = useReducedMotion()
  const [lit, setLit] = useState(false)

  useEffect(() => {
    if (reducedMotion) return
    return subscribeShimmer(setLit)
  }, [reducedMotion])

  return (
    <SkeletonBase
      style={{ width, height }}
      rounded={asToken<'$card'>(rounded)}
      opacity={reducedMotion ? 0.85 : lit ? 1 : 0.7}
      transition="quick"
      {...a11yProps(label ?? t('common.imageLoading'), 'progressbar')}
    >
      <InsetRing rounded={asToken<'$card'>(rounded)} />
      <InsetShade />
    </SkeletonBase>
  )
}

// A real WiiCard shaped like stand-card.tsx/devil-fruit-card.tsx/stage-card.tsx
// (140px image well + name bar + a couple of stat bars), so the loading grid
// has pixel-identical geometry to the loaded one - zero layout shift on the
// skeleton-to-content relay that's the visible proof this feature works.
export function SkeletonCard({ width = 280 }: { width?: number }) {
  return (
    <WiiCard padded width={width} gap="$3">
      <Skeleton height={140} />
      <YStack gap="$3">
        <YStack gap="$1.5">
          <Skeleton height={16} width="60%" rounded="$pill" />
          <Skeleton height={12} width="35%" rounded="$pill" />
        </YStack>
        <XStack flexWrap="wrap" gap="$2">
          {[0, 1, 2].map((i) => (
            <YStack key={i} flexBasis={72} grow={1} minW={72} items="center" gap="$0.5">
              <Skeleton height={10} width="70%" rounded="$pill" />
              <Skeleton height={14} width="50%" rounded="$pill" />
            </YStack>
          ))}
        </XStack>
      </YStack>
    </WiiCard>
  )
}

export function SkeletonGrid({
  count = 8,
  cardWidth = 280,
}: {
  count?: number
  cardWidth?: number
}) {
  return (
    <XStack flexWrap="wrap" gap="$4" justify="center">
      {Array.from({ length: count }, (_, i) => (
        <SkeletonCard key={i} width={cardWidth} />
      ))}
    </XStack>
  )
}
