import { Image as ExpoImage, type ImageContentFit } from 'expo-image'
import { useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import type { View } from 'react-native'
import { YStack } from 'tamagui'

import { GlowText } from '@/shared/components/presentational/glow-text'
import { Skeleton } from '@/shared/components/presentational/skeleton'
import { InsetRing, InsetShade } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { useImageSlot } from '@/shared/hooks/use-image-slot'
import { useInViewport } from '@/shared/hooks/use-in-viewport'
import type { QueueLane } from '@/shared/lib/image-queue'
import { asToken } from '@/shared/lib/tamagui-token'
import type { PictureStatus } from '@/shared/contracts/enums'

export type LazyImageState = 'queued' | 'granted' | 'loaded' | 'error'

type Props = {
  uri: string | null
  lqip?: string | null
  // Required unless aspectRatio is given instead (e.g. a full-width 16:9
  // hero banner that has no fixed pixel height) - Tamagui's own aspectRatio
  // sizing takes over when height is omitted.
  height?: number | string
  aspectRatio?: number
  contentFit?: ImageContentFit
  rounded?: string
  fallback?: React.ReactNode
  // Position within the grid, used for priority (lower loads first).
  order?: number
  priorityHint?: QueueLane
  recyclingKey?: string
  pictureStatus?: PictureStatus
  retryToken?: number
  onStateChange?: (state: LazyImageState) => void
}

// Owns the entire image well - sizing, InsetRing/InsetShade chrome,
// skeleton, LQIP crossfade, and the fallback icon - so adopting it means
// deleting the ~15 duplicated lines each catalogue card currently has, not
// wrapping them (see the plan: stand-card.tsx/devil-fruit-card.tsx/
// stage-card.tsx/power-block.tsx/stage-banner.tsx/loadout-card.tsx all
// carry a copy of this ternary today).
//
// Purely presentational: no query/store/router hooks, only the UI-mechanics
// hooks (useImageSlot/useInViewport), same category as use-roving-group.ts.
// Retry is reported via onStateChange, never rendered as an interactive
// element in here - nesting a button inside the card's own Pressable would
// create a second, invalid tab stop (norma-teclado.md). The caller's
// existing Pressable should redirect its onPress to retry when state is
// 'error'.
export function LazyImage({
  uri,
  lqip,
  height,
  aspectRatio,
  contentFit = 'cover',
  rounded = '$card',
  fallback,
  order = 0,
  priorityHint = 'grid',
  recyclingKey,
  pictureStatus,
  retryToken,
  onStateChange,
}: Props) {
  const { t } = useTranslation()
  const wellRef = useRef<View>(null)
  const visibility = useInViewport(wellRef)
  const { state, onLoad, onError } = useImageSlot({
    uri,
    visibility,
    order,
    lane: priorityHint,
    retryToken,
  })

  useEffect(() => {
    onStateChange?.(state)
  }, [state, onStateChange])

  const generating = pictureStatus === 'PENDING'

  return (
    <YStack
      ref={wellRef}
      width="100%"
      style={height !== undefined ? { height } : undefined}
      aspectRatio={aspectRatio}
      rounded={asToken<'$card'>(rounded)}
      overflow="hidden"
      position="relative"
      bg="$plasticEdge"
    >
      <InsetRing rounded={asToken<'$card'>(rounded)} />
      {state === 'granted' || state === 'loaded' ? (
        <ExpoImage
          source={{ uri: uri ?? undefined }}
          placeholder={lqip ? { uri: lqip } : undefined}
          style={{ width: '100%', height: '100%' }}
          contentFit={contentFit}
          transition={200}
          recyclingKey={recyclingKey}
          onLoad={onLoad}
          onError={onError}
          {...a11yProps(t('common.imageLoading'))}
        />
      ) : state === 'error' ? (
        <YStack
          flex={1}
          items="center"
          justify="center"
          gap="$1"
          {...a11yProps(t('common.imageUnavailable'))}
        >
          {fallback}
          {generating ? (
            <GlowText level="label" tone="soft">
              {t('common.imageGenerating')}
            </GlowText>
          ) : null}
        </YStack>
      ) : (
        <Skeleton
          height="100%"
          rounded={asToken<'$card'>(rounded)}
          label={t('common.imageLoading')}
        />
      )}
      <InsetShade />
    </YStack>
  )
}
