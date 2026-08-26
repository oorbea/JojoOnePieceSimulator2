import { useCallback, useState } from 'react'

import { isWeb } from '@/shared/lib/web-blur'

export type RovingItemProps = {
  tabIndex: number
  id?: string
  onKeyDown?: (e: { key: string; preventDefault?: () => void }) => void
  onFocus?: () => void
}

type UseRovingGroupOptions = {
  groupId: string
  count: number
  /** Called when an item is activated via Enter/Space (not on plain arrow
   * navigation - moving focus never fires this). */
  onActivate: (index: number) => void
  /** The item that should hold the roving tabIndex=0 when nothing has been
   * focused yet - typically the selected option, so re-entering the group
   * with Tab lands on the current choice instead of always the first. */
  initialIndex?: number
}

// useRovingGroup implements the standard roving-tabindex pattern for a
// same-row set of options (a radio group here; generic enough for any small
// fixed list): exactly one item is a Tab stop at a time (tabIndex 0, every
// other item -1), arrow keys move which one that is (wrapping), Home/End
// jump to the ends, and Enter/Space activate the currently focused item.
//
// Moving focus programmatically needs a real DOM node to call .focus() on.
// GlossButton already owns its own internal ref (for the tooltip's
// .measure() anchoring) and doesn't forward an external one, so this hook
// deliberately focuses via `document.getElementById` off a caller-supplied
// stable `id` per item instead of fighting that internal ref - the same
// escape hatch any component library forces when it doesn't expose ref
// forwarding. Web-only (isWeb branch): native has no keyboard focus model
// to roam, so getItemProps degrades to a plain tabIndex there and the
// existing tap/press handlers stay the only way to interact, unchanged.
export function useRovingGroup({ groupId, count, onActivate, initialIndex = 0 }: UseRovingGroupOptions) {
  const [activeIndex, setActiveIndex] = useState(Math.min(initialIndex, Math.max(0, count - 1)))

  const itemId = useCallback((index: number) => `${groupId}-${index}`, [groupId])

  const focusIndex = useCallback(
    (index: number) => {
      setActiveIndex(index)
      if (typeof document !== 'undefined') {
        document.getElementById(itemId(index))?.focus()
      }
    },
    [itemId]
  )

  const getItemProps = useCallback(
    (index: number): RovingItemProps => {
      if (!isWeb) return { tabIndex: index === activeIndex ? 0 : -1 }

      return {
        tabIndex: index === activeIndex ? 0 : -1,
        id: itemId(index),
        onFocus: () => setActiveIndex(index),
        onKeyDown: (e) => {
          switch (e.key) {
            case 'ArrowRight':
            case 'ArrowDown':
              e.preventDefault?.()
              focusIndex((index + 1) % count)
              break
            case 'ArrowLeft':
            case 'ArrowUp':
              e.preventDefault?.()
              focusIndex((index - 1 + count) % count)
              break
            case 'Home':
              e.preventDefault?.()
              focusIndex(0)
              break
            case 'End':
              e.preventDefault?.()
              focusIndex(count - 1)
              break
            case 'Enter':
            case ' ':
              e.preventDefault?.()
              onActivate(index)
              break
          }
        },
      }
    },
    [activeIndex, count, focusIndex, itemId, onActivate]
  )

  return { activeIndex, getItemProps }
}
