// Native has no window-level scroll event to hook a single global listener
// into (unlike web, where `useHoverTrigger` just listens on `window` with
// `capture: true`) - every RN `ScrollView`/`FlatList` only reports scrolling
// to its own `onScroll` prop. This is the native-side equivalent: any
// scrollable wires its `onScroll` to `notifyScroll`, and anything that wants
// to react (currently: hiding an open tooltip/hover-card so it doesn't stay
// stuck floating over content that scrolled out from under it) subscribes.
type Listener = () => void

const listeners = new Set<Listener>()

export function subscribeScroll(listener: Listener): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function notifyScroll(): void {
  listeners.forEach((listener) => listener())
}
