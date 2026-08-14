import { useEffect, useRef, useState } from 'react'

import { revealStepMs } from '@/features/game/lib/loadout-reveal'

type Params = {
  order: string[]
  active: boolean
  reducedMotion: boolean
  markRevealed: () => void
}

type Result = {
  revealedIds: Set<string>
  isRevealing: boolean
  skip: () => void
}

// Drives the one-card-at-a-time loadout flip. markRevealed() is called as
// soon as the reveal STARTS, not when it finishes - the "has this assignment
// been revealed" bookkeeping in the socket store is per-assignment-seq, not
// per-animation-completion, so a remount mid-reveal (e.g. a re-render from an
// unrelated STATE frame) must not restart the whole sequence from scratch.
export function useLoadoutReveal({ order, active, reducedMotion, markRevealed }: Params): Result {
  const [revealedCount, setRevealedCount] = useState(0)
  const timers = useRef<ReturnType<typeof setTimeout>[]>([])

  const clearTimers = () => {
    timers.current.forEach(clearTimeout)
    timers.current = []
  }

  useEffect(() => {
    clearTimers()

    if (!active) {
      setRevealedCount(order.length)
      return clearTimers
    }

    markRevealed()

    if (reducedMotion) {
      setRevealedCount(order.length)
      return clearTimers
    }

    setRevealedCount(0)
    const step = revealStepMs(false)
    order.forEach((_, index) => {
      const timer = setTimeout(() => {
        setRevealedCount((current) => Math.max(current, index + 1))
      }, step * (index + 1))
      timers.current.push(timer)
    })

    return clearTimers
    // eslint-disable-next-line react-hooks/exhaustive-deps -- re-runs only when order/active identity changes, not on every markRevealed/reducedMotion re-render
  }, [order, active])

  const skip = () => {
    clearTimers()
    setRevealedCount(order.length)
  }

  if (!active) {
    return { revealedIds: new Set(order), isRevealing: false, skip: () => {} }
  }

  return {
    revealedIds: new Set(order.slice(0, revealedCount)),
    isRevealing: revealedCount < order.length,
    skip,
  }
}
