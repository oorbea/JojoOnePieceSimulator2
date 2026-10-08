// Its own file for the same reason power-reveal-card.test.tsx is: an
// evolution effect puts a real RN <Modal> on screen, which drives real timers
// and corrupts other renders sharing the file (see frontend-stack.md).
//
// Covers what RevealStage draws during a power effect's beat (the sorteo's
// "this power just changed something" moment): the cause line, the stat's drawn
// value before it rises and the new one after, and - for a Stand/DevilFruit
// evolution - the card, walking from the drawn stage. The timeline that decides
// WHEN each phase plays is covered by loadout-reveal.test.ts.
import { renderWithProviders, screen } from '@/test/render'
import type { DevilFruitResponse, StandResponse } from '@/shared/contracts/dto'

import { RevealStage } from '../match/reveal-stage'

// Same reason as power-reveal-card.test.tsx: the evolution landing sound pulls
// in expo-audio's native module, unavailable under RNTL/jest.
jest.mock('@/shared/hooks/use-sound', () => ({
  useSound: () => ({ play: jest.fn(), stop: jest.fn() }),
}))
// RevealStage reads the catalogue only to build the case strip's decoys; none
// is rendered during an effect, so an empty one is enough.
jest.mock('@/features/stands', () => ({ useStands: () => ({ data: [] }) }))
jest.mock('@/features/devil-fruits', () => ({ useDevilFruits: () => ({ data: [] }) }))

function stand(id: string, name: string, evolvesFrom: StandResponse | null = null): StandResponse {
  return {
    id,
    name,
    description: `${name} description`,
    rarity: 'EPIC',
    skills: [],
    picture: '',
    pictureThumb: '',
    pictureCard: '',
    pictureStatus: 'READY',
    pictureLqip: '',
    focalX: 0.5,
    focalY: 0.5,
    attackPower: 'A',
    speed: 'A',
    attackRange: 'A',
    endurance: 'A',
    precision: 'A',
    potential: 'A',
    evolvesFrom,
  }
}

function fruit(
  id: string,
  name: string,
  evolvesFrom: DevilFruitResponse | null = null
): DevilFruitResponse {
  return {
    id,
    name,
    description: `${name} description`,
    rarity: 'LEGENDARY',
    skills: [],
    picture: '',
    pictureThumb: '',
    pictureCard: '',
    pictureStatus: 'READY',
    pictureLqip: '',
    focalX: 0.5,
    focalY: 0.5,
    fruitType: 'MYTHICAL_ZOAN',
    evolvesFrom,
  }
}

const acto1 = stand('acto1', 'Tusk: Act 1')
const acto2 = stand('acto2', 'Tusk: Act 2', acto1)
const acto4 = stand('acto4', 'Tusk: Act 4', stand('acto3', 'Tusk: Act 3', acto2))

function snapshotWith(loadout: Record<string, unknown>) {
  return {
    id: 'g1',
    mode: 'GAUNTLET',
    rounds: [],
    config: {
      powerMangas: ['JOJO'],
      revealSpeed: 'SWIFT',
      poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
    },
    participants: [
      {
        id: 'p1',
        displayName: 'Jotaro',
        teamId: 't1',
        kind: 'HUMAN',
        connected: true,
        abandoned: false,
        avatarThumb: '',
        avatarFocalX: 0.5,
        avatarFocalY: 0.5,
        loadout,
      },
    ],
  } as never
}

const BASE_LOADOUT = {
  spin: 'NONE',
  hamon: 'PERFECT',
  fruitMastery: 'NONE',
  armamentHaki: 'NONE',
  observationHaki: 'NONE',
  conquerorHaki: 'NONE',
  physicalForm: 'PRIVATE',
  effects: [],
}

function stage(
  snapshot: never,
  over: Partial<Parameters<typeof RevealStage>[0]> & {
    phase: Parameters<typeof RevealStage>[0]['phase']
  }
) {
  return (
    <RevealStage
      snapshot={snapshot}
      selfId="p1"
      participantIndex={0}
      slotIndex={0}
      totalSlots={4}
      evolveStage={-1}
      effectIndex={0}
      effectsApplied={0}
      scale={1}
      readyCount={null}
      readyTotal={null}
      onSkip={jest.fn()}
      reducedMotion
      {...over}
    />
  )
}

describe('RevealStage - power effects', () => {
  const floorLoadout = {
    ...BASE_LOADOUT,
    spin: 'BASIC',
    effects: [
      {
        kind: 'STAT_FLOOR',
        slot: 'SPIN',
        from: 'NONE',
        to: 'BASIC',
        causeSlot: 'HAMON',
        cause: 'PERFECT',
      },
    ],
  }

  it('shows the drawn value and what triggered the effect while a stat floor starts', async () => {
    await renderWithProviders(stage(snapshotWith(floorLoadout), { phase: 'effectIntro' }))

    expect(screen.getByText('With Hamon at Perfect, your Spin rises to Basic!')).toBeTruthy()
    expect(screen.getByText('None')).toBeTruthy() // the drawn spin, not yet raised
    expect(screen.queryByText('LEVELLED UP!')).toBeNull()
  })

  it('shows the raised value and the stamp once the stat floor lands', async () => {
    await renderWithProviders(
      stage(snapshotWith(floorLoadout), { phase: 'effectLand', effectsApplied: 1 })
    )

    expect(screen.getByText('Basic')).toBeTruthy()
    expect(screen.getByText('LEVELLED UP!')).toBeTruthy()
  })

  it('names the power that triggered a floor when a power did', async () => {
    const loadout = {
      ...BASE_LOADOUT,
      hamon: 'BASIC',
      effects: [
        {
          kind: 'STAT_FLOOR',
          slot: 'HAMON',
          from: 'NONE',
          to: 'BASIC',
          causeSlot: 'STAND',
          cause: 'Hermit purple',
        },
      ],
    }
    await renderWithProviders(stage(snapshotWith(loadout), { phase: 'effectIntro' }))

    expect(screen.getByText('Hermit purple lifts your Hamon to Basic!')).toBeTruthy()
  })

  it('walks a Stand evolution from the drawn stage: the card shows it, with the cause line', async () => {
    const loadout = {
      ...BASE_LOADOUT,
      spin: 'INFINITE',
      stand: acto4,
      effects: [
        {
          kind: 'EVOLUTION',
          slot: 'STAND',
          from: 'acto2',
          to: 'acto4',
          causeSlot: 'SPIN',
          cause: 'INFINITE',
        },
      ],
    }
    const snapshot = snapshotWith(loadout)

    await renderWithProviders(stage(snapshot, { phase: 'effectIntro' }))
    // The stage's own narrator and the card's cause line both carry it; the card
    // (a Modal) is what covers the stage on screen.
    expect(
      screen.getAllByText('With Infinite Spin, Tusk: Act 2 evolves into Tusk: Act 4!').length
    ).toBeGreaterThan(0)
    expect(screen.getAllByText('Tusk: Act 2').length).toBeGreaterThan(0)
    expect(screen.queryByText('Tusk: Act 4')).toBeNull()
  })

  it('lands a Stand evolution on the evolved form with the evolution stamp', async () => {
    const loadout = {
      ...BASE_LOADOUT,
      spin: 'INFINITE',
      stand: acto4,
      effects: [
        {
          kind: 'EVOLUTION',
          slot: 'STAND',
          from: 'acto2',
          to: 'acto4',
          causeSlot: 'SPIN',
          cause: 'INFINITE',
        },
      ],
    }
    await renderWithProviders(
      stage(snapshotWith(loadout), { phase: 'effectLand', effectsApplied: 1 })
    )

    expect(screen.getAllByText('Tusk: Act 4').length).toBeGreaterThan(0)
    expect(screen.getByText('EVOLUTION!')).toBeTruthy()
  })
  it('walks a DevilFruit evolution the same way: Gomu Gomu no mi first, Model Nika once landed', async () => {
    const gomu = fruit('gomu', 'Gomu Gomu no mi')
    const nika = fruit('nika', 'Hito Hito no mi: Model Nika', gomu)
    const loadout = {
      ...BASE_LOADOUT,
      fruitMastery: 'AWAKENED',
      devilFruit: nika,
      effects: [
        {
          kind: 'EVOLUTION',
          slot: 'DEVIL_FRUIT',
          from: 'gomu',
          to: 'nika',
          causeSlot: 'FRUIT_MASTERY',
          cause: 'AWAKENED',
        },
      ],
    }
    const snapshot = snapshotWith(loadout)

    const intro = await renderWithProviders(stage(snapshot, { phase: 'effectIntro' }))
    expect(
      screen.getAllByText(
        'With Awakened mastery, Gomu Gomu no mi evolves into Hito Hito no mi: Model Nika!'
      ).length
    ).toBeGreaterThan(0)
    expect(screen.getAllByText('Gomu Gomu no mi').length).toBeGreaterThan(0)
    expect(screen.queryByText('Hito Hito no mi: Model Nika')).toBeNull()
    intro.unmount()

    await renderWithProviders(stage(snapshot, { phase: 'effectLand', effectsApplied: 1 }))
    expect(screen.getAllByText('Hito Hito no mi: Model Nika').length).toBeGreaterThan(0)
    expect(screen.getByText('EVOLUTION!')).toBeTruthy()
  })
})
