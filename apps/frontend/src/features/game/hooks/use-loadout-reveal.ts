import { useEffect, useRef, useState } from 'react'

import { flipStepMs, slotStepMs, type RevealStep } from '@/features/game/lib/loadout-reveal'

type Params = {
  steps: RevealStep[]
  active: boolean
  reducedMotion: boolean
  markRevealed: () => void
}

type Result = {
  revealedIds: Set<string>
  /** How many of each participant's loadout slots are visible so far, in
   * `loadoutSlots` order. A participant absent from the map (not yet
   * flipped) has 0 visible slots. */
  visibleSlotsById: Record<string, number>
  isRevealing: boolean
  skip: () => void
}

// Reduces the flat step timeline down to "revealed up to step `count`
// (exclusive)" - a card is flipped the moment its own flip step (slot -1)
// is reached; a slot becomes visible the moment its own step is reached.
function deriveState(steps: RevealStep[], count: number): { revealedIds: Set<string>; visibleSlotsById: Record<string, number> } {
  const revealedIds = new Set<string>()
  const visibleSlotsById: Record<string, number> = {}
  const upTo = Math.min(count, steps.length)
  for (let i = 0; i < upTo; i++) {
    const step = steps[i]
    if (step.slot === -1) {
      revealedIds.add(step.participantId)
      continue
    }
    visibleSlotsById[step.participantId] = Math.max(visibleSlotsById[step.participantId] ?? 0, step.slot + 1)
  }
  return { revealedIds, visibleSlotsById }
}

// Drives the poder-a-poder loadout reveal: each participant's card flips,
// then its loadout slots (physicalForm, stand, devilFruit, ...) pop in one
// at a time, in the exact order LoadoutBuilder drew them. markRevealed() is
// called as soon as the reveal STARTS, not when it finishes - the "has this
// assignment been revealed" bookkeeping in the socket store is per-
// assignment-seq, not per-animation-completion, so a remount mid-reveal
// (e.g. a re-render from an unrelated STATE frame) must not restart the
// whole sequence from scratch.
//
// The bug this hook used to have (fixed 2026-08-14, see
// game-match-assignment-frontend.md): the scheduling effect returned
// `clearTimers` as its cleanup, keyed on `[active, stepsKey]`. But
// `markRevealed()` flips `active` back to `false` on the very next render
// (it catches `revealedAssignmentSeq` up to `assignmentSeq` in the store) -
// so React ran that cleanup one render after scheduling, cancelling every
// timer before the first one could ever fire. The animation never played;
// only Skip (which doesn't go through this effect) could reveal anything,
// and it revealed everything at once. Fix: this effect no longer returns a
// cleanup tied to `active` flipping - pending timers are only cleared (a)
// right before a genuinely NEW sequence schedules its own timers, and (b)
// on unmount, via a separate effect with an empty dependency array.
export function useLoadoutReveal({ steps, active, reducedMotion, markRevealed }: Params): Result {
  const stepsKey = steps.map((s) => `${s.participantId}:${s.slot}`).join(',')
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
  if (active && stepsKey !== seededKey) {
    setSeededKey(stepsKey)
    if (!reducedMotion && steps.length > 0) {
      setRevealedCount(0)
      setRevealing(true)
    }
  }

  useEffect(() => {
    if (!active || startedRef.current === stepsKey) return
    startedRef.current = stepsKey

    markRevealed()
    // Clear any timers left over from a previous sequence before scheduling
    // this one - NOT returned as this effect's cleanup (see the file-level
    // comment above for why that distinction is the actual fix).
    clearTimers()

    if (reducedMotion || steps.length === 0) return

    let elapsedMs = 0
    steps.forEach((step, index) => {
      elapsedMs += step.slot === -1 ? flipStepMs(false) : slotStepMs(false)
      const timer = setTimeout(() => {
        setRevealedCount((current) => Math.max(current, index + 1))
        if (index === steps.length - 1) setRevealing(false)
      }, elapsedMs)
      timers.current.push(timer)
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps -- startedRef (keyed on stepsKey) guards re-entry; steps/reducedMotion/markRevealed are read fresh on the run they gate, not meant to re-trigger it on their own
  }, [active, stepsKey])

  // Only clears pending timers on unmount - deliberately not tied to
  // `active`/`stepsKey` changing (that's the bug described above).
  useEffect(() => clearTimers, [])

  const skip = () => {
    clearTimers()
    setRevealedCount(steps.length)
    setRevealing(false)
  }

  if (reducedMotion || !revealing) {
    return { ...deriveState(steps, steps.length), isRevealing: false, skip: () => {} }
  }

  return { ...deriveState(steps, revealedCount), isRevealing: true, skip }
}
