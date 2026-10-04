import { renderWithProviders, screen } from '@/test/render'

import { EffectLevelUp } from '../effect-level-up'

function props(overrides: Partial<Parameters<typeof EffectLevelUp>[0]> = {}) {
  return {
    statLabel: 'Spin',
    fromLabel: 'None',
    toLabel: 'Basic',
    raised: false,
    stamp: 'LEVELLED UP!',
    reducedMotion: true,
    ...overrides,
  }
}

describe('EffectLevelUp', () => {
  it('shows the drawn value alone before the stat rises', async () => {
    await renderWithProviders(<EffectLevelUp {...props()} />)

    expect(screen.getByText('Spin')).toBeTruthy()
    expect(screen.getByText('None')).toBeTruthy()
    expect(screen.queryByText('Basic')).toBeNull()
    expect(screen.queryByText('LEVELLED UP!')).toBeNull()
  })

  it('shows the new value and the stamp once raised, keeping the drawn one beside it', async () => {
    await renderWithProviders(<EffectLevelUp {...props({ raised: true })} />)

    expect(screen.getByText('None')).toBeTruthy()
    expect(screen.getByText('Basic')).toBeTruthy()
    expect(screen.getByText('LEVELLED UP!')).toBeTruthy()
  })
})
