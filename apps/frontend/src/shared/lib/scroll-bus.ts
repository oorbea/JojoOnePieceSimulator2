// Native has no window-level scroll event to hook a single global listener
// into (unlike web, where `useHoverTrigger` just listens on `window` with
// `capture: true`) - every RN `ScrollView`/`FlatList` only reports scrolling
// to its own `onScroll` prop. This is the native-side equivalent: any
// scrollable wires its `onScroll` to `notifyScroll`, and anything that wants
// to react (currently: hiding an open tooltip/hover-card so it doesn't stay
// stuck floating over content that scrolled out from under it) subscribes.
//
// Widened (backward-compatibly - existing listeners take no args and ignore
// the payload, see tooltip.tsx) for use-in-viewport.native.ts: every screen
// already wires `onScroll={notifyScroll}` directly as the RN scroll handler,
// which always calls it with the native scroll event, so this only starts
// reading what was already being passed and previously discarded.
// Named `ScrollTick`, not `...Payload` - contracts.test.ts's wire-type guard
// flags any `type \w*Payload = {...}` declaration as a hand-mirrored wire
// type, and this one isn't (it never crosses the network).
type ScrollTick = { offsetY: number }
type Listener = (tick?: ScrollTick) => void

type NativeScrollEventLike = { nativeEvent?: { contentOffset?: { y?: number } } }

const listeners = new Set<Listener>()
let lastOffsetY = 0

export function subscribeScroll(listener: Listener): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function notifyScroll(event?: NativeScrollEventLike): void {
  const offsetY = event?.nativeEvent?.contentOffset?.y
  if (typeof offsetY === 'number') lastOffsetY = offsetY
  const tick: ScrollTick = { offsetY: lastOffsetY }
  listeners.forEach((listener) => listener(tick))
}

export function currentScrollOffsetY(): number {
  return lastOffsetY
}
