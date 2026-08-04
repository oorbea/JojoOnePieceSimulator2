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
    dock: {
      top: { position: 'absolute', t: 0, l: 0, r: 0 },
      bottom: { position: 'absolute', b: 0, l: 0, r: 0 },
      static: {},
    },
    density: {
      compact: { height: 52 },
      regular: {},
    },
  } as const,

  defaultVariants: {
    dock: 'static',
    density: 'regular',
  },
})

type ChannelBarProps = React.ComponentProps<typeof ChannelBarFrame>

// The floating Wii channel menu: a glass pill with a gloss cut across the
// top, real backdrop blur on web, alpha-compensated on native.
export function ChannelBar({ children, style, ...rest }: ChannelBarProps) {
  return (
    <ChannelBarFrame
      {...rest}
      bg={isWeb ? '$channelFill' : '$glassFillNative'}
      style={[WEB_BLUR_STYLE, style as object]}
    >
      <GlossOverlay coverage="half" shape="pill" />
      {children}
    </ChannelBarFrame>
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
