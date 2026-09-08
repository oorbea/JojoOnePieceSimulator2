// Module-level ticker for Skeleton's shimmer, modelled on scroll-bus.ts: one
// interval for the whole app, only running while at least one Skeleton is
// mounted. Deliberately not Reanimated (mocked as synchronous in
// jest.setup.ts, and its web driver over RNW 0.21 is an unnecessary risk
// here) and not CSS keyframes (tamagui-web.css is generated, see
// metro.config.js). Just an opacity toggle on a `transition="quick"` prop,
// the same vocabulary meter-bar.tsx already uses.
type Listener = (lit: boolean) => void

const listeners = new Set<Listener>()
let interval: ReturnType<typeof setInterval> | null = null
let lit = false

const INTERVAL_MS = 900

function tick(): void {
  lit = !lit
  listeners.forEach((listener) => listener(lit))
}

export function subscribeShimmer(listener: Listener): () => void {
  listeners.add(listener)
  listener(lit)
  if (!interval) interval = setInterval(tick, INTERVAL_MS)
  return () => {
    listeners.delete(listener)
    if (listeners.size === 0 && interval) {
      clearInterval(interval)
      interval = null
    }
  }
}

export function __resetShimmerTicker(): void {
  if (interval) clearInterval(interval)
  interval = null
  lit = false
  listeners.clear()
}
