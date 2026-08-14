import { act, render, screen } from '@testing-library/react-native'
import { useState } from 'react'
import { Text } from 'react-native'

import { useLoadoutReveal } from '@/features/game/hooks/use-loadout-reveal'
import type { RevealStep } from '@/features/game/lib/loadout-reveal'

async function advance(ms: number) {
  await act(async () => {
    jest.advanceTimersByTime(ms)
  })
}

// Reproduces the exact real-world sequence that used to break the reveal
// (fixed 2026-08-14, see game-match-assignment-frontend.md): `active` starts
// true, and the hook's own `markRevealed()` call flips it back to false on
// the very next render - `shouldReveal()` in the real container does this
// via `revealedAssignmentSeq` catching up to `assignmentSeq`. This harness
// models that with its own bit of state instead of the real socket store,
// since only the active:true->false-on-markRevealed shape matters here.
function Harness({ steps, reducedMotion }: { steps: RevealStep[]; reducedMotion: boolean }) {
  const [revealedOnce, setRevealedOnce] = useState(false)
  const active = !revealedOnce
  const result = useLoadoutReveal({
    steps,
    active,
    reducedMotion,
    markRevealed: () => setRevealedOnce(true),
  })
  return (
    <Text testID="state">
      {JSON.stringify({
        revealedIds: [...result.revealedIds].sort(),
        visibleSlotsById: result.visibleSlotsById,
        isRevealing: result.isRevealing,
      })}
    </Text>
  )
}

function readState() {
  return JSON.parse(screen.getByTestId('state').props.children as string)
}

describe('useLoadoutReveal', () => {
  beforeEach(() => {
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('still plays every step even though active flips true->false on the same tick markRevealed fires (the historic timer-cancellation bug)', async () => {
    const steps: RevealStep[] = [
      { participantId: 'p1', slot: -1 },
      { participantId: 'p1', slot: 0 },
      { participantId: 'p1', slot: 1 },
    ]

    await render(<Harness steps={steps} reducedMotion={false} />)

    // Nothing revealed yet - the flip step hasn't fired.
    expect(readState()).toEqual({ revealedIds: [], visibleSlotsById: {}, isRevealing: true })

    // Past the flip step (450ms): the card is face-up, no slots yet.
    await advance(451)
    expect(readState().revealedIds).toEqual(['p1'])
    expect(readState().visibleSlotsById).toEqual({})
    expect(readState().isRevealing).toBe(true)

    // Past the first slot step (450 + 220ms).
    await advance(220)
    expect(readState().visibleSlotsById).toEqual({ p1: 1 })

    // Past the second (last) slot step - reveal finishes.
    await advance(220)
    expect(readState().visibleSlotsById).toEqual({ p1: 2 })
    expect(readState().isRevealing).toBe(false)
  })

  it('reveals everything at once under reduced motion', async () => {
    const steps: RevealStep[] = [
      { participantId: 'p1', slot: -1 },
      { participantId: 'p1', slot: 0 },
    ]

    await render(<Harness steps={steps} reducedMotion />)

    expect(readState()).toEqual({ revealedIds: ['p1'], visibleSlotsById: { p1: 1 }, isRevealing: false })
  })
})
