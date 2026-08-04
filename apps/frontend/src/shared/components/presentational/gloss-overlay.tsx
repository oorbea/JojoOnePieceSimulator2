import { LinearGradient } from '@tamagui/linear-gradient'
import { styled } from 'tamagui'

// The RN substitute for a CSS `::before` gloss highlight: React Native has
// no pseudo-elements, so this is a real absolutely-positioned sibling/child
// instead. The Tamagui compiler flattens this to static CSS on web / a
// static style object on native, so it costs nothing per frame — only
// mount it, never animate it. The parent MUST set `overflow="hidden"` or
// the gloss will bleed past rounded corners.
export const GlossOverlay = styled(LinearGradient, {
  name: 'GlossOverlay',
  position: 'absolute',
  t: 0,
  l: 0,
  r: 0,
  pointerEvents: 'none',
  z: '$gloss',
  colors: ['$glossPeak', '$glossFade', '$glossNil'],
  start: [0, 0],
  end: [0, 1],

  variants: {
    coverage: {
      half: { height: '50%' },
      third: { height: '34%' },
      full: { height: '100%' },
    },
    shape: {
      pill: { borderTopLeftRadius: '$pill', borderTopRightRadius: '$pill' },
      card: { borderTopLeftRadius: '$card', borderTopRightRadius: '$card' },
      circle: { borderRadius: '$circle' },
    },
  } as const,

  defaultVariants: {
    coverage: 'half',
    shape: 'card',
  },
})
