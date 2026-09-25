import { View } from 'react-native'

import { GlowText } from '@/shared/components/presentational/glow-text'
import { asToken } from '@/shared/lib/tamagui-token'

type Props = {
  children: string
  /** The verdict's own fill colour - a literal hex, not a theme token: this
   * is the manga cinematic's own fixed palette (morado/magenta/oro for
   * victory, rojo tinta/negro for defeat - see ObsidianVault/
   * game-victory-defeat-cinematic-2026-09-14.md), unrelated to light/dark
   * theme switching (the Modal it lives in already commits to one look). */
  fillColor: string
  strokeColor?: string
  align?: 'center' | 'left'
  /** A tamagui font-size token (e.g. '$11') - defaults to GlowText's own
   * 'display' level size. */
  fontSize?: string
}

// 8 offset copies in a ring - RN/react-native-web have no CSS text-stroke
// equivalent, so stacking duplicate glyphs behind the real fill is the
// standard mobile trick for a thick comic-book outline (Bangers reads as
// "manga" specifically BECAUSE of this stroke - without it, it's just a
// display font).
const STROKE_OFFSETS: [number, number][] = [
  [-2, -2],
  [2, -2],
  [-2, 2],
  [2, 2],
  [0, -2.5],
  [0, 2.5],
  [-2.5, 0],
  [2.5, 0],
]

// MangaVerdictText is the sorteo cinematics' own outlined verdict headline
// ("¡VICTORIA!" / "DERROTA" / "EL ESCUADRÓN HA SOBREVIVIDO") - Bangers
// (GlowText's 'display' level) with the stacked-outline trick above. The
// FIRST copy is invisible (opacity 0, normal flow) and exists only to size
// the wrapping View; every visible copy is absolutely positioned over it,
// so the whole stack occupies exactly one line's worth of space.
export function MangaVerdictText({
  children,
  fillColor,
  strokeColor = '#140b06',
  align = 'center',
  fontSize,
}: Props) {
  const size = fontSize ? asToken<'$11'>(fontSize) : undefined
  const fill = asToken<'$panelText'>(fillColor)
  const stroke = asToken<'$panelText'>(strokeColor)

  return (
    <View
      style={{ position: 'relative', alignItems: align === 'center' ? 'center' : 'flex-start' }}
    >
      <GlowText level="display" fontSize={size} align={align} opacity={0}>
        {children}
      </GlowText>
      {STROKE_OFFSETS.map(([dx, dy], i) => (
        <GlowText
          key={i}
          level="display"
          fontSize={size}
          align={align}
          color={stroke}
          position="absolute"
          l={dx}
          t={dy}
        >
          {children}
        </GlowText>
      ))}
      <GlowText
        level="display"
        fontSize={size}
        align={align}
        color={fill}
        position="absolute"
        l={0}
        t={0}
      >
        {children}
      </GlowText>
    </View>
  )
}
