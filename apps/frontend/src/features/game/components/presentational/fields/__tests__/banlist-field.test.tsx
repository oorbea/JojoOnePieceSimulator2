import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { BanlistField, type BannableItem } from '../banlist-field'

function item(overrides: Partial<BannableItem> = {}): BannableItem {
  return {
    id: 'stand-1',
    name: 'Star Platinum',
    kind: 'STAND',
    rarity: 'LEGENDARY',
    ...overrides,
  } as BannableItem
}

const ITEMS: BannableItem[] = [
  item({ id: 'stand-1', name: 'Star Platinum', kind: 'STAND' }),
  item({ id: 'fruit-1', name: 'Gomu Gomu no Mi', kind: 'DEVIL_FRUIT', rarity: 'RARE' }),
]

describe('BanlistField', () => {
  it('matches results by name case/diacritic-insensitively', async () => {
    const items: BannableItem[] = [item({ id: 'x', name: 'Méra Méra no Mi', kind: 'DEVIL_FRUIT' })]
    await renderWithProviders(
      <BanlistField editable banned={[]} items={items} onAddBan={jest.fn()} onRemoveBan={jest.fn()} />
    )

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search Stands and Devil Fruits'), 'mera mera')
    })

    expect(screen.getByText('Méra Méra no Mi')).toBeTruthy()
  })

  it('tapping a result calls onAddBan and clears the search', async () => {
    const onAddBan = jest.fn()
    const { rerender } = await renderWithProviders(
      <BanlistField editable banned={[]} items={ITEMS} onAddBan={onAddBan} onRemoveBan={jest.fn()} />
    )

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search Stands and Devil Fruits'), 'star')
    })
    await act(async () => {
      fireEvent.press(screen.getByLabelText('Star Platinum'))
    })

    expect(onAddBan).toHaveBeenCalledWith('stand-1')

    // Search text was cleared as a side effect - simulate the container
    // committing the new ban and confirm the result row is gone (the item
    // is now excluded as already-banned, not just because the search text
    // reset - re-typing the same query proves which).
    await act(async () => {
      rerender(
        <BanlistField editable banned={['stand-1']} items={ITEMS} onAddBan={onAddBan} onRemoveBan={jest.fn()} />
      )
    })
    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search Stands and Devil Fruits'), 'star')
    })
    expect(screen.queryByLabelText('Star Platinum')).toBeNull()
  })

  it('the clear-search button clears results', async () => {
    await renderWithProviders(
      <BanlistField editable banned={[]} items={ITEMS} onAddBan={jest.fn()} onRemoveBan={jest.fn()} />
    )

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search Stands and Devil Fruits'), 'star')
    })
    expect(screen.getByLabelText('Star Platinum')).toBeTruthy()

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Clear search'))
    })

    expect(screen.queryByLabelText('Star Platinum')).toBeNull()
  })

  it('shows "no matches" copy for a nonsense query', async () => {
    await renderWithProviders(
      <BanlistField editable banned={[]} items={ITEMS} onAddBan={jest.fn()} onRemoveBan={jest.fn()} />
    )

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search Stands and Devil Fruits'), 'zzzzzz')
    })

    expect(screen.getByText('No powers match your search')).toBeTruthy()
  })

  it('renders the clear-banlist button only when banned is non-empty and calls the callback', async () => {
    const onClearBanlist = jest.fn()
    const { rerender } = await renderWithProviders(
      <BanlistField
        editable
        banned={[]}
        items={ITEMS}
        onAddBan={jest.fn()}
        onRemoveBan={jest.fn()}
        onClearBanlist={onClearBanlist}
      />
    )
    expect(screen.queryByLabelText('Clear banlist')).toBeNull()

    await act(async () => {
      rerender(
        <BanlistField
          editable
          banned={['stand-1']}
          items={ITEMS}
          onAddBan={jest.fn()}
          onRemoveBan={jest.fn()}
          onClearBanlist={onClearBanlist}
        />
      )
    })
    await act(async () => {
      fireEvent.press(screen.getByLabelText('Clear banlist'))
    })
    expect(onClearBanlist).toHaveBeenCalledTimes(1)
  })

  it('hides search and clear buttons when editable is false', async () => {
    await renderWithProviders(
      <BanlistField
        editable={false}
        banned={['stand-1']}
        items={ITEMS}
        onAddBan={jest.fn()}
        onRemoveBan={jest.fn()}
        onClearBanlist={jest.fn()}
      />
    )

    expect(screen.queryByLabelText('Search Stands and Devil Fruits')).toBeNull()
    expect(screen.queryByLabelText('Clear banlist')).toBeNull()
    expect(screen.queryByLabelText('Clear search')).toBeNull()
  })
})
