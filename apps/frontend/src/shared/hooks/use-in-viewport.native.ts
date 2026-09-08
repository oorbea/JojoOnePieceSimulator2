import { useEffect, useState } from 'react'
import { Dimensions, type View } from 'react-native'

import { subscribeScroll } from '@/shared/lib/scroll-bus'

export type ViewportBucket = 'visible' | 'near' | 'far'

// No IntersectionObserver on native. The six catalogue screens already wire
// `onScroll={notifyScroll}` (see scroll-bus.ts) - reused here instead of
// touching PageShell. `measureInWindow` gives absolute on-screen position
// directly, so recomputing on every scroll tick is enough; no offset
// bookkeeping needed.
export function useInViewport(ref: React.RefObject<View | null>): ViewportBucket {
  const [bucket, setBucket] = useState<ViewportBucket>('far')

  useEffect(() => {
    const measure = () => {
      const node = ref.current as unknown as {
        measureInWindow?: (
          cb: (x: number, y: number, width: number, height: number) => void
        ) => void
      } | null
      if (!node?.measureInWindow) return
      node.measureInWindow((_x, y, _width, height) => {
        const viewportHeight = Dimensions.get('window').height
        if (y + height > 0 && y < viewportHeight) {
          setBucket('visible')
          return
        }
        const distance = Math.min(Math.abs(y), Math.abs(y + height - viewportHeight))
        setBucket(distance < viewportHeight * 1.5 ? 'near' : 'far')
      })
    }
    measure()
    return subscribeScroll(measure)
  }, [ref])

  return bucket
}
