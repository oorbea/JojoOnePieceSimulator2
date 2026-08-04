import { SizableText, styled } from 'tamagui'

// Rounded, chunky Fredoka headings with a soft text-shadow — reads clearly
// over glass/gloss without needing a solid text plate behind it.
export const GlowText = styled(SizableText, {
  name: 'GlowText',
  fontFamily: '$heading',
  fontWeight: '800',
  letterSpacing: -0.3,
  color: '$panelText',
  textShadowColor: '$textGlow',
  textShadowOffset: { width: 0, height: 1 },
  textShadowRadius: 3,

  variants: {
    level: {
      hero: { fontSize: '$11', fontFamily: '$heading' },
      title: { fontSize: '$9', fontFamily: '$heading' },
      heading: { fontSize: '$7', fontFamily: '$heading' },
      label: {
        fontSize: '$4',
        fontFamily: '$body',
        fontWeight: '700',
        letterSpacing: 0,
        textShadowRadius: 0,
        color: '$panelTextSoft',
      },
    },
    tone: {
      ink: { color: '$panelText' },
      onColor: { color: 'white' },
      soft: { color: '$panelTextSoft' },
    },
    align: {
      center: { text: 'center' },
      left: { text: 'left' },
    },
  } as const,

  defaultVariants: {
    level: 'heading',
  },
})
