import { useEffect, useRef, useState } from 'react'
import type { View } from 'react-native'

export type ViewportBucket = 'visible' | 'near' | 'far'

// `root: null` intersects against the real browser viewport, not a Tamagui
// ScrollView's own div - correct even though the six catalogue screens
// scroll via that div, because `public/index.html` disables body scroll, so
// the ScrollView's div effectively *is* the scrolling viewport and a null
// root still respects ancestor clipping. `rootMargin: 400px` is the "near"
// prefetch band; anything outside both thresholds is "far" but still gets
// queued (see image-queue.ts - far items aren't skipped, just deprioritized).
export function useInViewport(ref: React.RefObject<View | null>): ViewportBucket {
  const [bucket, setBucket] = useState<ViewportBucket>(
    typeof IntersectionObserver === 'undefined' ? 'visible' : 'far'
  )
  const nearObserved = useRef(false)

  useEffect(() => {
    if (typeof IntersectionObserver === 'undefined') return
    const node = ref.current as unknown as Element | null
    if (!node) return

    const visibleObserver = new IntersectionObserver(
      ([entry]) =>
        setBucket(entry?.isIntersecting ? 'visible' : nearObserved.current ? 'near' : 'far'),
      { root: null, threshold: 0 }
    )
    const nearObserver = new IntersectionObserver(
      ([entry]) => {
        nearObserved.current = entry?.isIntersecting ?? false
        setBucket((prev) => (prev === 'visible' ? prev : nearObserved.current ? 'near' : 'far'))
      },
      { root: null, rootMargin: '400px 0px' }
    )
    visibleObserver.observe(node)
    nearObserver.observe(node)
    return () => {
      visibleObserver.disconnect()
      nearObserver.disconnect()
    }
  }, [ref])

  return bucket
}
