import { styled, XStack, YStack } from 'tamagui'

import { isWeb, WEB_BLUR_STYLE } from '@/shared/lib/web-blur'

import { GlossOverlay } from './gloss-overlay'

const ChannelBarFrame = styled(XStack, {
  name: 'ChannelBar',
  bg: '$channelFill',
  rounded: '$pill',
  height: 64,
  items: 'center',
  gap: '$2',
  px: '$3',
  borderWidth: 1.5,
  borderColor: '$glassEdge',
  shadowColor: '$hardShadow',
  shadowRadius: 28,
  shadowOffset: { width: 0, height: 10 },
  shadowOpacity: 1,
  elevation: 10,
  overflow: 'hidden',
  position: 'relative',
  z: '$nav',
  self: 'center',
  maxW: 1080,
  width: '100%',

  variants: {
    density: {
      compact: { height: 52 },
      regular: {},
    },
  } as const,

  defaultVariants: {
    density: 'regular',
  },
})

type ChannelBarDock = 'top' | 'bottom' | 'static'

type ChannelBarProps = React.ComponentProps<typeof ChannelBarFrame> & {
  dock?: ChannelBarDock
}

// The floating Wii channel menu: a glass pill with a gloss cut across the
// top, real backdrop blur on web, alpha-compensated on native.
//
// `dock="top"|"bottom"` used to put `position:absolute; left:0; right:0`
// directly on the pill, which over-constrains it against the pill's own
// `maxW:1080; self:'center'` — an absolutely positioned box with both `left`
// and `right` pinned ignores `alignSelf`, so past 1080px wide the bar hugged
// the left edge instead of centering (and pushed the logout button to the
// 1080px mark rather than the right edge). Docking now wraps the pill — kept
// in normal flow, so `self:'center'` finally works — in a thin full-bleed
// absolute host whose only job is centering; the host itself never captures
// touches, so it can't intercept clicks outside the pill's own bounds.
export function ChannelBar({ children, style, dock = 'static', t, b, ...rest }: ChannelBarProps) {
  const frame = (
    <ChannelBarFrame
      {...rest}
      bg={isWeb ? '$channelFill' : '$glassFillNative'}
      style={[WEB_BLUR_STYLE, style as object]}
    >
      <GlossOverlay coverage="half" shape="pill" />
      {children}
    </ChannelBarFrame>
  )

  if (dock === 'static') return frame

  return (
    <YStack
      position="absolute"
      t={dock === 'top' ? t : undefined}
      b={dock === 'bottom' ? b : undefined}
      l={0}
      r={0}
      items="center"
      style={{ pointerEvents: 'box-none' }}
    >
      {frame}
    </YStack>
  )
}

export const ChannelBarItem = styled(YStack, {
  name: 'ChannelBarItem',
  height: 48,
  minW: 48,
  rounded: '$pill',
  px: '$3',
  items: 'center',
  justify: 'center',
  transition: 'bouncy',
  cursor: 'pointer',
  hoverStyle: { scale: 1.06 },
  pressStyle: { scale: 0.93 },
  z: '$content',

  variants: {
    active: {
      true: { bg: '$channelActive' },
    },
    iconOnly: {
      true: { aspectRatio: 1, px: 0 },
    },
  } as const,
})
