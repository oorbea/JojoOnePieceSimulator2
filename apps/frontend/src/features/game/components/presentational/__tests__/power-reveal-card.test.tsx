// Its own file: PowerRevealCard renders a real RN <Modal>, which drives
// real timers and corrupts other renders sharing the file (see
// frontend-stack.md). Asserts the responsive fix's actual contract - every
// skill's full text renders, nothing is truncated to a single-line pill -
// and that Skip stays reachable. Geometry itself lives in
// reveal-layout.test.ts; RNTL doesn't render real layout so it can't prove
// sizing.
import { fireEvent, renderWithProviders, screen } from '@/test/render'
import type { StandResponse } from '@/shared/contracts/dto'

import { PowerRevealCard } from '../match/power-reveal-card'

function stand(overrides: Partial<StandResponse> = {}): StandResponse {
  return {
    id: 'stand-1',
    name: 'Crazy Diamond',
    description: 'Repairs anything by returning it to a previous, undamaged state.',
    rarity: 'RARE',
    skills: [
      'Restoration: Repairs any damaged object or living being back to a prior working state.',
      'Rapid Punches: A relentless barrage of close-range strikes.',
    ],
    picture: '',
    pictureThumb: '',
    pictureCard: '',
    pictureStatus: 'READY',
    pictureLqip: '',
    focalX: 0.5,
    focalY: 0.5,
    attackPower: 'A',
    speed: 'A',
    attackRange: 'C',
    endurance: 'B',
    precision: 'A',
    potential: 'A',
    evolvesFrom: null,
    ...overrides,
  }
}

describe('PowerRevealCard', () => {
  it('renders each skill in full, not truncated to a single-line pill', async () => {
    await renderWithProviders(
      <PowerRevealCard
        visible
        kind="stand"
        stand={stand()}
        participantName="jotaro"
        onSkip={jest.fn()}
      />
    )

    expect(
      screen.getByText(
        'Restoration: Repairs any damaged object or living being back to a prior working state.'
      )
    ).toBeTruthy()
    expect(
      screen.getByText('Rapid Punches: A relentless barrage of close-range strikes.')
    ).toBeTruthy()
  })

  // 2026-09-25 "algo está pasando" evolution reveal.
  it('shows the evolving message instead of the description/skills during evolvePhase="evolving"', async () => {
    await renderWithProviders(
      <PowerRevealCard
        visible
        kind="stand"
        stand={stand({ name: 'Silver Chariot' })}
        participantName="jotaro"
        onSkip={jest.fn()}
        evolvePhase="evolving"
        evolveMessage="Algo está pasando... ¡tu stand está evolucionando!"
      />
    )

    // MangaVerdictText stacks several copies of its own text on native (the
    // comic-book outline trick, see its own doc) - getAllByText, not
    // getByText, since RNTL's getByText throws on more than one match.
    expect(
      screen.getAllByText('Algo está pasando... ¡tu stand está evolucionando!').length
    ).toBeGreaterThan(0)
    expect(screen.getByText('Silver Chariot')).toBeTruthy()
    expect(
      screen.queryByText(
        'Restoration: Repairs any damaged object or living being back to a prior working state.'
      )
    ).toBeNull()
  })

  it('shows the description/skills again once landed, with the evolution stamp for the final stage', async () => {
    await renderWithProviders(
      <PowerRevealCard
        visible
        kind="stand"
        stand={stand({ name: 'Chariot Requiem' })}
        participantName="jotaro"
        onSkip={jest.fn()}
        evolvePhase="final"
      />
    )

    expect(screen.getByText('Chariot Requiem')).toBeTruthy()
    expect(
      screen.getByText(
        'Restoration: Repairs any damaged object or living being back to a prior working state.'
      )
    ).toBeTruthy()
    expect(screen.getByText('EVOLUTION!')).toBeTruthy()
  })

  it('keeps the Skip button present and reachable', async () => {
    const onSkip = jest.fn()
    await renderWithProviders(
      <PowerRevealCard
        visible
        kind="stand"
        stand={stand()}
        participantName="jotaro"
        onSkip={onSkip}
      />
    )

    fireEvent.press(screen.getByText('Skip'))
    expect(onSkip).toHaveBeenCalledTimes(1)
  })
})
