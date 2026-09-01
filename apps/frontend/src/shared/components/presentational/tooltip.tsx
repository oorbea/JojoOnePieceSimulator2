import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react'
import { Dimensions, Modal, type LayoutChangeEvent } from 'react-native'
import { createPortal } from 'react-dom'
import { YStack } from 'tamagui'

import { clampOverlayPosition } from '@/shared/lib/overlay-position'
import { subscribeScroll } from '@/shared/lib/scroll-bus'
import { isWeb } from '@/shared/lib/web-blur'

import { GlassPanel } from './glass-panel'
import { GlowText } from './glow-text'

const NATIVE_AUTO_HIDE_MS = 1800
const VIEWPORT_EDGE_MARGIN = 8

type Anchor = { x: number; y: number; width: number; height: number }

type Measurable = {
  measure?: (
    callback: (
      x: number,
      y: number,
      width: number,
      height: number,
      pageX: number,
      pageY: number
    ) => void
  ) => void
}

type HoverTriggerOptions = {
  /** Web only: how long the pointer must hover before `visible` flips true.
   * Native has no hover, so this only delays `onHoverIn`/`onFocus`. */
  delayMs?: number
  /** Native only: auto-hide after this many ms once a long-press reveals it
   * (the default bubble behaviour - lift your finger, still get a moment to
   * read it). Pass `null` to disable the timer and dismiss on `onPressOut`
   * instead - what a hover CARD wants, since a card is read while still
   * pressing/hovering, not after release. */
  nativeAutoHideMs?: number | null
}

// The cross-platform show/hide/anchor mechanics shared by every tooltip-like
// overlay in this file (the plain text bubble AND the bigger hover card):
// web shows on hover/focus (optionally after `delayMs`), native has no
// hover at all so it shows on long-press instead. Anchoring is via
// `.measure()`, not a nested `position:absolute`, for the same reason
// `useTooltipTrigger` always has: RN only compares z-index between direct
// siblings, so a trigger a few levels deep would otherwise render clipped
// or painted over - see `TooltipBubble`'s doc for the full explanation.
export function useHoverTrigger(opts?: HoverTriggerOptions) {
  const delayMs = opts?.delayMs ?? 0
  const nativeAutoHideMs =
    opts?.nativeAutoHideMs === undefined ? NATIVE_AUTO_HIDE_MS : opts.nativeAutoHideMs

  const ref = useRef<Measurable | null>(null)
  const [visible, setVisible] = useState(false)
  const [anchor, setAnchor] = useState<Anchor | null>(null)
  const showTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const hideTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const clearShowTimer = () => {
    if (showTimer.current) {
      clearTimeout(showTimer.current)
      showTimer.current = null
    }
  }
  const clearHideTimer = () => {
    if (hideTimer.current) {
      clearTimeout(hideTimer.current)
      hideTimer.current = null
    }
  }

  const show = useCallback(() => {
    ref.current?.measure?.((_x, _y, width, height, pageX, pageY) => {
      setAnchor({ x: pageX, y: pageY, width, height })
      setVisible(true)
    })
  }, [])

  const hide = useCallback(() => {
    clearShowTimer()
    setVisible(false)
  }, [])

  const scheduleShow = useCallback(() => {
    clearShowTimer()
    if (delayMs > 0) {
      showTimer.current = setTimeout(show, delayMs)
    } else {
      show()
    }
  }, [delayMs, show])

  // Neither timer is tied to `visible` or any other state the existing
  // effects clean up, so a trigger unmounting mid-delay (web hover) or
  // mid-auto-hide-window (native long-press) would otherwise fire `show`/
  // `hide` against a gone component - harmless here since both only ever
  // touch this hook's own refs/state, but it leaks the timer and is the
  // documented root cause of tooltip.test.tsx's flakiness under parallel
  // runs (a still-pending timer from one test firing during the next).
  useEffect(() => {
    return () => {
      clearShowTimer()
      clearHideTimer()
    }
  }, [])

  // Neither platform's hover/focus events fire when the trigger scrolls out
  // from under a stationary pointer/finger, so without this the bubble is
  // left stuck floating on screen, anchored to wherever the trigger used to
  // be, until some unrelated hover/focus change happens to clear it. Web has
  // a real global scroll signal (`capture: true` catches scrolling on any
  // nested container, since `scroll` doesn't bubble); native scrollables
  // have no such thing, so they opt in individually via `notifyScroll`
  // (`scroll-bus.ts`) - this only needs to subscribe while actually visible.
  useEffect(() => {
    if (!visible) return
    if (isWeb) {
      window.addEventListener('scroll', hide, { capture: true, passive: true })
      return () => window.removeEventListener('scroll', hide, { capture: true })
    }
    return subscribeScroll(hide)
  }, [visible, hide])

  const triggerProps = isWeb
    ? {
        onHoverIn: scheduleShow,
        onHoverOut: hide,
        onFocus: scheduleShow,
        onBlur: hide,
      }
    : {
        onLongPress: () => {
          show()
          if (nativeAutoHideMs !== null) {
            clearHideTimer()
            hideTimer.current = setTimeout(hide, nativeAutoHideMs)
          }
        },
        // Only the auto-hide-less (card) variant dismisses on release -
        // the plain bubble relies on its timer so lifting your finger
        // doesn't instantly hide the thing you just long-pressed to read.
        onPressOut: nativeAutoHideMs === null ? hide : undefined,
      }

  return { visible, anchor, triggerRef: ref, triggerProps }
}

// useTooltipTrigger is useHoverTrigger specialised for a short text label -
// every existing call site (GlossButton, InfoHint, ...) passes a label and
// gets the original bubble behaviour unchanged. `delayMs` defaults to 0 so
// none of them start hovering-with-a-delay by accident.
export function useTooltipTrigger(label?: string, opts?: { delayMs?: number }) {
  const hover = useHoverTrigger({ delayMs: opts?.delayMs })
  if (!label) {
    return {
      visible: false,
      anchor: null,
      triggerRef: hover.triggerRef,
      triggerProps: {} as Record<string, unknown>,
    }
  }
  return hover
}

type Size = { width: number; height: number }

// Shared "where does the bubble/card go" logic for both TooltipBubble and
// TooltipCard below: measures the overlay's own rendered size via
// `onLayout` (fires on web AND native, unlike the old web-only
// `.offsetWidth` ref-callback trick this replaced), then computes an
// explicit `top`/`left` pixel position clamped to the viewport - flipping
// above/below the anchor when there isn't room, and clamping the far edge
// too so a card taller than the space on EITHER side still stays fully
// on-screen instead of running off whichever edge it flipped away from
// (the bug: a ~450px LoadoutCard hover-card is routinely taller than the
// room above or below a roster tile, and the old logic only ever checked
// the near edge, never the far one). `Dimensions.get('window')` (not
// `window.innerWidth/innerHeight`) so the exact same math runs on both
// platforms - native previously got no clamping at all.
function usePositionedOverlay(visible: boolean, anchor: Anchor | null) {
  const [measured, setMeasured] = useState<Size | null>(null)
  const [wasVisible, setWasVisible] = useState(visible)

  // A stale size from a previous open must not position the very first
  // frame of a new one - reset it the moment this overlay hides. Compared
  // and reset during render (state, not a ref - `react-hooks/refs` flags
  // reading/writing a ref's `.current` during render) rather than in an
  // effect, which would trip `react-hooks/set-state-in-effect`'s
  // cascading-render warning for an unconditional setState in an effect
  // body - same "reset state when a derived value changes" pattern
  // use-loadout-reveal.ts's `seededKey` already documents.
  if (wasVisible !== visible) {
    setWasVisible(visible)
    if (!visible && measured !== null) setMeasured(null)
  }

  const onLayout = useCallback((e: LayoutChangeEvent) => {
    const { width, height } = e.nativeEvent.layout
    setMeasured((prev) =>
      prev && prev.width === width && prev.height === height ? prev : { width, height }
    )
  }, [])

  if (!anchor) return { onLayout, left: 0, top: 0 }

  // Before the first onLayout, fall back to a reasonable guess - corrects
  // to the real clamp on the very next frame once measured, the same
  // one-tick tolerance the previous approach already had.
  const size = { width: measured?.width ?? anchor.width, height: measured?.height ?? 40 }
  const { top, left } = clampOverlayPosition(
    anchor,
    size,
    Dimensions.get('window'),
    VIEWPORT_EDGE_MARGIN
  )

  return { onLayout, left, top }
}

// Web skips RN's `Modal`: RNW's implementation wraps content in a fixed,
// full-viewport div (`ModalAnimation`'s own wrapper, not the one holding our
// content) that has no `pointer-events` style of its own and no prop to set
// one - it swallows hover/click for the whole screen for as long as ANY
// overlay is visible. That stole hover off the trigger the instant the
// overlay opened, flipping visible on/off forever and blocking every click
// behind it. A plain fixed-position portal to `document.body` gives the
// same "floats above everything, anchored to a measured screen position"
// behavior without that wrapper.
function OverlayPortal({ children }: { children: ReactNode }) {
  if (isWeb) {
    return createPortal(
      <div style={{ position: 'fixed', inset: 0, pointerEvents: 'none' }}>{children}</div>,
      document.body
    )
  }
  return (
    <Modal visible transparent animationType="none" statusBarTranslucent>
      <YStack flex={1} style={{ pointerEvents: 'none' }}>
        {children}
      </YStack>
    </Modal>
  )
}

type TooltipBubbleProps = { visible: boolean; label?: string; anchor: Anchor | null }

// Floats above EVERYTHING (its own `Modal` layer), anchored to the
// trigger's measured screen position via explicit `top`/`left` (identical
// math on web and native - see usePositionedOverlay's doc). Inert to the
// pointer throughout, so it never eats the hover/press meant for the
// button underneath.
export function TooltipBubble({ visible, label, anchor }: TooltipBubbleProps) {
  const { onLayout, left, top } = usePositionedOverlay(visible, anchor)

  if (!visible || !label || !anchor) return null

  return (
    <OverlayPortal>
      <YStack
        onLayout={onLayout}
        position="absolute"
        t={top}
        l={left}
        maxW={220}
        style={{ pointerEvents: 'none' }}
      >
        <GlassPanel tone="strong" elevate={2} px="$2.5" py="$1.5" rounded="$pill">
          <GlowText level="label" fontSize="$1" numberOfLines={2}>
            {label}
          </GlowText>
        </GlassPanel>
      </YStack>
    </OverlayPortal>
  )
}

type TooltipCardProps = { visible: boolean; anchor: Anchor | null; children: ReactNode }

// TooltipBubble's bigger sibling: same anchoring/portal/inert-to-pointer
// contract, but hosts arbitrary content (a full LoadoutCard) instead of a
// one-line label - no `maxW`/`numberOfLines` squeeze, the content dictates
// its own size.
export function TooltipCard({ visible, anchor, children }: TooltipCardProps) {
  const { onLayout, left, top } = usePositionedOverlay(visible, anchor)

  if (!visible || !anchor) return null

  return (
    <OverlayPortal>
      <YStack
        onLayout={onLayout}
        position="absolute"
        t={top}
        l={left}
        style={{ pointerEvents: 'none' }}
      >
        {children}
      </YStack>
    </OverlayPortal>
  )
}
