import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import type { StandResponse } from '@/features/stands/types/stands.types'

import { StandCard } from '../stand-card'

function baseStand(overrides: Partial<StandResponse> = {}): StandResponse {
  return {
    id: 'stand-1',
    name: 'Star Platinum',
    description: 'A powerful close-range Stand.',
    rarity: 'LEGENDARY',
    skills: ['Time Stop', 'ORA Barrage'],
    picture: '',
    pictureThumb: '',
    pictureStatus: 'NONE',
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

describe('StandCard', () => {
  it('renders the name, rarity and stats', async () => {
    await renderWithProviders(
      <StandCard stand={baseStand()} onOpenDetail={jest.fn()} onEdit={jest.fn()} onDelete={jest.fn()} />
    )

    expect(screen.getByText('Star Platinum')).toBeTruthy()
    expect(screen.getByText('Legendary')).toBeTruthy()
  })

  it('shows the evolvesFrom badge when present', async () => {
    const stand = baseStand({ evolvesFrom: baseStand({ id: 'stand-0', name: 'Star Platinum: OVER HEAVEN' }) })
    await renderWithProviders(
      <StandCard stand={stand} onOpenDetail={jest.fn()} onEdit={jest.fn()} onDelete={jest.fn()} />
    )

    expect(screen.getByText('From: Star Platinum: OVER HEAVEN')).toBeTruthy()
  })

  it('fires onEdit and onDelete from their own buttons', async () => {
    const onEdit = jest.fn()
    const onDelete = jest.fn()
    await renderWithProviders(
      <StandCard stand={baseStand()} onOpenDetail={jest.fn()} onEdit={onEdit} onDelete={onDelete} />
    )

    // Each press gets its own awaited act() - two bare fireEvent.press
    // calls back to back leave overlapping act scopes that corrupt
    // React's act-nesting for whatever test renders next in this file.
    await act(async () => {
      fireEvent.press(screen.getByLabelText('Edit Star Platinum'))
    })
    expect(onEdit).toHaveBeenCalledTimes(1)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Delete Star Platinum'))
    })
    expect(onDelete).toHaveBeenCalledTimes(1)
  })

  it('fires onOpenDetail when the card body (not the thumbnail) is pressed', async () => {
    const onOpenDetail = jest.fn()
    await renderWithProviders(
      <StandCard stand={baseStand()} onOpenDetail={onOpenDetail} onEdit={jest.fn()} onDelete={jest.fn()} />
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('View Star Platinum details'))
    })
    expect(onOpenDetail).toHaveBeenCalledTimes(1)
  })

  it('hides the edit/delete buttons when readOnly', async () => {
    await renderWithProviders(<StandCard stand={baseStand()} onOpenDetail={jest.fn()} readOnly />)

    expect(screen.queryByLabelText('Edit Star Platinum')).toBeNull()
    expect(screen.queryByLabelText('Delete Star Platinum')).toBeNull()
  })
})
