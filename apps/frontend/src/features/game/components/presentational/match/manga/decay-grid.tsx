import Svg, { G, Line, Mask, RadialGradient, Rect, Stop } from 'react-native-svg'

const FILL_STYLE = { position: 'absolute' as const, top: 0, left: 0, right: 0, bottom: 0 }

const CELL = 18
const COLS = 12
const ROWS = 12

// "Aura de decadencia" (owner decision, 2026-09-25 playtest feedback -
// see ObsidianVault/game-victory-defeat-cinematic-2026-09-14.md): a defeat
// beat that DECAYS instead of impacting - a faint grid that only holds
// together near the centre and dissolves into black toward the edges,
// explicitly not rays radiating outward (that reads as a triumphant
// impact, wrong for a loss). Two masks do the work: `gridFade` keeps the
// grid lines themselves visible only near the centre, `shade` darkens
// everything toward the edges on top of it - same pair the HTML mockup
// used (`.decay-grid`/`.decay-shade`), reimplemented natively since RN has
// no CSS mask-image/radial-gradient.
export function DecayGrid() {
  const lines = []
  for (let c = 0; c <= COLS; c++) {
    lines.push(
      <Line
        key={`v${c}`}
        x1={c * CELL}
        y1={0}
        x2={c * CELL}
        y2={ROWS * CELL}
        stroke="#fff"
        strokeWidth={1}
      />
    )
  }
  for (let r = 0; r <= ROWS; r++) {
    lines.push(
      <Line
        key={`h${r}`}
        x1={0}
        y1={r * CELL}
        x2={COLS * CELL}
        y2={r * CELL}
        stroke="#fff"
        strokeWidth={1}
      />
    )
  }

  return (
    <Svg
      width="100%"
      height="100%"
      viewBox={`0 0 ${COLS * CELL} ${ROWS * CELL}`}
      style={[FILL_STYLE]}
      preserveAspectRatio="xMidYMid slice"
      pointerEvents="none"
    >
      {/* Gradient/mask defs as direct Svg children (no <Defs> wrapper) -
          react-native-svg renders them identically either way, and this
          sidesteps a TS typing gap in this version's <Defs> children
          type. */}
      <RadialGradient id="gridFade" cx="50%" cy="42%" r="55%">
        <Stop offset="0%" stopColor="#fff" stopOpacity={0.6} />
        <Stop offset="30%" stopColor="#fff" stopOpacity={0.6} />
        <Stop offset="100%" stopColor="#fff" stopOpacity={0} />
      </RadialGradient>
      <Mask id="gridFadeMask">
        <Rect x={0} y={0} width={COLS * CELL} height={ROWS * CELL} fill="url(#gridFade)" />
      </Mask>
      <RadialGradient id="shade" cx="50%" cy="42%" r="60%">
        <Stop offset="0%" stopColor="#000" stopOpacity={0} />
        <Stop offset="30%" stopColor="#000" stopOpacity={0} />
        <Stop offset="100%" stopColor="#000" stopOpacity={0.8} />
      </RadialGradient>
      <G mask="url(#gridFadeMask)">{lines}</G>
      <Rect x={0} y={0} width={COLS * CELL} height={ROWS * CELL} fill="url(#shade)" />
    </Svg>
  )
}
