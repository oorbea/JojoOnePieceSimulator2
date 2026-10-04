import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'
import type { StandResponse } from '@/features/stands/types/stands.types'
import {
  displayLoadoutAt,
  effectAnchors,
  effectSteps,
  effectTargetSlot,
  revealPlayerFor,
  type LoadoutEffect,
} from '../power-effects'
import type { GameLoadout } from '@/features/game/types/game.types'

function stand(id: string, evolvesFrom: StandResponse | null = null): StandResponse {
  return { id, name: id, evolvesFrom } as StandResponse
}

function fruit(id: string, evolvesFrom: DevilFruitResponse | null = null): DevilFruitResponse {
  return { id, name: id, evolvesFrom } as DevilFruitResponse
}

function loadout(overrides: Partial<GameLoadout> = {}): GameLoadout {
  return {
    spin: 'NONE',
    hamon: 'NONE',
    fruitMastery: 'NONE',
    armamentHaki: 'NONE',
    observationHaki: 'NONE',
    conquerorHaki: 'NONE',
    physicalForm: 'PRIVATE',
    effects: [],
    ...overrides,
  }
}

function floor(
  slot: LoadoutEffect['slot'],
  from: string,
  to: string,
  causeSlot: LoadoutEffect['causeSlot'],
  cause = 'x'
): LoadoutEffect {
  return { kind: 'STAT_FLOOR', slot, from, to, causeSlot, cause }
}

function evolution(
  slot: 'STAND' | 'DEVIL_FRUIT',
  from: string,
  to: string,
  causeSlot: LoadoutEffect['causeSlot'],
  cause: string
): LoadoutEffect {
  return { kind: 'EVOLUTION', slot, from, to, causeSlot, cause }
}

const acto1 = stand('acto1')
const acto2 = stand('acto2', acto1)
const acto3 = stand('acto3', acto2)
const acto4 = stand('acto4', acto3)
const gomu = fruit('gomu')
const nika = fruit('nika', gomu)

describe('effectTargetSlot', () => {
  it('is the slot the effect changes', () => {
    expect(effectTargetSlot(floor('SPIN', 'NONE', 'BASIC', 'STAND'))).toBe('spin')
    expect(effectTargetSlot(evolution('STAND', 'a', 'b', 'SPIN', 'INFINITE'))).toBe('stand')
  })

  it('puts a haki type granted from nothing on the haki set, which exists in the drawn loadout', () => {
    expect(effectTargetSlot(floor('ARMAMENT_HAKI', 'NONE', 'PRIVATE', 'HAMON'))).toBe('hakiSet')
  })

  it('keeps a raised haki level on its own slot', () => {
    expect(effectTargetSlot(floor('OBSERVATION_HAKI', 'PRIVATE', 'YONKO_COMMANDER', 'STAND'))).toBe(
      'observationHaki'
    )
  })
})

describe('effectAnchors', () => {
  it('is empty for no effects', () => {
    expect(effectAnchors([])).toEqual([])
  })

  it('anchors a spin-driven evolution at the spin, after the stand it changes', () => {
    expect(effectAnchors([evolution('STAND', 'acto1', 'acto4', 'SPIN', 'INFINITE')])).toEqual([
      'spin',
    ])
  })

  it('anchors a Zoan raising physical form at the fruit, which comes after the form', () => {
    expect(
      effectAnchors([
        floor('PHYSICAL_FORM', 'PRIVATE', 'MARINE_CAPTAIN', 'DEVIL_FRUIT', 'Some Zoan'),
      ])
    ).toEqual(['devilFruit'])
  })

  it('anchors Gomu to Nika at the mastery that triggered it', () => {
    expect(
      effectAnchors([evolution('DEVIL_FRUIT', 'gomu', 'nika', 'FRUIT_MASTERY', 'AWAKENED')])
    ).toEqual(['fruitMastery'])
  })

  it('anchors a hamon-driven spin floor at the spin, the later of the two', () => {
    expect(effectAnchors([floor('SPIN', 'NONE', 'BASIC', 'HAMON', 'PERFECT')])).toEqual(['spin'])
  })

  it('anchors a granted haki type at the haki set, after the hamon that triggered it', () => {
    expect(effectAnchors([floor('ARMAMENT_HAKI', 'NONE', 'PRIVATE', 'HAMON', 'PERFECT')])).toEqual([
      'hakiSet',
    ])
  })

  it('never lets an anchor move earlier than the effect before it, so applied effects are a prefix', () => {
    const effects = [
      floor('SPIN', 'NONE', 'BASIC', 'HAMON', 'PERFECT'), // spin
      floor('PHYSICAL_FORM', 'PRIVATE', 'MARINE_CAPTAIN', 'DEVIL_FRUIT', 'Zoan'), // would be devilFruit
    ]
    expect(effectAnchors(effects)).toEqual(['spin', 'spin'])
  })

  it('anchors the Nika cascade in list order', () => {
    const effects = [
      evolution('DEVIL_FRUIT', 'gomu', 'nika', 'FRUIT_MASTERY', 'AWAKENED'),
      floor('PHYSICAL_FORM', 'PRIVATE', 'MARINE_CAPTAIN', 'DEVIL_FRUIT', 'nika'),
      floor('HAMON', 'NONE', 'ADVANCED', 'DEVIL_FRUIT', 'nika'),
    ]
    expect(effectAnchors(effects)).toEqual(['fruitMastery', 'fruitMastery', 'hamon'])
  })
})

describe('displayLoadoutAt', () => {
  const effects = [
    evolution('STAND', 'acto2', 'acto4', 'SPIN', 'INFINITE'),
    floor('PHYSICAL_FORM', 'PRIVATE', 'MARINE_CAPTAIN', 'DEVIL_FRUIT', 'Zoan'),
    floor('HAMON', 'NONE', 'ADVANCED', 'DEVIL_FRUIT', 'nika'),
  ]
  const final = loadout({
    stand: acto4,
    spin: 'INFINITE',
    physicalForm: 'MARINE_CAPTAIN',
    hamon: 'ADVANCED',
    effects,
  })

  it('returns the loadout itself once every effect applied', () => {
    expect(displayLoadoutAt(final, 3)).toBe(final)
    expect(displayLoadoutAt(final, 99)).toBe(final)
  })

  it('shows the drawn values before any effect', () => {
    const drawn = displayLoadoutAt(final, 0)
    expect(drawn.stand?.id).toBe('acto2')
    expect(drawn.physicalForm).toBe('PRIVATE')
    expect(drawn.hamon).toBe('NONE')
    expect(drawn.spin).toBe('INFINITE') // spin was drawn that way; nothing changed it
  })

  it('applies the effects in list order', () => {
    expect(displayLoadoutAt(final, 1).stand?.id).toBe('acto4')
    expect(displayLoadoutAt(final, 1).physicalForm).toBe('PRIVATE')
    expect(displayLoadoutAt(final, 2).physicalForm).toBe('MARINE_CAPTAIN')
    expect(displayLoadoutAt(final, 2).hamon).toBe('NONE')
  })

  it("undoes two effects on one stat back to the first one's from", () => {
    const l = loadout({
      hamon: 'ADVANCED',
      effects: [
        floor('HAMON', 'NONE', 'BASIC', 'STAND'),
        floor('HAMON', 'BASIC', 'ADVANCED', 'DEVIL_FRUIT'),
      ],
    })
    expect(displayLoadoutAt(l, 0).hamon).toBe('NONE')
    expect(displayLoadoutAt(l, 1).hamon).toBe('BASIC')
  })

  it('walks a fruit evolution back along its chain', () => {
    const l = loadout({
      devilFruit: nika,
      effects: [evolution('DEVIL_FRUIT', 'gomu', 'nika', 'FRUIT_MASTERY', 'AWAKENED')],
    })
    expect(displayLoadoutAt(l, 0).devilFruit?.id).toBe('gomu')
    expect(displayLoadoutAt(l, 1).devilFruit?.id).toBe('nika')
  })

  it('does not mutate the loadout it was given', () => {
    displayLoadoutAt(final, 0)
    expect(final.stand?.id).toBe('acto4')
    expect(final.hamon).toBe('ADVANCED')
  })

  it('tolerates a loadout with no effects field (a snapshot from before they existed)', () => {
    const legacy = { ...loadout(), effects: undefined } as unknown as GameLoadout
    expect(displayLoadoutAt(legacy, 0)).toBe(legacy)
  })
})

describe('effectSteps', () => {
  it('is 0 for a stat floor', () => {
    expect(effectSteps(loadout(), floor('SPIN', 'NONE', 'BASIC', 'STAND'))).toBe(0)
  })

  it('counts the stages an evolution spans on the final stand chain', () => {
    const l = loadout({ stand: acto4 })
    expect(effectSteps(l, evolution('STAND', 'acto1', 'acto4', 'SPIN', 'INFINITE'))).toBe(3)
    expect(effectSteps(l, evolution('STAND', 'acto2', 'acto4', 'SPIN', 'INFINITE'))).toBe(2)
    expect(effectSteps(l, evolution('STAND', 'acto3', 'acto4', 'SPIN', 'INFINITE'))).toBe(1)
  })

  it('counts fruit stages the same way', () => {
    expect(
      effectSteps(
        loadout({ devilFruit: nika }),
        evolution('DEVIL_FRUIT', 'gomu', 'nika', 'FRUIT_MASTERY', 'AWAKENED')
      )
    ).toBe(1)
  })

  it('plays one stage when it cannot find the ends on the chain', () => {
    expect(
      effectSteps(
        loadout({ stand: acto4 }),
        evolution('STAND', 'nope', 'acto4', 'SPIN', 'INFINITE')
      )
    ).toBe(1)
  })
})

describe('revealPlayerFor', () => {
  it('lands nothing for no loadout', () => {
    expect(revealPlayerFor(null)).toMatchObject({
      hasStand: false,
      hasDevilFruit: false,
      hasArmamentHaki: false,
      standEvolutionSteps: 0,
      fruitEvolutionSteps: 0,
      effects: [],
    })
  })

  it('describes the DRAWN loadout, mirroring the backend RevealPlayerFor', () => {
    const effects = [
      evolution('STAND', 'acto2', 'acto4', 'SPIN', 'INFINITE'),
      floor('ARMAMENT_HAKI', 'NONE', 'PRIVATE', 'HAMON', 'PERFECT'),
      floor('OBSERVATION_HAKI', 'PRIVATE', 'YONKO_COMMANDER', 'STAND', 'King Crimson'),
    ]
    const l = loadout({
      stand: acto4,
      devilFruit: nika,
      spin: 'INFINITE',
      fruitMastery: 'AWAKENED',
      hamon: 'PERFECT',
      armamentHaki: 'PRIVATE',
      observationHaki: 'YONKO_COMMANDER',
      effects,
    })
    expect(revealPlayerFor(l)).toEqual({
      hasStand: true,
      hasDevilFruit: true,
      hasArmamentHaki: false, // granted by an effect: no level slot of its own
      hasObservationHaki: true,
      hasConquerorHaki: false,
      standEvolutionSteps: 1, // drawn Acto 2
      fruitEvolutionSteps: 1, // Nika drawn directly plays from Gomu
      effects: [
        { kind: 'EVOLUTION', steps: 2, anchor: 'spin' },
        { kind: 'STAT_FLOOR', steps: 0, anchor: 'spin' },
        { kind: 'STAT_FLOOR', steps: 0, anchor: 'spin' },
      ],
    })
  })
})
