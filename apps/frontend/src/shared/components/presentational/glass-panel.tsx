import { Platform } from 'react-native'
import { styled, YStack } from 'tamagui'

import { isWeb, WEB_BLUR_STYLE } from '@/shared/lib/web-blur'

import { GlossOverlay } from './gloss-overlay'

const GlassPanelFrame = styled(YStack, {
  name: 'GlassPanel',
  overflow: 'hidden',
  position: 'relative',
  borderWidth: 1.5,
  borderColor: '$glassEdge',
  bg: isWeb ? '$glassFill' : '$glassFillNative',
  rounded: '$panel',
  shadowColor: '$softShadow',
  shadowOffset: { width: 0, height: 12 },
  shadowRadius: 30,
  shadowOpacity: 1,
  elevation: 8,

  variants: {
    tone: {
      glass: {},
      strong: { bg: isWeb ? '$glassFillStrong' : '$glassFillNative' },
      plastic: { bg: '$plasticFill', borderColor: '$plasticEdgeColor' },
      flat: { shadowRadius: 0, shadowOffset: { width: 0, height: 0 }, elevation: 0 },
    },
    elevate: {
      0: { shadowRadius: 0, shadowOffset: { width: 0, height: 0 }, elevation: 0 },
      1: { shadowRadius: 14, shadowOffset: { width: 0, height: 6 }, elevation: 4 },
      2: { shadowRadius: 30, shadowOffset: { width: 0, height: 12 }, elevation: 8 },
      3: { shadowRadius: 48, shadowOffset: { width: 0, height: 20 }, elevation: 14 },
    },
    radiusSize: {
      card: { rounded: '$card' },
      panel: { rounded: '$panel' },
      hero: { rounded: '$hero' },
      bubble: { rounded: '$bubble' },
    },
  } as const,

  defaultVariants: {
    tone: 'glass',
  },
})

type GlassPanelProps = React.ComponentProps<typeof GlassPanelFrame> & {
  /** Adds the iOS-6 top-half gloss highlight. */
  glossy?: boolean
}

// Aero panel: translucent fill, bright thin border, soft big shadow. Real
// `backdrop-filter` blur on web; native compensates with a higher opacity
// fill (no `expo-blur` dependency added for this pass).
export function GlassPanel({ glossy, children, style, ...rest }: GlassPanelProps) {
  return (
    <GlassPanelFrame {...rest} style={[WEB_BLUR_STYLE, style as object]}>
      {glossy ? <GlossOverlay coverage="half" shape="card" /> : null}
      <YStack z="$content" flex={1}>
        {children}
      </YStack>
    </GlassPanelFrame>
  )
}

// Re-exported so screens can guard platform-only visual tweaks without a
// second Platform import.
export const isWebPlatform = Platform.OS === 'web'
