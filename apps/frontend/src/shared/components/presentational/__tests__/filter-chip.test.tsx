import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { FilterChip } from '../filter-chip'

describe('FilterChip', () => {
  it('renders its label and reports the checked state', async () => {
    await renderWithProviders(
      <FilterChip label="No picture" tooltip="Show only entries without a picture" active onToggle={jest.fn()} />
    )

    const chip = screen.getByLabelText('No picture')
    expect(chip.props.accessibilityState?.checked ?? chip.props['aria-checked']).toBe(true)
  })

  it('calls onToggle when pressed', async () => {
    const onToggle = jest.fn()
    await renderWithProviders(
      <FilterChip
        label="No picture"
        tooltip="Show only entries without a picture"
        active={false}
        onToggle={onToggle}
      />
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('No picture'))
    })

    expect(onToggle).toHaveBeenCalledTimes(1)
  })
})
