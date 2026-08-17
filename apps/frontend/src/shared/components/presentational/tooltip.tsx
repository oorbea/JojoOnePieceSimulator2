import { useCallback, useRef, useState, type ReactNode } from 'react'
import { Modal } from 'react-native'
import { createPortal } from 'react-dom'
import { YStack } from 'tamagui'

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

// Shared "where does the bubble/card go" logic for both TooltipBubble and
// TooltipCard below - centres on the anchor's midpoint, clamped to the
// viewport, flipping above/below when there isn't room. See TooltipBubble's
// doc for why this needs the bubble's own rendered size (a ref callback,
// not a `useLayoutEffect`, to avoid the `react-hooks/set-state-in-effect`
// cascading-render lint rule).
function usePositionedOverlay(visible: boolean, anchor: Anchor | null) {
  const [clampedCenterX, setClampedCenterX] = useState<number | null>(null)
  const [placement, setPlacement] = useState<'above' | 'below'>('above')

  const measureAndPosition = useCallback(
    (el: HTMLElement | null) => {
      if (!isWeb || !visible || !anchor || !el) {
        setClampedCenterX(null)
        setPlacement('above')
        return
      }
      const width = el.offsetWidth
      const height = el.offsetHeight
      if (!width || !height) return
      const half = width / 2
      const centerX = anchor.x + anchor.width / 2
      const min = half + VIEWPORT_EDGE_MARGIN
      const max = window.innerWidth - half - VIEWPORT_EDGE_MARGIN
      setClampedCenterX(Math.min(Math.max(centerX, min), max))
      setPlacement(anchor.y - 8 - height < VIEWPORT_EDGE_MARGIN ? 'below' : 'above')
    },
    [visible, anchor?.x, anchor?.y, anchor?.width, anchor?.height]
  )

  const centerX = anchor ? (clampedCenterX ?? anchor.x + anchor.width / 2) : 0
  const topY = anchor
    ? placement === 'below'
      ? anchor.y + anchor.height + 8
      : Math.max(anchor.y - 8, 4)
    : 0

  return { measureAndPosition, centerX, topY, placement }
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
// trigger's measured screen position. Centering above the anchor uses a
// web-only `transform` (RN's native transform doesn't take percentage
// values, so native anchors from the trigger's top-left corner instead -
// fine for the short labels these carry). Inert to the pointer throughout,
// so it never eats the hover/press meant for the button underneath.
export function TooltipBubble({ visible, label, anchor }: TooltipBubbleProps) {
  const { measureAndPosition, centerX, topY, placement } = usePositionedOverlay(visible, anchor)

  if (!visible || !label || !anchor) return null

  return (
    <OverlayPortal>
      <YStack
        ref={measureAndPosition as never}
        position="absolute"
        t={topY}
        l={centerX}
        maxW={220}
        style={{
          pointerEvents: 'none',
          ...(isWeb
            ? {
                transform: [
                  { translateX: '-50%' },
                  { translateY: placement === 'below' ? '0%' : '-100%' },
                ],
              }
            : null),
        }}
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
  const { measureAndPosition, centerX, topY, placement } = usePositionedOverlay(visible, anchor)

  if (!visible || !anchor) return null

  return (
    <OverlayPortal>
      <YStack
        ref={measureAndPosition as never}
        position="absolute"
        t={topY}
        l={centerX}
        style={{
          pointerEvents: 'none',
          ...(isWeb
            ? {
                transform: [
                  { translateX: '-50%' },
                  { translateY: placement === 'below' ? '0%' : '-100%' },
                ],
              }
            : null),
        }}
      >
        {children}
      </YStack>
    </OverlayPortal>
  )
}
