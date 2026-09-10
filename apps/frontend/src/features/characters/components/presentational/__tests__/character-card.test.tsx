import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import enGB from '@/shared/i18n/locales/en-GB.json'
import { JOJO_STAT_ROWS } from '@/features/characters/lib/character-stats'
import type { JojoCharacterResponse } from '@/features/characters/types/characters.types'

import { CharacterCard } from '../character-card'

function baseCharacter(overrides: Partial<JojoCharacterResponse> = {}): JojoCharacterResponse {
  return {
    id: 'char-1',
    name: 'Jotaro Kujo',
    description: 'A delinquent with a powerful Stand.',
    rarity: 'LEGENDARY',
    picture: '',
    pictureThumb: '',
    pictureCard: '',
    pictureStatus: 'NONE',
    pictureLqip: '',
    hamon: 'NONE',
    spin: 'NONE',
    battleIq: 130,
    ...overrides,
  }
}

describe('CharacterCard', () => {
  it('renders the name, rarity and stat block', async () => {
    await renderWithProviders(
      <CharacterCard
        character={baseCharacter()}
        rows={JOJO_STAT_ROWS}
        onOpenDetail={jest.fn()}
        onEdit={jest.fn()}
        onDelete={jest.fn()}
      />
    )

    expect(screen.getByText('Jotaro Kujo')).toBeTruthy()
    expect(screen.getByText(enGB.enums.rarity.LEGENDARY)).toBeTruthy()
    expect(screen.getByText(enGB.characters.stats.hamon)).toBeTruthy()
    expect(screen.getByText(enGB.characters.stats.battleIq)).toBeTruthy()
  })

  it('fires onEdit and onDelete from their own buttons', async () => {
    const onEdit = jest.fn()
    const onDelete = jest.fn()
    await renderWithProviders(
      <CharacterCard
        character={baseCharacter()}
        rows={JOJO_STAT_ROWS}
        onOpenDetail={jest.fn()}
        onEdit={onEdit}
        onDelete={onDelete}
      />
    )

    // Each press gets its own awaited act() - see devil-fruit-card.test.tsx's
    // equivalent test for why.
    await act(async () => {
      fireEvent.press(screen.getByLabelText('Edit Jotaro Kujo'))
    })
    expect(onEdit).toHaveBeenCalledTimes(1)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Delete Jotaro Kujo'))
    })
    expect(onDelete).toHaveBeenCalledTimes(1)
  })

  it('fires onOpenDetail when the card body is pressed', async () => {
    const onOpenDetail = jest.fn()
    await renderWithProviders(
      <CharacterCard
        character={baseCharacter()}
        rows={JOJO_STAT_ROWS}
        onOpenDetail={onOpenDetail}
        onEdit={jest.fn()}
        onDelete={jest.fn()}
      />
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('View Jotaro Kujo details'))
    })
    expect(onOpenDetail).toHaveBeenCalledTimes(1)
  })

  it('hides the edit/delete buttons when readOnly', async () => {
    await renderWithProviders(
      <CharacterCard
        character={baseCharacter()}
        rows={JOJO_STAT_ROWS}
        onOpenDetail={jest.fn()}
        readOnly
      />
    )

    expect(screen.queryByLabelText('Edit Jotaro Kujo')).toBeNull()
    expect(screen.queryByLabelText('Delete Jotaro Kujo')).toBeNull()
  })
})
