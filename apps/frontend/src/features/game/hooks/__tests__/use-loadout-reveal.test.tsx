import { act, render, screen } from '@testing-library/react-native'
import { useEffect, useState } from 'react'
import { Text } from 'react-native'

import { useLoadoutReveal } from '@/features/game/hooks/use-loadout-reveal'
import { REVEAL_INTRO_MS, REVEAL_SPIN_MS, revealDurationMs } from '@/features/game/lib/loadout-reveal'
import type { Manga } from '@/shared/lib/zod'

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
function Harness({ mangas, serverRevealMs }: { mangas: Manga[]; serverRevealMs: number | null }) {
  const [revealedOnce, setRevealedOnce] = useState(false)
  const active = !revealedOnce
  const result = useLoadoutReveal({
    mangas,
    active,
    markRevealed: () => setRevealedOnce(true),
    serverRevealMs,
  })
  return (
    <Text testID="state">
      {JSON.stringify({
        isRevealing: result.isRevealing,
        phase: result.phase,
        slotIndex: result.slotIndex,
        totalSlots: result.totalSlots,
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

  it('still plays through every phase even though active flips true->false on the same tick markRevealed fires (the historic timer-cancellation bug)', async () => {
    const mangas: Manga[] = ['JOJO']

    await render(<Harness mangas={mangas} serverRevealMs={null} />)

    // Intro hasn't elapsed yet.
    expect(readState()).toEqual({ isRevealing: true, phase: 'intro', slotIndex: -1, totalSlots: 3 })

    // Past the intro: the first slot (stand) starts spinning.
    await advance(REVEAL_INTRO_MS + 1)
    expect(readState().phase).toBe('spin')
    expect(readState().slotIndex).toBe(0)

    // Past the whole timeline: reveal finishes and the last phase is outro.
    await advance(revealDurationMs(mangas))
    expect(readState().isRevealing).toBe(false)
  })

  it('scales every phase by serverRevealMs/localTotal - a constants drift degrades pacing, not sync with the server-authoritative duration', async () => {
    const mangas: Manga[] = ['JOJO']
    const localTotal = revealDurationMs(mangas)
    const serverRevealMs = localTotal * 2 // server thinks the reveal takes twice as long

    await render(<Harness mangas={mangas} serverRevealMs={serverRevealMs} />)

    // At the LOCAL intro duration, the scaled version hasn't reached the
    // first spin yet (it needs roughly double the time).
    await advance(REVEAL_INTRO_MS + 1)
    expect(readState().phase).toBe('intro')

    // The full scaled duration finishes the whole thing.
    await advance(serverRevealMs)
    expect(readState().isRevealing).toBe(false)
  })

  it('skip jumps straight to done', async () => {
    const mangas: Manga[] = ['JOJO', 'ONE_PIECE']
    const skipRef: { current: (() => void) | null } = { current: null }

    function SkipHarness() {
      const result = useLoadoutReveal({ mangas, active: true, markRevealed: () => {}, serverRevealMs: null })
      useEffect(() => {
        skipRef.current = result.skip
      }, [result.skip])
      return <Text testID="state">{JSON.stringify({ isRevealing: result.isRevealing })}</Text>
    }

    await render(<SkipHarness />)
    expect(readState().isRevealing).toBe(true)

    await act(async () => {
      skipRef.current?.()
    })
    expect(readState().isRevealing).toBe(false)
  })

  it('spin phases advance one slot at a time in order', async () => {
    const mangas: Manga[] = ['JOJO']

    await render(<Harness mangas={mangas} serverRevealMs={null} />)

    await advance(REVEAL_INTRO_MS + 1)
    expect(readState()).toMatchObject({ phase: 'spin', slotIndex: 0 })

    await advance(REVEAL_SPIN_MS + 1)
    expect(readState()).toMatchObject({ phase: 'land', slotIndex: 0 })
  })
})
