import { Platform, useWindowDimensions, View } from 'react-native'

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
  /** A tamagui font-size token (e.g. '$11') - defaults to a responsive size
   * derived from `children`'s length (see `responsiveFontSize` below), so a
   * long title (a full team name, say) doesn't overflow/overlap on its
   * own. */
  fontSize?: string
}

// 8 offset copies in a ring - RN has no CSS text-stroke equivalent, so
// stacking duplicate glyphs behind the real fill is the standard mobile
// trick for a thick comic-book outline (Bangers reads as "manga"
// specifically BECAUSE of this stroke - without it, it's just a display
// font). Native only (see below) - web uses a single-paint CSS stroke
// instead, which has no wrapping-mismatch failure mode to begin with.
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

// Long titles (a full team name, a verbose localized verdict string) need a
// smaller display size than a short "DERROTA" - otherwise a single word can
// blow past the 90%-width cap below and wrap awkwardly. Thresholds are
// rough char counts, not measured text width (font metrics vary per
// locale/family) - generous enough to only kick in for genuinely long
// strings.
function responsiveFontSize(text: string): '$11' | '$9' | '$7' {
  if (text.length > 30) return '$7'
  if (text.length > 18) return '$9'
  return '$11'
}

// Caps the verdict block at 90% of the viewport so a long title never runs
// edge-to-edge or past it. Deliberately NOT a CSS/Yoga percentage: this
// component's immediate parent is an auto-sized (`flex: 0 0 auto` /
// shrink-to-fit) Animated.View one level down from the cinematic's
// full-screen root - a percentage `maxWidth` there resolves against that
// AUTO-sized ancestor, not the actual screen, which collapses circularly
// (verified live: a 7-char "VICTORY" measured ~131px wide instead of ~90%
// of a 1568px window, wrapping mid-word into "VICTOR"/"Y"). Both
// react-native-web's CSS flexbox and RN's own Yoga engine have this same
// percentage-on-undefined-size-parent ambiguity, so `useWindowDimensions`-
// derived pixel values are used instead on both platforms.
const MAX_WIDTH_FRACTION = 0.9

// MangaVerdictText is the sorteo cinematics' own outlined verdict headline
// ("¡VICTORIA!" / "DERROTA" / "EL ESCUADRÓN HA SOBREVIVIDO") - Bangers
// (GlowText's 'display' level) with a thick comic-book outline.
//
// Web: a single copy using CSS `-webkit-text-stroke` + `paint-order:
// stroke fill` - a real one-paint outline, not stacked glyphs. This
// sidesteps the web-only bug the stacked approach had: an absolutely
// positioned copy with only `left`/`top` set (no `right`) auto-sizes to its
// own content width instead of the sizer's wrapped width, so on a long
// title the copies could wrap differently than the invisible sizer and
// visibly overlap.
//
// Native (RN has no text-stroke equivalent): keeps the original
// stacked-copies trick, now with a matching `r` offset (`-dx`, so width
// stays `parentWidth - l - r` = `parentWidth` while the box still
// translates by `dx`) on every copy so each one wraps exactly like the
// sizer.
export function MangaVerdictText({
  children,
  fillColor,
  strokeColor = '#140b06',
  align = 'center',
  fontSize,
}: Props) {
  const size = asToken<'$11'>(fontSize ?? responsiveFontSize(children))
  const fill = asToken<'$panelText'>(fillColor)
  const stroke = asToken<'$panelText'>(strokeColor)
  const { width: windowWidth } = useWindowDimensions()
  const maxWidthPx = windowWidth * MAX_WIDTH_FRACTION

  if (Platform.OS === 'web') {
    return (
      <GlowText
        level="display"
        fontSize={size}
        align={align}
        color={fill}
        style={{
          // Web-only CSS (react-native-web's style types already allow
          // arbitrary CSS properties, so no `as any`/ts-expect-error is
          // needed) - passed through to the DOM untouched.
          WebkitTextStroke: `2px ${strokeColor}`,
          paintOrder: 'stroke fill',
          maxWidth: maxWidthPx,
        }}
      >
        {children}
      </GlowText>
    )
  }

  return (
    <View
      style={{
        position: 'relative',
        alignItems: align === 'center' ? 'center' : 'flex-start',
        maxWidth: maxWidthPx,
      }}
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
          r={-dx}
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
        r={0}
        t={0}
      >
        {children}
      </GlowText>
    </View>
  )
}
