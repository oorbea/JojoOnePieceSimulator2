import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import enGB from '@/shared/i18n/locales/en-GB.json'
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
    await renderWithProviders(
      <DevilFruitCard devilFruit={baseFruit()} onOpenDetail={jest.fn()} onEdit={jest.fn()} onDelete={jest.fn()} />
    )

    expect(screen.getByText('Gomu Gomu no Mi')).toBeTruthy()
    expect(screen.getByText(enGB.enums.rarity.EPIC)).toBeTruthy()
    expect(screen.getByText(enGB.enums.fruitType.PARAMECIA)).toBeTruthy()
  })

  it('fires onEdit and onDelete from their own buttons', async () => {
    const onEdit = jest.fn()
    const onDelete = jest.fn()
    await renderWithProviders(
      <DevilFruitCard devilFruit={baseFruit()} onOpenDetail={jest.fn()} onEdit={onEdit} onDelete={onDelete} />
    )

    // Each press gets its own awaited act() - two bare fireEvent.press
    // calls back to back leave overlapping act scopes that corrupt
    // React's act-nesting for whatever test renders next in this file.
    await act(async () => {
      fireEvent.press(screen.getByLabelText('Edit Gomu Gomu no Mi'))
    })
    expect(onEdit).toHaveBeenCalledTimes(1)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Delete Gomu Gomu no Mi'))
    })
    expect(onDelete).toHaveBeenCalledTimes(1)
  })

  it('fires onOpenDetail when the card body is pressed', async () => {
    const onOpenDetail = jest.fn()
    await renderWithProviders(
      <DevilFruitCard devilFruit={baseFruit()} onOpenDetail={onOpenDetail} onEdit={jest.fn()} onDelete={jest.fn()} />
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('View Gomu Gomu no Mi details'))
    })
    expect(onOpenDetail).toHaveBeenCalledTimes(1)
  })

  it('hides the edit/delete buttons when readOnly', async () => {
    await renderWithProviders(<DevilFruitCard devilFruit={baseFruit()} onOpenDetail={jest.fn()} readOnly />)

    expect(screen.queryByLabelText('Edit Gomu Gomu no Mi')).toBeNull()
    expect(screen.queryByLabelText('Delete Gomu Gomu no Mi')).toBeNull()
  })
})
