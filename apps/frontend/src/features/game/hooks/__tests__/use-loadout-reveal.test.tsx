import { act, render, screen } from '@testing-library/react-native'
import { useEffect, useState } from 'react'
import { Text } from 'react-native'

import { useLoadoutReveal } from '@/features/game/hooks/use-loadout-reveal'
import {
  REVEAL_INTRO_MS,
  REVEAL_NARRATOR_MS,
  REVEAL_PLAYER_INTRO_MS,
  REVEAL_SPIN_BASE_MS,
  revealDurationMs,
} from '@/features/game/lib/loadout-reveal'
import type { GameParticipant } from '@/features/game/types/game.types'
import type { Manga, RevealSpeed } from '@/shared/contracts/enums'

async function advance(ms: number) {
  await act(async () => {
    jest.advanceTimersByTime(ms)
  })
}

function participant(id: string): GameParticipant {
  return {
    id,
    displayName: id,
    teamId: 't1',
    kind: 'HUMAN',
    connected: true,
    avatarThumb: '',
  }
}

// Reproduces the exact real-world sequence that used to break the reveal
// (fixed 2026-08-14, see game-match-assignment-frontend.md): `active` starts
// true, and the hook's own `markRevealed()` call flips it back to false on
// the very next render - `shouldReveal()` in the real container does this
// via `revealedAssignmentSeq` catching up to `assignmentSeq`. This harness
// models that with its own bit of state instead of the real socket store,
// since only the active:true->false-on-markRevealed shape matters here.
function Harness({
  mangas,
  participants,
  speed = 'SWIFT',
  serverRevealMs,
}: {
  mangas: Manga[]
  participants: GameParticipant[]
  speed?: RevealSpeed
  serverRevealMs: number | null
}) {
  const [revealedOnce, setRevealedOnce] = useState(false)
  const active = !revealedOnce
  const result = useLoadoutReveal({
    gameId: 'g1',
    roundIndex: 0,
    mangas,
    participants,
    speed,
    active,
    markRevealed: () => setRevealedOnce(true),
    sendRevealReady: () => {},
    serverRevealMs,
  })
  return (
    <Text testID="state">
      {JSON.stringify({
        isRevealing: result.isRevealing,
        phase: result.phase,
        participantIndex: result.participantIndex,
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
    const participants = [participant('p1')]

    await render(<Harness mangas={mangas} participants={participants} serverRevealMs={null} />)

    // Intro hasn't elapsed yet.
    expect(readState()).toMatchObject({ isRevealing: true, phase: 'intro', participantIndex: -1 })

    // Past intro + playerIntro: the first slot starts spinning.
    await advance(REVEAL_INTRO_MS + REVEAL_PLAYER_INTRO_MS + 1)
    expect(readState()).toMatchObject({ phase: 'narrator', participantIndex: 0, slotIndex: 0 })

    // Past the whole timeline: reveal finishes.
    await advance(revealDurationMs('g1', 0, mangas, [{ hasStand: false, hasDevilFruit: false, hasArmamentHaki: false, hasObservationHaki: false, hasConquerorHaki: false }], 'SWIFT'))
    expect(readState().isRevealing).toBe(false)
  })

  it('scales every phase by serverRevealMs/localTotal - a constants drift degrades pacing, not sync with the server-authoritative duration', async () => {
    const mangas: Manga[] = ['JOJO']
    const participants = [participant('p1')]
    const localTotal = revealDurationMs('g1', 0, mangas, [{ hasStand: false, hasDevilFruit: false, hasArmamentHaki: false, hasObservationHaki: false, hasConquerorHaki: false }], 'SWIFT')
    const serverRevealMs = localTotal * 2 // server thinks the reveal takes twice as long

    await render(
      <Harness mangas={mangas} participants={participants} serverRevealMs={serverRevealMs} />
    )

    // At the LOCAL intro duration, the scaled version hasn't reached the
    // first slot's narrator beat yet (it needs roughly double the time).
    await advance(REVEAL_INTRO_MS + 1)
    expect(readState().phase).toBe('intro')

    // The full scaled duration finishes the whole thing.
    await advance(serverRevealMs)
    expect(readState().isRevealing).toBe(false)
  })

  it('skip jumps straight to done', async () => {
    const mangas: Manga[] = ['JOJO', 'ONE_PIECE']
    const participants = [participant('p1')]
    const skipRef: { current: (() => void) | null } = { current: null }

    function SkipHarness() {
      const result = useLoadoutReveal({
        gameId: 'g1',
        roundIndex: 0,
        mangas,
        participants,
        speed: 'SWIFT',
        active: true,
        markRevealed: () => {},
        sendRevealReady: () => {},
        serverRevealMs: null,
      })
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

  it('narrator/spin/land phases advance one slot at a time, in order, for the current player', async () => {
    const mangas: Manga[] = ['JOJO']
    const participants = [participant('p1')]

    await render(<Harness mangas={mangas} participants={participants} serverRevealMs={null} />)

    await advance(REVEAL_INTRO_MS + REVEAL_PLAYER_INTRO_MS + 1)
    expect(readState()).toMatchObject({ phase: 'narrator', participantIndex: 0, slotIndex: 0 })

    // Narrator holds, then spin (1 or 2 cycles - either way, REVEAL_SPIN_BASE_MS*2 safely overshoots into it).
    await advance(REVEAL_NARRATOR_MS + REVEAL_SPIN_BASE_MS * 2 + 1)
    expect(readState()).toMatchObject({ phase: 'land', participantIndex: 0, slotIndex: 0 })
  })

  it('plays every participant in turn, never in parallel', async () => {
    const mangas: Manga[] = ['JOJO']
    const participants = [participant('p1'), participant('p2')]

    await render(<Harness mangas={mangas} participants={participants} serverRevealMs={null} />)

    const total = revealDurationMs('g1', 0, mangas, [
      { hasStand: false, hasDevilFruit: false, hasArmamentHaki: false, hasObservationHaki: false, hasConquerorHaki: false },
      { hasStand: false, hasDevilFruit: false, hasArmamentHaki: false, hasObservationHaki: false, hasConquerorHaki: false },
    ], 'SWIFT')

    // Somewhere in the middle of the timeline, exactly one participant's
    // turn should be active (never both).
    await advance(Math.floor(total / 2))
    const mid = readState()
    expect([0, 1]).toContain(mid.participantIndex)

    await advance(total)
    expect(readState().isRevealing).toBe(false)
  })
})
