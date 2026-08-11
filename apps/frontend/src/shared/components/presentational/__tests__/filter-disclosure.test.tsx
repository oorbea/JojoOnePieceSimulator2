import { Text } from 'react-native'
import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { FilterDisclosure } from '../filter-disclosure'

describe('FilterDisclosure', () => {
  it('does not render children while collapsed', async () => {
    await renderWithProviders(
      <FilterDisclosure
        label="More filters"
        activeCount={0}
        expanded={false}
        onToggle={jest.fn()}
        clearLabel="Clear"
      >
        <Text>Attack Power</Text>
      </FilterDisclosure>
    )

    expect(screen.queryByText('Attack Power')).toBeNull()
  })

  it('renders children once expanded', async () => {
    await renderWithProviders(
      <FilterDisclosure
        label="More filters"
        activeCount={0}
        expanded
        onToggle={jest.fn()}
        clearLabel="Clear"
      >
        <Text>Attack Power</Text>
      </FilterDisclosure>
    )

    expect(screen.getByText('Attack Power')).toBeTruthy()
  })

  it('calls onToggle when the header is pressed', async () => {
    const onToggle = jest.fn()
    await renderWithProviders(
      <FilterDisclosure
        label="More filters"
        activeCount={0}
        expanded={false}
        onToggle={onToggle}
        clearLabel="Clear"
      >
        <Text>Attack Power</Text>
      </FilterDisclosure>
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('More filters'))
    })

    expect(onToggle).toHaveBeenCalledTimes(1)
  })

  it('shows the active-filter count as a badge and hides it at zero', async () => {
    await renderWithProviders(
      <FilterDisclosure
        label="More filters"
        activeCount={2}
        expanded={false}
        onToggle={jest.fn()}
        clearLabel="Clear"
      >
        <Text>Attack Power</Text>
      </FilterDisclosure>
    )

    expect(screen.getByText('2')).toBeTruthy()
  })

  it('only shows the clear-all action once a filter is active', async () => {
    const { rerender } = await renderWithProviders(
      <FilterDisclosure
        label="More filters"
        activeCount={0}
        expanded={false}
        onToggle={jest.fn()}
        onClearAll={jest.fn()}
        clearLabel="Clear all"
      >
        <Text>Attack Power</Text>
      </FilterDisclosure>
    )

    expect(screen.queryByLabelText('Clear all')).toBeNull()

    const onClearAll = jest.fn()
    await rerender(
      <FilterDisclosure
        label="More filters"
        activeCount={1}
        expanded={false}
        onToggle={jest.fn()}
        onClearAll={onClearAll}
        clearLabel="Clear all"
      >
        <Text>Attack Power</Text>
      </FilterDisclosure>
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Clear all'))
    })
    expect(onClearAll).toHaveBeenCalledTimes(1)
  })
})
