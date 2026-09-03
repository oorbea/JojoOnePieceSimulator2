import { styled, XStack, YStack } from 'tamagui'

import { isWeb, WEB_BLUR_STYLE } from '@/shared/lib/web-blur'

import { GlossOverlay } from './gloss-overlay'
import { TooltipBubble, useTooltipTrigger } from './tooltip'

const ChannelBarFrame = styled(XStack, {
  name: 'ChannelBar',
  bg: '$channelFill',
  rounded: '$pill',
  minH: 64,
  items: 'center',
  justify: 'center',
  // Content that doesn't fit one row makes the pill grow instead of getting
  // clipped/overlapping — see channel-bar.tsx's ChannelBar doc comment.
  // AppShell measures the real rendered height via onLayout, so growth here
  // never covers page content underneath.
  flexWrap: 'wrap',
  gap: '$2',
  px: '$3',
  py: '$2',
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
      compact: { minH: 52 },
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

const ChannelBarItemFrame = styled(YStack, {
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
    // `active` no longer paints a bg here — the sliding
    // ChannelBarIndicator owns that pill now, positioned from each item's
    // own measured layout (see AppShell). Kept as a no-op variant so
    // callers can keep passing `active` for typing/clarity without it
    // doing anything visually by itself.
    active: {
      true: {},
    },
    iconOnly: {
      true: { aspectRatio: 1, px: 0 },
    },
  } as const,
})

type ChannelBarItemProps = React.ComponentProps<typeof ChannelBarItemFrame> & {
  tooltip?: string
}

export function ChannelBarItem({ tooltip, children, ...rest }: ChannelBarItemProps) {
  const { visible, anchor, triggerRef, triggerProps } = useTooltipTrigger(tooltip)
  return (
    <>
      <ChannelBarItemFrame ref={triggerRef as never} {...rest} {...triggerProps}>
        {children}
      </ChannelBarItemFrame>
      <TooltipBubble visible={visible} label={tooltip} anchor={anchor} />
    </>
  )
}
