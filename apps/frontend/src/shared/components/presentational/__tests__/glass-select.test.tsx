import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { GlassSelect } from '../glass-select'

const OPTIONS = [
  { value: 'COMMON', label: 'COMMON' },
  { value: 'RARE', label: 'RARE' },
  { value: 'EPIC', label: 'EPIC' },
]

// Opening the picker flips the Modal's `visible` prop via internal state —
// like GlassField's onChangeText (see skills-field.test.tsx), that update
// needs an explicit act() flush before the next query sees the modal's
// content, or the query runs against the still-closed tree.
async function openPicker(label: string) {
  await act(async () => {
    fireEvent.press(screen.getByLabelText(label))
  })
}

describe('GlassSelect', () => {
  it('shows the placeholder when no value is selected', async () => {
    await renderWithProviders(
      <GlassSelect label="Rarity" options={OPTIONS} value={null} onChange={jest.fn()} placeholder="Select…" />
    )

    expect(screen.getByText('Select…')).toBeTruthy()
  })

  it('shows the selected option label', async () => {
    await renderWithProviders(
      <GlassSelect label="Rarity" options={OPTIONS} value="RARE" onChange={jest.fn()} />
    )

    expect(screen.getByText('RARE')).toBeTruthy()
  })

  it('opens the picker and calls onChange with the picked value', async () => {
    const onChange = jest.fn()
    await renderWithProviders(<GlassSelect label="Rarity" options={OPTIONS} value={null} onChange={onChange} />)

    await openPicker('Rarity: Select…')
    fireEvent.press(screen.getByLabelText('EPIC'))

    expect(onChange).toHaveBeenCalledWith('EPIC')
  })

  it('clears the value when the clear button is pressed', async () => {
    const onChange = jest.fn()
    await renderWithProviders(
      <GlassSelect label="Rarity" options={OPTIONS} value="RARE" onChange={onChange} clearable />
    )

    fireEvent.press(screen.getByLabelText('Clear Rarity'))

    expect(onChange).toHaveBeenCalledWith(null)
  })

  it('filters options by the search field when searchable', async () => {
    const onChange = jest.fn()
    await renderWithProviders(
      <GlassSelect label="Rarity" options={OPTIONS} value={null} onChange={onChange} searchable />
    )

    await openPicker('Rarity: Select…')
    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search'), 'epi')
    })

    expect(screen.getByText('EPIC')).toBeTruthy()
    expect(screen.queryByText('COMMON')).toBeNull()
  })
})
