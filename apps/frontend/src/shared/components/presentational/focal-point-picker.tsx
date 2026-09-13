import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Image as RNImage, PanResponder, type LayoutChangeEvent } from 'react-native'
import { XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { focalPosition } from '@/shared/lib/picture-source'

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

function clamp01(v: number): number {
  return Math.min(1, Math.max(0, v))
}

// Lets the admin pick which part of a picture stays visible when a card
// crops it to 'cover' - a crosshair over the full (uncropped) image on the
// left, dragged/tapped to a new point, with a live mini-preview of the
// actual card crop on the right so the effect is never a guess. Purely
// presentational: x/y/onChange are the form's own Controller value/
// onChange, same shape as every other field in these modals.
export function FocalPointPicker({ uri, x, y, onChange }: Props) {
  const { t } = useTranslation()
  const [wellSize, setWellSize] = useState({ width: 0, height: 0 })

  const onLayout = (e: LayoutChangeEvent) => {
    const { width, height } = e.nativeEvent.layout
    setWellSize({ width, height })
  }

  // One PanResponder handles both a tap and a drag - RN's touch model gives
  // locationX/locationY relative to the responder view for both the initial
  // grant and every subsequent move, so a single handler covers "click to
  // set" and "drag to adjust" without two separate gesture recognizers.
  // Rebuilt fresh each render (cheap - a plain object, no native binding
  // happens until a touch starts) rather than cached in a ref, so its
  // closures always see this render's `wellSize`/`onChange` directly.
  const panResponder = PanResponder.create({
    onStartShouldSetPanResponder: () => true,
    onMoveShouldSetPanResponder: () => true,
    onPanResponderGrant: (e) => {
      const { width, height } = wellSize
      if (!width || !height) return
      onChange(clamp01(e.nativeEvent.locationX / width), clamp01(e.nativeEvent.locationY / height))
    },
    onPanResponderMove: (e) => {
      const { width, height } = wellSize
      if (!width || !height) return
      onChange(clamp01(e.nativeEvent.locationX / width), clamp01(e.nativeEvent.locationY / height))
    },
  })

  const percentLabel = t('focalPoint.a11y', { x: Math.round(x * 100), y: Math.round(y * 100) })

  return (
    <YStack gap="$2">
      <GlowText level="label">{t('focalPoint.title')}</GlowText>
      <GlowText level="label" tone="soft" fontSize="$3">
        {t('focalPoint.hint')}
      </GlowText>
      <XStack gap="$3" flexWrap="wrap">
        <YStack
          flex={1}
          minW={160}
          height={160}
          rounded="$card"
          overflow="hidden"
          position="relative"
          bg="$plasticEdge"
          onLayout={onLayout}
          {...panResponder.panHandlers}
          {...a11yProps(percentLabel, 'adjustable')}
        >
          <InsetRing rounded="$card" />
          {uri ? (
            <RNImage
              source={{ uri }}
              resizeMode="contain"
              style={{ width: '100%', height: '100%' }}
            />
          ) : null}
          <YStack
            position="absolute"
            width={20}
            height={20}
            l={`${x * 100}%` as never}
            t={`${y * 100}%` as never}
            style={{ transform: [{ translateX: -10 }, { translateY: -10 }] }}
            rounded="$circle"
            borderWidth={2.5}
            borderColor="white"
            bg="rgba(59,130,246,0.55)"
            shadowColor="$softShadow"
            shadowRadius={4}
            shadowOpacity={1}
          />
        </YStack>

        <YStack flex={1} minW={160} gap="$1.5">
          <GlowText level="label" tone="soft" fontSize="$3">
            {t('focalPoint.previewLabel')}
          </GlowText>
          <LazyImage
            uri={uri}
            height={140}
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
