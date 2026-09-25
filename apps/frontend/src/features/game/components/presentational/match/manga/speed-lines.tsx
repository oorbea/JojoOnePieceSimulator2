import Svg, { Polygon } from 'react-native-svg'

const FILL_STYLE = { position: 'absolute' as const, top: 0, left: 0, right: 0, bottom: 0 }

const RAY_COUNT = 28
const CENTER = 100
const INNER_RADIUS = 14
const OUTER_RADIUS = 150
const HALF_WIDTH_DEG = 3.4

function toXY(angleDeg: number, radius: number): [number, number] {
  const rad = (angleDeg * Math.PI) / 180
  return [CENTER + radius * Math.cos(rad), CENTER + radius * Math.sin(rad)]
}

// Procedurally generates a ring of thin triangular wedges radiating from
// the centre - a manga "shonen impact" sunburst, alternating opacity so it
// reads as individual rays rather than a solid disc. Static content (no
// per-frame recompute); the caller fades/rotates the WHOLE thing in via a
// wrapping Animated.View's opacity/transform, per this project's
// transform/opacity-only motion norm (gameplay-power-fx.md) - this
// component itself never animates.
export function SpeedLines({ color = '#FFFFFF' }: { color?: string }) {
  const rays = Array.from({ length: RAY_COUNT }, (_, i) => {
    const centerAngle = (360 / RAY_COUNT) * i
    const [ix, iy] = toXY(centerAngle, INNER_RADIUS)
    const [x1, y1] = toXY(centerAngle - HALF_WIDTH_DEG, OUTER_RADIUS)
    const [x2, y2] = toXY(centerAngle + HALF_WIDTH_DEG, OUTER_RADIUS)
    return {
      key: i,
      points: `${ix},${iy} ${x1},${y1} ${x2},${y2}`,
      opacity: i % 2 === 0 ? 0.4 : 0.15,
    }
  })

  return (
    <Svg viewBox="0 0 200 200" style={[FILL_STYLE]} pointerEvents="none">
      {rays.map((r) => (
        <Polygon key={r.key} points={r.points} fill={color} opacity={r.opacity} />
      ))}
    </Svg>
  )
}
