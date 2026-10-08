// Whether the most recent pointer input on web came from a finger/pen rather
// than a mouse. Browsers answer a tap with *emulated* mouse events
// (`mouseover`/`mouseenter`, then the click) and only emit the matching
// `mouseleave` when the user taps somewhere else - so anything driven by
// hover-in/hover-out (the tooltip) is left stuck open after a tap. Tamagui's
// own `onMouseEnter` guards its hover *style* with the same idea
// (`lastInteractionWasTouch`) but still calls `onHoverIn` unconditionally.
//
// Pointer events, unlike the emulated mouse events, always carry the real
// device type and are not re-fired for the compatibility mouse events, so the
// last one seen is a reliable answer when `mouseenter` arrives.
let lastWasTouch = false

function record(e: PointerEvent) {
  lastWasTouch = e.pointerType === 'touch' || e.pointerType === 'pen'
}

if (typeof window !== 'undefined' && typeof window.addEventListener === 'function') {
  window.addEventListener('pointerdown', record, { capture: true, passive: true })
  window.addEventListener('pointermove', record, { capture: true, passive: true })
}

export function lastInteractionWasTouch(): boolean {
  return lastWasTouch
}
