import { revealOrder, revealStepMs, shouldReveal } from '@/features/game/lib/loadout-reveal'
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

function loadout(): GameLoadout {
  return {
    spin: 'NONE',
    hamon: 'NONE',
    fruitMastery: 'NONE',
    armamentHaki: 'PRIVATE',
    observationHaki: 'PRIVATE',
    conquerorHaki: 'PRIVATE',
    physicalForm: 'PRIVATE',
  }
}

function live(overrides: Partial<LiveMatchState> = {}): LiveMatchState {
  return {
    assignmentSeq: 0,
    revealedAssignmentSeq: 0,
    assignedRoundIndex: null,
    votingRoundIndex: null,
    votingClosesAt: null,
    tiebreak: false,
    ...overrides,
  }
}

describe('revealOrder', () => {
  it('GAUNTLET: uses roster order with self moved last', () => {
    const snap = snapshot({
      mode: 'GAUNTLET',
      participants: [
        participant({ id: 'p1' }),
        participant({ id: 'p2' }),
        participant({ id: 'p3' }),
      ],
    })
    expect(revealOrder(snap, 'p2')).toEqual(['p1', 'p3', 'p2'])
  })

  it('VERSUS: groups by team order with self moved last', () => {
    const snap = snapshot({
      mode: 'VERSUS',
      teams: [
        { id: 'tA', name: 'A', color: 0, memberIds: ['p1', 'p2'] },
        { id: 'tB', name: 'B', color: 0, memberIds: ['p3', 'p4'] },
      ],
      participants: [
        participant({ id: 'p1', teamId: 'tA' }),
        participant({ id: 'p2', teamId: 'tA' }),
        participant({ id: 'p3', teamId: 'tB' }),
        participant({ id: 'p4', teamId: 'tB' }),
      ],
    })
    expect(revealOrder(snap, 'p3')).toEqual(['p1', 'p2', 'p4', 'p3'])
  })
})

describe('shouldReveal', () => {
  const snap = snapshot({
    participants: [
      participant({ id: 'p1', loadout: loadout() }),
      participant({ id: 'p2', loadout: loadout() }),
    ],
    rounds: [{ index: 0, stage: {} as never, options: [], tiebreakUsed: false, votedParticipantIds: [] }],
  })

  it('is false when assignmentSeq equals revealedAssignmentSeq', () => {
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 1, assignedRoundIndex: 0 }), snap)).toBe(
      false
    )
  })

  it('is false when the frame arrived before the snapshot caught up (race)', () => {
    const staleSnap = snapshot({
      participants: [participant({ id: 'p1', loadout: loadout() }), participant({ id: 'p2' })],
      rounds: [{ index: 0, stage: {} as never, options: [], tiebreakUsed: false, votedParticipantIds: [] }],
    })
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0, assignedRoundIndex: 0 }), staleSnap)).toBe(
      false
    )
  })

  it('is false when the current round index does not match the assigned round', () => {
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0, assignedRoundIndex: 1 }), snap)).toBe(
      false
    )
  })

  it('is true when a new assignment landed, all loadouts are present, and the round matches', () => {
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0, assignedRoundIndex: 0 }), snap)).toBe(
      true
    )
  })

  it('is unaffected by reduced motion (a purely gating function)', () => {
    // shouldReveal takes no reducedMotion param - verifying it stays true
    // regardless of what the caller intends to do with the result.
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0, assignedRoundIndex: 0 }), snap)).toBe(
      true
    )
  })
})

describe('revealStepMs', () => {
  it('is zero under reduced motion', () => {
    expect(revealStepMs(true)).toBe(0)
  })

  it('is a positive delay otherwise', () => {
    expect(revealStepMs(false)).toBeGreaterThan(0)
  })
})
