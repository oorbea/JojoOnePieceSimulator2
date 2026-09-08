import { useEffect, useRef, useState } from 'react'

import {
  enqueueImage,
  priorityFor,
  type ImageHandle,
  type QueueLane,
} from '@/shared/lib/image-queue'

export type ImageSlotState = 'queued' | 'granted' | 'loaded' | 'error'

type Visibility = 'visible' | 'near' | 'far'

// React binding over shared/lib/image-queue.ts. `uri` re-enqueues on change
// (a different image); `retryToken` re-enqueues without changing `uri` (the
// user pressed retry); `visibility`/`order` only reprioritize an entry
// that's still queued - once granted a slot, priority doesn't matter.
export function useImageSlot(opts: {
  uri: string | null
  visibility: Visibility
  order: number
  lane?: QueueLane
  retryToken?: number
}) {
  const { uri, visibility, order, lane = 'grid', retryToken = 0 } = opts
  const sessionKey = `${uri ?? ''}|${retryToken}`
  const [state, setState] = useState<ImageSlotState>(uri ? 'queued' : 'error')
  const [seenKey, setSeenKey] = useState(sessionKey)
  const handleRef = useRef<ImageHandle | null>(null)

  // Reset during render, not inside the effect below - React's own pattern
  // for "reset state when a derived key changes" (see use-loadout-reveal.ts
  // for the precedent in this repo), and what keeps this compliant with
  // react-hooks/set-state-in-effect: an unconditional setState as the first
  // thing an effect does is flagged as a likely cascading-render bug, while
  // the enqueueImage callbacks below (onGrant/onTimeout) are the legitimate
  // "external system changed, sync it into state" case the rule allows.
  if (sessionKey !== seenKey) {
    setSeenKey(sessionKey)
    setState(uri ? 'queued' : 'error')
  }

  useEffect(() => {
    if (!uri) return
    const handle = enqueueImage({
      key: uri,
      priority: priorityFor(visibility, order),
      lane,
      onGrant: () => setState('granted'),
      onTimeout: () => setState('error'),
    })
    handleRef.current = handle
    return () => {
      handle.cancel()
      handleRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- visibility/order changes reprioritize via the effect below, not a re-enqueue
  }, [uri, lane, retryToken])

  useEffect(() => {
    handleRef.current?.reprioritize(priorityFor(visibility, order))
  }, [visibility, order])

  const onLoad = () => {
    handleRef.current?.settle()
    setState('loaded')
  }
  const onError = () => {
    handleRef.current?.settle()
    setState('error')
  }

  return { state, onLoad, onError }
}
