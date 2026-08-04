import { useEffect } from 'react'
import { YStack } from 'tamagui'

const KEYFRAMES_ID = 'jops-bubble-kf'

// Injects the @keyframes block once per document. Idempotent (id-guarded)
// and SSR-safe (`output: "single"` still does a static pass, where
// `document` is undefined) so this is safe to call from every mount,
// including React StrictMode's deliberate double-invoke.
function ensureBubbleKeyframes() {
  if (typeof document === 'undefined') return
  if (document.getElementById(KEYFRAMES_ID)) return

  const style = document.createElement('style')
  style.id = KEYFRAMES_ID
  style.textContent = `
    @keyframes jopsRise {
      0%   { transform: translate3d(0, 110vh, 0) scale(0.7); opacity: 0; }
      10%  { opacity: 0.9; }
      90%  { opacity: 0.9; }
      100% { transform: translate3d(28px, -12vh, 0) scale(1.05); opacity: 0; }
    }
    @media (prefers-reduced-motion: reduce) {
      .jops-bubble { animation: none !important; }
    }
  `
  document.head.appendChild(style)
}

// Deterministic geometry — no Math.random() at render, which would differ
// between the static export pass and client hydration and cause a mismatch.
const BUBBLE_GEOMETRY = [
  { size: 22, leftPct: 6, dur: 18, delay: -2 },
  { size: 14, leftPct: 16, dur: 14, delay: -9 },
  { size: 30, leftPct: 26, dur: 22, delay: -4 },
  { size: 18, leftPct: 38, dur: 16, delay: -11 },
  { size: 26, leftPct: 48, dur: 20, delay: -1 },
  { size: 12, leftPct: 58, dur: 13, delay: -7 },
  { size: 34, leftPct: 68, dur: 24, delay: -14 },
  { size: 16, leftPct: 76, dur: 15, delay: -3 },
  { size: 28, leftPct: 84, dur: 19, delay: -10 },
  { size: 20, leftPct: 91, dur: 17, delay: -6 },
  { size: 24, leftPct: 12, dur: 21, delay: -16 },
  { size: 15, leftPct: 55, dur: 12, delay: -5 },
  { size: 32, leftPct: 96, dur: 23, delay: -18 },
  { size: 19, leftPct: 33, dur: 16, delay: -8 },
  { size: 27, leftPct: 72, dur: 20, delay: -12 },
  { size: 13, leftPct: 44, dur: 14, delay: -0 },
]

type BubbleFieldProps = { count: number }

// Pure CSS keyframes, no per-frame JS, no React re-renders — only
// `transform` and `opacity` animate, so this stays fully GPU-composited on
// web and survives tab throttling.
export function BubbleField({ count }: BubbleFieldProps) {
  useEffect(ensureBubbleKeyframes, [])

  const bubbles = BUBBLE_GEOMETRY.slice(0, count)

  return (
    <>
      {bubbles.map((bubble, index) => (
        <YStack
          key={index}
          className="jops-bubble"
          position="absolute"
          b={0}
          l={`${bubble.leftPct}%`}
          width={bubble.size}
          height={bubble.size}
          rounded="$circle"
          bg="$bubbleFill"
          borderWidth={1.5}
          borderColor="$bubbleEdge"
          style={
            {
              animationName: 'jopsRise',
              animationDuration: `${bubble.dur}s`,
              animationDelay: `${bubble.delay}s`,
              animationIterationCount: 'infinite',
              animationTimingFunction: 'linear',
            } as React.CSSProperties as object
          }
        >
          <YStack
            position="absolute"
            t="18%"
            l="18%"
            width="30%"
            height="30%"
            rounded="$circle"
            bg="rgba(255,255,255,0.75)"
          />
        </YStack>
      ))}
    </>
  )
}
