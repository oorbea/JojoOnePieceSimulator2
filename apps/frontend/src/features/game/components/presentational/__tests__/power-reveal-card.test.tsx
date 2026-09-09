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
