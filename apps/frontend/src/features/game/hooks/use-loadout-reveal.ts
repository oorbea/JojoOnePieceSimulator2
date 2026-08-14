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
//
// That same markRevealed() call flips `active` back to false on the very
// next render (it catches revealedAssignmentSeq up to assignmentSeq in the
// store), so `active` itself cannot be used to decide whether a reveal is
// still in flight - only whether a NEW one should start. Progress is tracked
// with its own `revealing` state instead, set true when a reveal starts and
// false only once every card has actually flipped (or skip() is called).
export function useLoadoutReveal({ order, active, reducedMotion, markRevealed }: Params): Result {
  const orderKey = order.join(',')
  const [revealedCount, setRevealedCount] = useState(0)
  const [revealing, setRevealing] = useState(false)
  const [seededKey, setSeededKey] = useState<string | null>(null)
  const timers = useRef<ReturnType<typeof setTimeout>[]>([])
  const startedRef = useRef<string | null>(null)

  const clearTimers = () => {
    timers.current.forEach(clearTimeout)
    timers.current = []
  }

  // Resets the count the moment a genuinely new reveal starts, during
  // render rather than inside an effect (React's own recommended pattern for
  // "reset state when a derived key changes") - lands before paint (no
  // stale-count flash) and avoids react-hooks/set-state-in-effect's
  // cascading-render warning for an unconditional setState in an effect body.
  if (active && orderKey !== seededKey) {
    setSeededKey(orderKey)
    if (!reducedMotion && order.length > 0) {
      setRevealedCount(0)
      setRevealing(true)
    }
  }

  useEffect(() => {
    if (!active || startedRef.current === orderKey) return clearTimers
    startedRef.current = orderKey

    markRevealed()
    clearTimers()

    if (reducedMotion || order.length === 0) return clearTimers

    const step = revealStepMs(false)
    order.forEach((_, index) => {
      const timer = setTimeout(
        () => {
          setRevealedCount((current) => Math.max(current, index + 1))
          if (index === order.length - 1) setRevealing(false)
        },
        step * (index + 1)
      )
      timers.current.push(timer)
    })

    return clearTimers
    // eslint-disable-next-line react-hooks/exhaustive-deps -- startedRef (keyed on orderKey) guards re-entry; order/reducedMotion/markRevealed are read fresh on the run they gate, not meant to re-trigger it on their own
  }, [active, orderKey])

  const skip = () => {
    clearTimers()
    setRevealedCount(order.length)
    setRevealing(false)
  }

  if (reducedMotion || !revealing) {
    return { revealedIds: new Set(order), isRevealing: false, skip: () => {} }
  }

  return {
    revealedIds: new Set(order.slice(0, revealedCount)),
    isRevealing: true,
    skip,
  }
}
