import { currentRound, hasAllLoadouts, loadoutTraits, secondsUntil } from '@/features/game/lib/match-rules'
import type { GameLoadout, GameParticipant, GameRound, GameSnapshot } from '@/features/game/types/game.types'

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

describe('loadoutTraits', () => {
  it('orders traits physicalForm, armamentHaki, observationHaki, conquerorHaki, spin, hamon, fruitMastery', () => {
    const traits = loadoutTraits(
      loadout({ spin: 'BASIC', hamon: 'BASIC', fruitMastery: 'REGULAR' })
    )
    expect(traits.map((t) => t.key)).toEqual([
      'physicalForm',
      'armamentHaki',
      'observationHaki',
      'conquerorHaki',
      'spin',
      'hamon',
      'fruitMastery',
    ])
  })

  it('omits spin/hamon/fruitMastery when NONE', () => {
    const traits = loadoutTraits(loadout())
    expect(traits.map((t) => t.key)).toEqual([
      'physicalForm',
      'armamentHaki',
      'observationHaki',
      'conquerorHaki',
    ])
  })

  it('never filters haki/physicalForm even at their floor value', () => {
    const traits = loadoutTraits(loadout())
    expect(traits.find((t) => t.key === 'physicalForm')).toBeTruthy()
    expect(traits.find((t) => t.key === 'armamentHaki')).toBeTruthy()
    expect(traits.find((t) => t.key === 'observationHaki')).toBeTruthy()
    expect(traits.find((t) => t.key === 'conquerorHaki')).toBeTruthy()
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
