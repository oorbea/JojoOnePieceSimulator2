import Svg, { RadialGradient, Rect, Stop } from 'react-native-svg'

const FILL_STYLE = { position: 'absolute' as const, top: 0, left: 0, right: 0, bottom: 0 }

export type GlowStop = { offset: string; color: string; opacity?: number }

// A static full-bleed radial gradient background (RN has no CSS
// radial-gradient, unlike @tamagui/linear-gradient's expo-linear-gradient
// wrapper, which is linear-only) - the manga cinematics' own base
// background (victory: morado→magenta→oro; defeat: rojo tinta→negro, see
// ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md). Static
// content; the caller animates a wrapping Animated.View's opacity if it
// needs to fade this in, per the project's transform/opacity-only motion
// norm - this component itself never animates.
export function RadialGlow({
  stops,
  cx = '50%',
  cy = '45%',
  r = '75%',
}: {
  stops: GlowStop[]
  cx?: string
  cy?: string
  r?: string
}) {
  return (
    <Svg style={[FILL_STYLE]} pointerEvents="none">
      <RadialGradient id="glow" cx={cx} cy={cy} r={r} gradientUnits="objectBoundingBox">
        {stops.map((s, i) => (
          <Stop key={i} offset={s.offset} stopColor={s.color} stopOpacity={s.opacity ?? 1} />
        ))}
      </RadialGradient>
      <Rect x="0" y="0" width="100%" height="100%" fill="url(#glow)" />
    </Svg>
  )
}
