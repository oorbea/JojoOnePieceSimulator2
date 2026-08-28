import { useCallback, useEffect, useRef } from 'react'

import { subscribeScroll } from '@/shared/lib/scroll-bus'

type Zone = { x: number; y: number; width: number; height: number }

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

// Registers each team column's on-screen bounds (via `.measure()`, same
// page-coordinate mechanics `tooltip.tsx` already uses) and answers "which
// zone, if any, contains this drop point" for `use-player-drag.ts`'s
// `onPanResponderRelease` pageX/pageY. Re-measures on scroll (the same
// `scroll-bus.ts` that hides stuck tooltips) since a column's page position
// drifts as the lobby screen scrolls but `onLayout` alone never fires for
// that - only for an actual resize/reflow.
export function useDropZones() {
  const zones = useRef(new Map<string, Zone>())
  const nodes = useRef(new Map<string, Measurable | null>())

  const remeasure = useCallback((id: string) => {
    const node = nodes.current.get(id)
    node?.measure?.((_x, _y, width, height, pageX, pageY) => {
      zones.current.set(id, { x: pageX, y: pageY, width, height })
    })
  }, [])

  useEffect(
    () => subscribeScroll(() => nodes.current.forEach((_node, id) => remeasure(id))),
    [remeasure]
  )

  const registerZone = useCallback(
    (id: string) => (node: Measurable | null) => {
      nodes.current.set(id, node)
    },
    []
  )

  const onZoneLayout = useCallback(
    (id: string) => () => {
      // `onLayout`'s own `layout` is parent-relative, not the page
      // coordinates a drop point is compared against - `.measure()` gives
      // page coordinates, but only reads correctly a tick after layout
      // settles, same one-frame deferral `tooltip.tsx` documents.
      requestAnimationFrame(() => remeasure(id))
    },
    [remeasure]
  )

  const resolveZone = useCallback((pageX: number, pageY: number): string | null => {
    for (const [id, zone] of zones.current) {
      if (
        pageX >= zone.x &&
        pageX <= zone.x + zone.width &&
        pageY >= zone.y &&
        pageY <= zone.y + zone.height
      ) {
        return id
      }
    }
    return null
  }, [])

  return { registerZone, onZoneLayout, resolveZone }
}
