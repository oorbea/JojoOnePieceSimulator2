import {
  REVEAL_HOLD_BLOCK_MS,
  REVEAL_HOLD_SCALAR_MS,
  REVEAL_INTRO_MS,
  REVEAL_OUTRO_MS,
  REVEAL_SPIN_MS,
  revealDurationMs,
  revealPhases,
  shouldReveal,
} from '@/features/game/lib/loadout-reveal'
import type { LiveMatchState } from '@/features/game/stores/game-socket.store'
import type { GameLoadout, GameParticipant, GameSnapshot } from '@/features/game/types/game.types'

function participant(overrides: Partial<GameParticipant> = {}): GameParticipant {
  return {
    id: 'p1',
    displayName: 'Alice',
    teamId: 't1',
    kind: 'HUMAN',
    connected: true,
    ...overrides,
  }
}

function snapshot(overrides: Partial<GameSnapshot> = {}): GameSnapshot {
  return {
    id: 'g1',
    code: 'ABC123',
    state: 'VOTING',
    mode: 'GAUNTLET',
    hostId: 'p1',
    locked: false,
    config: {
      mangas: ['ONE_PIECE'],
      abilitySource: 'RANDOM',
      teamSize: 4,
      allowBots: false,
      visibility: 'PRIVATE',
      votingWindowSeconds: 30,
      poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
    },
    teams: [],
    participants: [],
    rounds: [],
    ...overrides,
  }
}

function loadout(overrides: Partial<GameLoadout> = {}): GameLoadout {
  return {
    spin: 'NONE',
    hamon: 'NONE',
    fruitMastery: 'NONE',
    armamentHaki: 'PRIVATE',
    observationHaki: 'PRIVATE',
    conquerorHaki: 'PRIVATE',
    physicalForm: 'PRIVATE',
    ...overrides,
  }
}

function live(overrides: Partial<LiveMatchState> = {}): LiveMatchState {
  return {
    assignmentSeq: 0,
    revealedAssignmentSeq: 0,
    assignedRoundIndex: null,
    revealMs: null,
    revealEndsAt: null,
    votingRoundIndex: null,
    votingClosesAt: null,
    tiebreak: false,
    ...overrides,
  }
}

describe('shouldReveal', () => {
  // No rounds exist yet - a Game sits in ASSIGNING with zero Rounds for the
  // whole sorteo delay (see game.RevealDuration/GameService.
  // scheduleRevealDelay), since OpenVoting (which creates the Round) only
  // runs once the delay elapses. shouldReveal must not depend on a Round
  // existing - see its doc for the bug this replaced.
  const snap = snapshot({
    state: 'ASSIGNING',
    participants: [
      participant({ id: 'p1', loadout: loadout() }),
      participant({ id: 'p2', loadout: loadout() }),
    ],
    rounds: [],
  })

  it('is false when assignmentSeq equals revealedAssignmentSeq', () => {
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 1 }), snap)).toBe(false)
  })

  it('is false when the frame arrived before the snapshot caught up (race)', () => {
    const staleSnap = snapshot({
      state: 'ASSIGNING',
      participants: [participant({ id: 'p1', loadout: loadout() }), participant({ id: 'p2' })],
      rounds: [],
    })
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0 }), staleSnap)).toBe(false)
  })

  it('is false once voting has actually opened, even if this assignment was never revealed', () => {
    const votingSnap = snapshot({
      ...snap,
      state: 'VOTING',
      rounds: [{ index: 0, stage: {} as never, options: [], tiebreakUsed: false, votedParticipantIds: [] }],
    })
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0 }), votingSnap)).toBe(false)
  })

  it('is true when a new assignment landed, all loadouts are present, and the game is still ASSIGNING', () => {
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0 }), snap)).toBe(true)
  })
})

describe('revealPhases', () => {
  it('both mangas: intro, 9 slots (spin+land each), outro', () => {
    const phases = revealPhases(['JOJO', 'ONE_PIECE'])
    expect(phases[0].phase).toEqual({ kind: 'intro' })
    expect(phases[0].durationMs).toBe(REVEAL_INTRO_MS)
    expect(phases[phases.length - 1].phase).toEqual({ kind: 'outro' })
    expect(phases[phases.length - 1].durationMs).toBe(REVEAL_OUTRO_MS)
    // intro + 9 * (spin, land) + outro
    expect(phases).toHaveLength(1 + 9 * 2 + 1)
  })

  it('stand/devilFruit slots hold longer than scalar slots (they carry art + a stat grid)', () => {
    const phases = revealPhases(['JOJO', 'ONE_PIECE'])
    const standSlotIndex = 1 // physicalForm(0), stand(1)
    const standLand = phases.find((p) => p.phase.kind === 'land' && p.phase.slot === standSlotIndex)
    expect(standLand?.durationMs).toBe(REVEAL_HOLD_BLOCK_MS)
    const physicalFormLand = phases.find((p) => p.phase.kind === 'land' && p.phase.slot === 0)
    expect(physicalFormLand?.durationMs).toBe(REVEAL_HOLD_SCALAR_MS)
  })

  it('every spin phase holds REVEAL_SPIN_MS regardless of slot kind', () => {
    const phases = revealPhases(['JOJO', 'ONE_PIECE'])
    const spins = phases.filter((p) => p.phase.kind === 'spin')
    expect(spins).toHaveLength(9)
    spins.forEach((s) => expect(s.durationMs).toBe(REVEAL_SPIN_MS))
  })
})

// revealDurationMs is the frontend half of the backend/frontend agreement
// pinned by TestRevealDuration_PinnedTotals (reveal_test.go) - keep both in
// sync if either side's constants or slot list ever change.
describe('revealDurationMs', () => {
  it('both mangas: 44750ms', () => {
    expect(revealDurationMs(['JOJO', 'ONE_PIECE'])).toBe(44750)
  })

  it('jojo only: 18350ms', () => {
    expect(revealDurationMs(['JOJO'])).toBe(18350)
  })

  it('one piece only: 30800ms', () => {
    expect(revealDurationMs(['ONE_PIECE'])).toBe(30800)
  })
})
