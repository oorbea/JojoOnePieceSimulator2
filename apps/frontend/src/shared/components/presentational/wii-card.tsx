import { LinearGradient } from '@tamagui/linear-gradient'
import { styled, YStack } from 'tamagui'

// RN has no CSS `inset` box-shadow, so the "recessed polished plastic" look
// is faked with two absolute children: a bright full-bleed ring (the inner
// highlight lip) and a bottom gradient to a soft black (the recessed
// shading). Exported separately so other primitives (e.g. circular avatar
// frames) can reuse just the inset look without the rest of WiiCard.
export const InsetRing = styled(YStack, {
  name: 'InsetRing',
  position: 'absolute',
  t: 0,
  l: 0,
  r: 0,
  b: 0,
  borderWidth: 2,
  borderColor: '$glossPeak',
  pointerEvents: 'none',
})

export const InsetShade = styled(LinearGradient, {
  name: 'InsetShade',
  position: 'absolute',
  l: 0,
  r: 0,
  b: 0,
  height: '38%',
  colors: ['$glossNil', 'rgba(0,0,0,0.07)'],
  start: [0, 0],
  end: [0, 1],
  pointerEvents: 'none',
})

export const WiiCard = styled(YStack, {
  name: 'WiiCard',
  position: 'relative',
  overflow: 'hidden',
  bg: '$plasticFill',
  rounded: '$card',
  borderWidth: 1.5,
  borderColor: '$glassEdge',
  shadowColor: '$softShadow',
  shadowOffset: { width: 0, height: 10 },
  shadowRadius: 22,
  shadowOpacity: 1,
  elevation: 6,

  variants: {
    interactive: {
      true: {
        transition: 'bouncy',
        cursor: 'pointer',
        hoverStyle: { scale: 1.05, y: -4 },
        pressStyle: { scale: 0.96, y: 2 },
      },
    },
    aspect: {
      square: { aspectRatio: 1 },
      channel: { aspectRatio: 16 / 9 },
      auto: {},
    },
    tone: {
      plastic: {},
      glass: { bg: '$glassFillStrong', borderColor: '$glassEdge' },
    },
    padded: {
      true: { p: '$4' },
    },
  } as const,

  defaultVariants: {
    aspect: 'auto',
    tone: 'plastic',
  },
})
