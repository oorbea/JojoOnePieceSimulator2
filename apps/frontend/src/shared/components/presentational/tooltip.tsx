import { useCallback, useRef, useState } from 'react'
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
    callback: (x: number, y: number, width: number, height: number, pageX: number, pageY: number) => void
  ) => void
}

// Cross-platform tooltip trigger: web shows on hover or keyboard focus (a
// mouse/keyboard user has no long-press affordance); native has no hover at
// all, so it shows on long-press with an auto-hide timer instead.
//
// The bubble itself renders through a root-level `Modal` (see
// `TooltipBubble` below), anchored to the trigger's on-screen position via
// `.measure()` - NOT a nested `position:absolute` inside the trigger's own
// tree. RN only compares z-index between direct siblings (the exact reason
// `ConfirmSheet` is a real `Modal` instead of an absolute overlay, per
// ObsidianVault/frontend-responsive-frutiger-aero.md) - a tooltip a few
// levels deep inside a form would otherwise render clipped by an ancestor's
// `overflow:hidden` or painted over by a later, unrelated sibling.
export function useTooltipTrigger(label?: string) {
  const ref = useRef<Measurable | null>(null)
  const [visible, setVisible] = useState(false)
  const [anchor, setAnchor] = useState<Anchor | null>(null)
  const hideTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  if (!label) {
    return { visible: false, anchor: null, triggerRef: ref, triggerProps: {} as Record<string, unknown> }
  }

  const clearHideTimer = () => {
    if (hideTimer.current) {
      clearTimeout(hideTimer.current)
      hideTimer.current = null
    }
  }

  const show = () => {
    ref.current?.measure?.((_x, _y, width, height, pageX, pageY) => {
      setAnchor({ x: pageX, y: pageY, width, height })
      setVisible(true)
    })
  }

  const triggerProps = isWeb
    ? {
        onHoverIn: show,
        onHoverOut: () => setVisible(false),
        onFocus: show,
        onBlur: () => setVisible(false),
      }
    : {
        onLongPress: () => {
          show()
          clearHideTimer()
          hideTimer.current = setTimeout(() => setVisible(false), NATIVE_AUTO_HIDE_MS)
        },
      }

  return { visible, anchor, triggerRef: ref, triggerProps }
}

type TooltipBubbleProps = { visible: boolean; label?: string; anchor: Anchor | null }

// Floats above EVERYTHING (its own `Modal` layer), anchored to the
// trigger's measured screen position. Centering above the anchor uses a
// web-only `transform` (RN's native transform doesn't take percentage
// values, so native anchors from the trigger's top-left corner instead -
// fine for the short labels these carry). Inert to the pointer throughout,
// so it never eats the hover/press meant for the button underneath.
export function TooltipBubble({ visible, label, anchor }: TooltipBubbleProps) {
  // Centering the bubble on the anchor's midpoint (via `translateX(-50%)`)
  // pushes it straight off-screen for any trigger near a screen edge (e.g.
  // "Admin" in the top-right nav) - `left`/`right` never clamped to the
  // viewport. Same edge problem vertically: a trigger flush against the top
  // of the viewport (the nav bar's own buttons) has no room above it, so
  // anchoring above and translating up by the bubble's own height
  // (`translateY(-100%)`) pushed the whole thing off the top of the screen.
  const [clampedCenterX, setClampedCenterX] = useState<number | null>(null)
  const [placement, setPlacement] = useState<'above' | 'below'>('above')

  // Both corrections need the bubble's REAL rendered size, which only exists
  // after it's painted - a `useLayoutEffect` doing the measure-then-setState
  // dance is the obvious way to get it, but `react-hooks/set-state-in-effect`
  // flags any unconditional `setState` in an effect body as a cascading-render
  // risk (see `join-lobby-container.tsx`/`use-loadout-reveal.ts` for the
  // render-time-setState escape hatch this project otherwise uses for that
  // rule - doesn't apply here since it only works for state derivable from
  // props, and DOM layout isn't available until after commit). A ref
  // callback runs at that same post-paint point in the commit, but isn't an
  // effect the rule's static check recognizes, so it's the standard React
  // pattern for exactly this ("measuring a node", react.dev) without
  // tripping the lint rule.
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
    [visible, anchor?.x, anchor?.y, anchor?.width, anchor?.height, label]
  )

  if (!visible || !label || !anchor) return null

  const centerX = clampedCenterX ?? anchor.x + anchor.width / 2
  const topY = placement === 'below' ? anchor.y + anchor.height + 8 : Math.max(anchor.y - 8, 4)

  const bubble = (
    <YStack
      ref={measureAndPosition as never}
      position="absolute"
      t={topY}
      l={centerX}
      maxW={220}
      style={{
        pointerEvents: 'none',
        ...(isWeb
          ? { transform: [{ translateX: '-50%' }, { translateY: placement === 'below' ? '0%' : '-100%' }] }
          : null),
      }}
    >
      <GlassPanel tone="strong" elevate={2} px="$2.5" py="$1.5" rounded="$pill">
        <GlowText level="label" fontSize="$1" numberOfLines={2}>
          {label}
        </GlowText>
      </GlassPanel>
    </YStack>
  )

  // Web skips RN's `Modal`: RNW's implementation wraps content in a fixed,
  // full-viewport div (`ModalAnimation`'s own wrapper, not the one holding our
  // content) that has no `pointer-events` style of its own and no prop to set
  // one - it swallows hover/click for the whole screen for as long as ANY
  // tooltip is visible. That stole hover off the trigger the instant the
  // tooltip opened, flipping visible on/off forever and blocking every click
  // behind it. A plain fixed-position portal to `document.body` gives the
  // same "floats above everything, anchored to a measured screen position"
  // behavior without that wrapper.
  if (isWeb) {
    return createPortal(
      <div style={{ position: 'fixed', inset: 0, pointerEvents: 'none' }}>{bubble}</div>,
      document.body
    )
  }

  return (
    <Modal visible transparent animationType="none" statusBarTranslucent>
      <YStack flex={1} style={{ pointerEvents: 'none' }}>
        {bubble}
      </YStack>
    </Modal>
  )
}
