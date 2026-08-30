import {
  currentRound,
  hasAllLoadouts,
  loadoutSlots,
  revealSlotKinds,
  secondsUntil,
  voteProgress,
} from '@/features/game/lib/match-rules'
import type { LiveMatchState } from '@/features/game/stores/game-socket.store'
import type {
  GameLoadout,
  GameParticipant,
  GameRound,
  GameSnapshot,
} from '@/features/game/types/game.types'
import type { Manga } from '@/shared/lib/zod'

const INITIAL_LIVE: LiveMatchState = {
  assignmentSeq: 0,
  revealedAssignmentSeq: 0,
  assignedRoundIndex: null,
  revealMs: null,
  revealEndsAt: null,
  revealReadyCount: null,
  revealReadyTotal: null,
  votingRoundIndex: null,
  votingClosesAt: null,
  tiebreak: false,
  votesCast: null,
  voters: null,
  resultEndsAt: null,
  resultDismissed: false,
}

function round(overrides: Partial<GameRound> = {}): GameRound {
  return {
    index: 0,
    stage: {
      id: 's1',
      manga: 'JOJO',
      order: 0,
      name: 'Phantom Blood',
      description: '',
      picture: '',
      pictureThumb: '',
      pictureStatus: 'READY',
    },
    options: ['SURVIVE', 'FALL'],
    tiebreakUsed: false,
    votedParticipantIds: [],
    ...overrides,
  }
}

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

describe('currentRound', () => {
  it('returns null when there are no rounds', () => {
    expect(currentRound(snapshot({ rounds: [] }))).toBeNull()
  })

  it('returns the last round', () => {
    const r0 = { index: 0 } as GameRound
    const r1 = { index: 1 } as GameRound
    expect(currentRound(snapshot({ rounds: [r0, r1] }))).toBe(r1)
  })
})

describe('hasAllLoadouts', () => {
  it('is true when every connected participant has a loadout', () => {
    const snap = snapshot({
      participants: [
        participant({ id: 'p1', loadout: loadout() }),
        participant({ id: 'p2', loadout: loadout() }),
      ],
    })
    expect(hasAllLoadouts(snap)).toBe(true)
  })

  it('is false when a connected participant is missing a loadout', () => {
    const snap = snapshot({
      participants: [participant({ id: 'p1', loadout: loadout() }), participant({ id: 'p2' })],
    })
    expect(hasAllLoadouts(snap)).toBe(false)
  })

  it('ignores a disconnected participant without a loadout', () => {
    const snap = snapshot({
      participants: [
        participant({ id: 'p1', loadout: loadout() }),
        participant({ id: 'p2', connected: false }),
      ],
    })
    expect(hasAllLoadouts(snap)).toBe(true)
  })

  it('always counts a bot even without an explicit connected flag change', () => {
    const snap = snapshot({
      participants: [participant({ id: 'p1', kind: 'BOT', loadout: loadout() })],
    })
    expect(hasAllLoadouts(snap)).toBe(true)
  })
})

describe('loadoutSlots', () => {
  const BOTH: Manga[] = ['JOJO', 'ONE_PIECE']

  it('orders slots physicalForm, stand, devilFruit, fruitMastery, hamon, hakiSet, armamentHaki, observationHaki, conquerorHaki, spin', () => {
    const slots = loadoutSlots(
      loadout({ spin: 'BASIC', hamon: 'BASIC', fruitMastery: 'REGULAR' }),
      BOTH
    )
    expect(slots.map((s) => s.key)).toEqual([
      'physicalForm',
      'stand',
      'devilFruit',
      'fruitMastery',
      'hamon',
      'hakiSet',
      'armamentHaki',
      'observationHaki',
      'conquerorHaki',
      'spin',
    ])
  })

  it('includes spin/hamon/fruitMastery even when NONE - the slot count depends only on mangas, never the draw', () => {
    const slots = loadoutSlots(loadout(), BOTH)
    expect(slots.map((s) => s.key)).toEqual([
      'physicalForm',
      'stand',
      'devilFruit',
      'fruitMastery',
      'hamon',
      'hakiSet',
      'armamentHaki',
      'observationHaki',
      'conquerorHaki',
      'spin',
    ])
    expect(slots.find((s) => s.key === 'fruitMastery')?.value).toBe('NONE')
    expect(slots.find((s) => s.key === 'hamon')?.value).toBe('NONE')
    expect(slots.find((s) => s.key === 'spin')?.value).toBe('NONE')
  })

  it('never filters haki/physicalForm even at their floor value', () => {
    const slots = loadoutSlots(loadout(), BOTH)
    expect(slots.find((s) => s.key === 'physicalForm')).toBeTruthy()
    expect(slots.find((s) => s.key === 'armamentHaki')).toBeTruthy()
    expect(slots.find((s) => s.key === 'observationHaki')).toBeTruthy()
    expect(slots.find((s) => s.key === 'conquerorHaki')).toBeTruthy()
  })

  it('a JOJO-only lobby never gets a One Piece slot, even with a fully-drawn loadout', () => {
    const slots = loadoutSlots(
      loadout({
        hamon: 'BASIC',
        spin: 'BASIC',
        fruitMastery: 'REGULAR',
        physicalForm: 'VICE_ADMIRAL',
      }),
      ['JOJO']
    )
    expect(slots.map((s) => s.key)).toEqual(['stand', 'hamon', 'spin'])
  })

  it('a ONE_PIECE-only lobby never gets a JoJo slot', () => {
    const slots = loadoutSlots(loadout({ fruitMastery: 'REGULAR' }), ['ONE_PIECE'])
    expect(slots.map((s) => s.key)).toEqual([
      'physicalForm',
      'devilFruit',
      'fruitMastery',
      'hakiSet',
      'armamentHaki',
      'observationHaki',
      'conquerorHaki',
    ])
  })
})

describe('revealSlotKinds', () => {
  it('both mangas: all 10 slots in draw order, haki-set before the individual levels', () => {
    expect(revealSlotKinds(['JOJO', 'ONE_PIECE'])).toEqual([
      'physicalForm',
      'stand',
      'devilFruit',
      'fruitMastery',
      'hamon',
      'hakiSet',
      'armamentHaki',
      'observationHaki',
      'conquerorHaki',
      'spin',
    ])
  })

  it('jojo only: stand, hamon, spin', () => {
    expect(revealSlotKinds(['JOJO'])).toEqual(['stand', 'hamon', 'spin'])
  })

  it('one piece only: physicalForm, devilFruit, fruitMastery, haki-set, the three hakis', () => {
    expect(revealSlotKinds(['ONE_PIECE'])).toEqual([
      'physicalForm',
      'devilFruit',
      'fruitMastery',
      'hakiSet',
      'armamentHaki',
      'observationHaki',
      'conquerorHaki',
    ])
  })
})

describe('voteProgress', () => {
  it('prefers the live frame when it is for the current round', () => {
    const snap = snapshot({
      participants: [participant({ id: 'p1' }), participant({ id: 'p2' }), participant({ id: 'p3' })],
      rounds: [round({ index: 0 })],
    })
    const live: LiveMatchState = { ...INITIAL_LIVE, votingRoundIndex: 0, votesCast: 1, voters: 3 }
    expect(voteProgress(snap, live)).toEqual({ cast: 1, total: 3 })
  })

  it('falls back to the snapshot when no frame has landed for this round yet (a reconnect)', () => {
    const snap = snapshot({
      participants: [
        participant({ id: 'p1' }),
        participant({ id: 'p2' }),
        participant({ id: 'p3' }),
      ],
      rounds: [round({ index: 0, votedParticipantIds: ['p1'] })],
    })
    expect(voteProgress(snap, INITIAL_LIVE)).toEqual({ cast: 1, total: 3 })
  })

  it('ignores a frame for a stale round, falling back to the snapshot instead', () => {
    const snap = snapshot({
      participants: [participant({ id: 'p1' }), participant({ id: 'p2' })],
      rounds: [round({ index: 1, votedParticipantIds: ['p1'] })],
    })
    // A late VOTE_CAST from round 0, after round 1 already opened.
    const live: LiveMatchState = { ...INITIAL_LIVE, votingRoundIndex: 0, votesCast: 0, voters: 2 }
    expect(voteProgress(snap, live)).toEqual({ cast: 1, total: 2 })
  })

  it('excludes bots from both cast and total', () => {
    const snap = snapshot({
      participants: [
        participant({ id: 'p1' }),
        participant({ id: 'bot1', kind: 'BOT' }),
      ],
      rounds: [round({ index: 0, votedParticipantIds: ['p1', 'bot1'] })],
    })
    expect(voteProgress(snap, INITIAL_LIVE)).toEqual({ cast: 1, total: 1 })
  })

  it('excludes a disconnected human from both cast and total', () => {
    const snap = snapshot({
      participants: [
        participant({ id: 'p1' }),
        participant({ id: 'p2', connected: false }),
      ],
      rounds: [round({ index: 0, votedParticipantIds: ['p1'] })],
    })
    expect(voteProgress(snap, INITIAL_LIVE)).toEqual({ cast: 1, total: 1 })
  })

  it('returns 0/0 with no current round', () => {
    const snap = snapshot({ rounds: [] })
    expect(voteProgress(snap, INITIAL_LIVE)).toEqual({ cast: 0, total: 0 })
  })
})

describe('secondsUntil', () => {
  it('returns null when closesAtMs is null', () => {
    expect(secondsUntil(null, 1000)).toBeNull()
  })

  it('rounds the remaining seconds', () => {
    expect(secondsUntil(10500, 5000)).toBe(6)
  })

  it('clamps to zero once past closesAt', () => {
    expect(secondsUntil(1000, 5000)).toBe(0)
  })
})
