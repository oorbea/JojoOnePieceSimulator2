import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Image as RNImage,
  PanResponder,
  View,
  type GestureResponderEvent,
  type LayoutChangeEvent,
} from 'react-native'
import { XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { focalFromLocation, focalPosition, imageRectForWell } from '@/shared/lib/picture-source'
import { isWeb } from '@/shared/lib/web-blur'

import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'
import { LazyImage } from './lazy-image'
import { InsetRing } from './wii-card'

type Props = {
  uri: string | null
  x: number
  y: number
  onChange: (x: number, y: number) => void
}

const KEY_STEP = 0.01
const KEY_STEP_LARGE = 0.1

// Lets the admin pick which part of a picture stays visible when a card
// crops it to 'cover' - a crosshair over the full (uncropped) image on the
// left, dragged/tapped to a new point, with live mini-previews of the three
// real shapes the app actually crops this picture into (card, circular
// avatar, wide banner) so the effect is never a guess. Purely
// presentational: x/y/onChange are the form's own Controller value/
// onChange (or the focal-point-modal's local draft), same shape as every
// other field in these modals.
export function FocalPointPicker({ uri, x, y, onChange }: Props) {
  const { t } = useTranslation()
  const [wellSize, setWellSize] = useState({ width: 0, height: 0 })
  const [naturalSize, setNaturalSize] = useState<{ width: number; height: number } | null>(null)
  const [lastUri, setLastUri] = useState(uri)

  // A newly picked image has a different aspect ratio than whatever was
  // measured before it - drop the stale natural size immediately so a
  // render never divides by the previous image's proportions. Adjusted
  // directly during render off a "have we seen this uri yet" key (same
  // pattern lobby-room-container.tsx uses) rather than a useEffect, which
  // would call setState synchronously inside the effect body
  // (react-hooks/set-state-in-effect).
  if (uri !== lastUri) {
    setLastUri(uri)
    setNaturalSize(null)
  }

  const onLayout = (e: LayoutChangeEvent) => {
    const { width, height } = e.nativeEvent.layout
    setWellSize({ width, height })
  }

  // The image renders letterboxed inside the well unless the well happens
  // to share its aspect ratio exactly - a raw location/wellSize division
  // (the original bug) lands on a different point than the one the user
  // actually touched. imageRectForWell computes the image's own displayed
  // rect so the gesture and the crosshair both work in the image's
  // coordinate space, not the well's.
  const imageRect = imageRectForWell(naturalSize, wellSize)

  const handleTouch = (e: GestureResponderEvent) => {
    if (!imageRect.width || !imageRect.height) return
    const { x: nextX, y: nextY } = focalFromLocation(
      e.nativeEvent.locationX,
      e.nativeEvent.locationY,
      imageRect.width,
      imageRect.height
    )
    onChange(nextX, nextY)
  }

  // Rebuilt fresh each render (cheap - a plain object, no native binding
  // happens until a touch starts) rather than cached in a ref, so its
  // closures always see this render's `imageRect`/`onChange` directly - see
  // the `react-hooks/refs` note in `use-player-drag.ts` for why a ref-based
  // cache is deliberately avoided in this codebase.
  const panResponder = PanResponder.create({
    onStartShouldSetPanResponder: () => true,
    onMoveShouldSetPanResponder: () => true,
    // Once the drag has started, never let an ancestor ScrollView (the
    // form modal scrolls on mobile) steal it mid-gesture.
    onPanResponderTerminationRequest: () => false,
    onShouldBlockNativeResponder: () => true,
    onPanResponderGrant: handleTouch,
    onPanResponderMove: handleTouch,
  })

  const percentLabel = t('focalPoint.a11y', { x: Math.round(x * 100), y: Math.round(y * 100) })

  const onKeyDown = isWeb
    ? (e: { key: string; shiftKey?: boolean; preventDefault?: () => void }) => {
        const step = e.shiftKey ? KEY_STEP_LARGE : KEY_STEP
        switch (e.key) {
          case 'ArrowLeft':
            e.preventDefault?.()
            onChange(Math.max(0, x - step), y)
            return
          case 'ArrowRight':
            e.preventDefault?.()
            onChange(Math.min(1, x + step), y)
            return
          case 'ArrowUp':
            e.preventDefault?.()
            onChange(x, Math.max(0, y - step))
            return
          case 'ArrowDown':
            e.preventDefault?.()
            onChange(x, Math.min(1, y + step))
            return
          default:
            return
        }
      }
    : undefined

  return (
    <YStack gap="$2">
      <GlowText level="label">{t('focalPoint.title')}</GlowText>
      <GlowText level="label" tone="soft" fontSize="$3">
        {t('focalPoint.hint')}
      </GlowText>

      <YStack
        width="100%"
        height={220}
        rounded="$card"
        overflow="hidden"
        position="relative"
        bg="$plasticEdge"
        items="center"
        justify="center"
        onLayout={onLayout}
      >
        <InsetRing rounded="$card" />
        {uri ? (
          <View
            testID="focal-point-well"
            style={{ width: imageRect.width || '100%', height: imageRect.height || '100%' }}
            {...panResponder.panHandlers}
            {...(isWeb ? { tabIndex: 0, onKeyDown } : undefined)}
            {...a11yProps(percentLabel, 'adjustable')}
          >
            <RNImage
              key={uri ?? undefined}
              source={{ uri }}
              resizeMode="cover"
              style={{ width: '100%', height: '100%' }}
              onLoad={(e) => {
                const { width, height } = e.nativeEvent.source
                setNaturalSize({ width, height })
              }}
            />
            <YStack
              position="absolute"
              width={20}
              height={20}
              l={`${x * 100}%` as never}
              t={`${y * 100}%` as never}
              style={{
                transform: [{ translateX: -10 }, { translateY: -10 }],
                pointerEvents: 'none',
              }}
              rounded="$circle"
              borderWidth={2.5}
              borderColor="white"
              bg="rgba(59,130,246,0.55)"
              shadowColor="$softShadow"
              shadowRadius={4}
              shadowOpacity={1}
            />
          </View>
        ) : null}
      </YStack>

      <XStack gap="$3" flexWrap="wrap">
        <YStack flex={1} minW={100} gap="$1.5">
          <GlowText level="label" tone="soft" fontSize="$3">
            {t('focalPoint.previewCard')}
          </GlowText>
          <LazyImage uri={uri} height={100} contentPosition={focalPosition({ focalX: x, focalY: y })} />
        </YStack>

        <YStack flex={1} minW={100} gap="$1.5">
          <GlowText level="label" tone="soft" fontSize="$3">
            {t('focalPoint.previewAvatar')}
          </GlowText>
          <LazyImage
            uri={uri}
            height={100}
            rounded="$circle"
            aspectRatio={1}
            contentPosition={focalPosition({ focalX: x, focalY: y })}
          />
        </YStack>

        <YStack flex={1} minW={140} gap="$1.5">
          <GlowText level="label" tone="soft" fontSize="$3">
            {t('focalPoint.previewBanner')}
          </GlowText>
          <LazyImage
            uri={uri}
            aspectRatio={16 / 9}
            contentPosition={focalPosition({ focalX: x, focalY: y })}
          />
        </YStack>
      </XStack>

      <XStack>
        <GlossButton
          tone="glass"
          btnSize="sm"
          onPress={() => onChange(0.5, 0.5)}
          accessibilityLabel={t('focalPoint.reset')}
        >
          {t('focalPoint.reset')}
        </GlossButton>
      </XStack>
    </YStack>
  )
}
