import {
  playerSlots,
  REVEAL_HOLD_EMPTY_MS,
  REVEAL_HOLD_FRUIT_MS,
  REVEAL_HOLD_SCALAR_MS,
  REVEAL_HOLD_STAND_MS,
  REVEAL_INTRO_MS,
  REVEAL_OUTRO_MS,
  REVEAL_PLAYER_INTRO_MS,
  revealDurationMs,
  revealSpinCycles,
  revealTimeline,
  shouldReveal,
  type RevealPlayer,
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
    avatarThumb: '',
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
      stageMangas: ['ONE_PIECE'],
      powerMangas: ['ONE_PIECE'],
      abilitySource: 'RANDOM',
      teamSize: 4,
      allowBots: false,
      visibility: 'PRIVATE',
      votingWindowSeconds: 30,
      revealSpeed: 'NORMAL',
    summaryDurationSeconds: 60,
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
    revealReadyCount: null,
    revealReadyTotal: null,
    summaryEndsAt: null,
    summaryReadyCount: null,
    summaryReadyTotal: null,
    votingRoundIndex: null,
    votingClosesAt: null,
    tiebreak: false,
    votesCast: null,
    voters: null,
    resultEndsAt: null,
    resultDismissed: false,
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
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0 }), staleSnap)).toBe(
      false
    )
  })

  it('is false once voting has actually opened, even if this assignment was never revealed', () => {
    const votingSnap = snapshot({
      ...snap,
      state: 'VOTING',
      rounds: [
        { index: 0, stage: {} as never, options: [], tiebreakUsed: false, votedParticipantIds: [] },
      ],
    })
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0 }), votingSnap)).toBe(
      false
    )
  })

  it('is true when a new assignment landed, all loadouts are present, and the game is still ASSIGNING', () => {
    expect(shouldReveal(live({ assignmentSeq: 1, revealedAssignmentSeq: 0 }), snap)).toBe(true)
  })
})

describe('revealSpinCycles', () => {
  it('always returns 1 or 2 (mirrors V1s generateRandomNumber(1, 2))', () => {
    for (let slot = 0; slot < 10; slot++) {
      for (let pi = 0; pi < 5; pi++) {
        const cycles = revealSpinCycles('game-1', 0, pi, slot)
        expect([1, 2]).toContain(cycles)
      }
    }
  })

  it('is deterministic - same inputs always produce the same result', () => {
    expect(revealSpinCycles('game-1', 2, 3, 4)).toBe(revealSpinCycles('game-1', 2, 3, 4))
  })
})

const NO_POWERS: RevealPlayer = {
  hasStand: false,
  hasDevilFruit: false,
  hasArmamentHaki: false,
  hasObservationHaki: false,
  hasConquerorHaki: false,
}
const BOTH_POWERS: RevealPlayer = {
  ...NO_POWERS,
  hasStand: true,
  hasDevilFruit: true,
}
const ALL_HAKI: RevealPlayer = {
  ...NO_POWERS,
  hasArmamentHaki: true,
  hasObservationHaki: true,
  hasConquerorHaki: true,
}

describe('playerSlots', () => {
  it('excludes a haki-level slot for a type the player does not have', () => {
    const onlyObservation: RevealPlayer = { ...NO_POWERS, hasObservationHaki: true }
    const slots = playerSlots(['ONE_PIECE'], onlyObservation)
    expect(slots).toContain('observationHaki')
    expect(slots).not.toContain('armamentHaki')
    expect(slots).not.toContain('conquerorHaki')
    // The synthetic "which types" summary is unaffected.
    expect(slots).toContain('hakiSet')
  })

  it('includes every haki-level slot when the player has all three types', () => {
    const slots = playerSlots(['ONE_PIECE'], ALL_HAKI)
    expect(slots).toEqual(
      expect.arrayContaining(['armamentHaki', 'observationHaki', 'conquerorHaki'])
    )
  })
})

describe('revealTimeline', () => {
  it('one player with every power/haki type, both mangas: intro, playerIntro, 10x(narrator+spin+land), playerOutro, outro', () => {
    const timeline = revealTimeline('g1', 0, ['JOJO', 'ONE_PIECE'], [{ ...ALL_HAKI, hasStand: true, hasDevilFruit: true }], 'SWIFT')
    expect(timeline[0].phase).toEqual({ kind: 'intro' })
    expect(timeline[0].durationMs).toBe(REVEAL_INTRO_MS)
    expect(timeline[1].phase).toEqual({ kind: 'playerIntro', participant: 0 })
    expect(timeline[1].durationMs).toBe(REVEAL_PLAYER_INTRO_MS)
    const last = timeline[timeline.length - 1]
    expect(last.phase).toEqual({ kind: 'outro' })
    expect(last.durationMs).toBe(REVEAL_OUTRO_MS)
    // intro + playerIntro + 10 * (narrator, spin, land) + playerOutro + outro
    expect(timeline).toHaveLength(1 + 1 + 10 * 3 + 1 + 1)
  })

  it('a player with no haki at all skips all three haki-level slots (7, not 10, for both mangas)', () => {
    const timeline = revealTimeline('g1', 0, ['JOJO', 'ONE_PIECE'], [NO_POWERS], 'SWIFT')
    // intro + playerIntro + 7 * (narrator, spin, land) + playerOutro + outro
    expect(timeline).toHaveLength(1 + 1 + 7 * 3 + 1 + 1)
  })

  it('stand/devilFruit hold longer when they actually land a power than when they land NONE', () => {
    const withPowers = revealTimeline('g1', 0, ['JOJO', 'ONE_PIECE'], [BOTH_POWERS], 'SWIFT')
    const withoutPowers = revealTimeline('g1', 0, ['JOJO', 'ONE_PIECE'], [NO_POWERS], 'SWIFT')

    const standLandWith = withPowers.find((p) => p.phase.kind === 'land' && p.phase.slot === 1)
    const standLandWithout = withoutPowers.find((p) => p.phase.kind === 'land' && p.phase.slot === 1)
    expect(standLandWith?.durationMs).toBe(REVEAL_HOLD_STAND_MS)
    expect(standLandWithout?.durationMs).toBe(REVEAL_HOLD_EMPTY_MS)

    const fruitLandWith = withPowers.find((p) => p.phase.kind === 'land' && p.phase.slot === 2)
    expect(fruitLandWith?.durationMs).toBe(REVEAL_HOLD_FRUIT_MS)

    const physicalFormLand = withPowers.find((p) => p.phase.kind === 'land' && p.phase.slot === 0)
    expect(physicalFormLand?.durationMs).toBe(REVEAL_HOLD_SCALAR_MS)
  })

  it('two players each play a full turn - never in parallel', () => {
    const timeline = revealTimeline('g1', 0, ['JOJO'], [NO_POWERS, NO_POWERS], 'SWIFT')
    const playerIntros = timeline.filter((p) => p.phase.kind === 'playerIntro')
    expect(playerIntros.map((p) => p.phase.participant)).toEqual([0, 1])
    const playerOutros = timeline.filter((p) => p.phase.kind === 'playerOutro')
    expect(playerOutros).toHaveLength(2)
  })

  it('scales every duration by the speed multiplier', () => {
    const swift = revealTimeline('g1', 0, ['JOJO'], [NO_POWERS], 'SWIFT')
    const relaxed = revealTimeline('g1', 0, ['JOJO'], [NO_POWERS], 'RELAXED')
    expect(relaxed[0].durationMs).toBeGreaterThan(swift[0].durationMs)
  })
})

describe('revealDurationMs', () => {
  it('is longer with more players', () => {
    const one = revealDurationMs('g1', 0, ['JOJO'], [NO_POWERS], 'SWIFT')
    const two = revealDurationMs('g1', 0, ['JOJO'], [NO_POWERS, NO_POWERS], 'SWIFT')
    expect(two).toBeGreaterThan(one)
  })

  it('agrees exactly with the sum of revealTimeline durations', () => {
    const timeline = revealTimeline('g1', 3, ['JOJO', 'ONE_PIECE'], [BOTH_POWERS, NO_POWERS], 'NORMAL')
    const total = timeline.reduce((sum, p) => sum + p.durationMs, 0)
    expect(revealDurationMs('g1', 3, ['JOJO', 'ONE_PIECE'], [BOTH_POWERS, NO_POWERS], 'NORMAL')).toBe(total)
  })

  it('scales with speed: RELAXED > NORMAL > SWIFT', () => {
    const players = [BOTH_POWERS]
    const mangas = ['JOJO', 'ONE_PIECE'] as const
    const swift = revealDurationMs('g1', 0, [...mangas], players, 'SWIFT')
    const normal = revealDurationMs('g1', 0, [...mangas], players, 'NORMAL')
    const relaxed = revealDurationMs('g1', 0, [...mangas], players, 'RELAXED')
    expect(swift).toBeLessThan(normal)
    expect(normal).toBeLessThan(relaxed)
  })
})
