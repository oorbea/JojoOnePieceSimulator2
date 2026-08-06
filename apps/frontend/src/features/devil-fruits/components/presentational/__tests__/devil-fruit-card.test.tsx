import { fireEvent, renderWithProviders, screen } from '@/test/render'
import type { DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

import { DevilFruitCard } from '../devil-fruit-card'

function baseFruit(overrides: Partial<DevilFruitResponse> = {}): DevilFruitResponse {
  return {
    id: 'fruit-1',
    name: 'Gomu Gomu no Mi',
    description: 'Grants rubber properties.',
    rarity: 'EPIC',
    skills: ['Gomu Gomu no Pistol'],
    picture: '',
    pictureThumb: '',
    pictureStatus: 'NONE',
    fruitType: 'PARAMECIA',
    ...overrides,
  }
}

describe('DevilFruitCard', () => {
  it('renders the name, rarity and fruit type', async () => {
    await renderWithProviders(<DevilFruitCard devilFruit={baseFruit()} onEdit={jest.fn()} onDelete={jest.fn()} />)

    expect(screen.getByText('Gomu Gomu no Mi')).toBeTruthy()
    expect(screen.getByText('EPIC')).toBeTruthy()
    expect(screen.getByText('PARAMECIA')).toBeTruthy()
  })

  it('fires onEdit and onDelete from their own buttons', async () => {
    const onEdit = jest.fn()
    const onDelete = jest.fn()
    await renderWithProviders(<DevilFruitCard devilFruit={baseFruit()} onEdit={onEdit} onDelete={onDelete} />)

    fireEvent.press(screen.getByLabelText('Edit Gomu Gomu no Mi'))
    expect(onEdit).toHaveBeenCalledTimes(1)

    fireEvent.press(screen.getByLabelText('Delete Gomu Gomu no Mi'))
    expect(onDelete).toHaveBeenCalledTimes(1)
  })
})
