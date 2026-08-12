import { act, render, screen } from '@testing-library/react-native'
import { Text } from 'react-native'

import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'

// Renders through a tiny probe component rather than a renderHook helper -
// this codebase's test harness (src/test/render.tsx) has no renderHook
// precedent, and this keeps the test on the same render()/act() path every
// other component test already uses.
function Probe({ value }: { value: string }) {
  const debounced = useDebouncedValue(value)
  return <Text testID="debounced">{debounced}</Text>
}

async function advance(ms: number) {
  await act(async () => {
    jest.advanceTimersByTime(ms)
  })
}

describe('useDebouncedValue', () => {
  beforeEach(() => {
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('returns the initial value immediately', async () => {
    await render(<Probe value="a" />)
    expect(screen.getByTestId('debounced').props.children).toBe('a')
  })

  it('does not update before the delay elapses', async () => {
    const { rerender } = await render(<Probe value="a" />)
    await rerender(<Probe value="ab" />)
    await advance(299)
    expect(screen.getByTestId('debounced').props.children).toBe('a')
  })

  it('updates once the delay elapses', async () => {
    const { rerender } = await render(<Probe value="a" />)
    await rerender(<Probe value="ab" />)
    await advance(300)
    expect(screen.getByTestId('debounced').props.children).toBe('ab')
  })

  it('resets the timer on every intermediate change, only settling on the last value', async () => {
    const { rerender } = await render(<Probe value="a" />)
    await rerender(<Probe value="ab" />)
    await advance(150)
    await rerender(<Probe value="abc" />)
    await advance(150)
    // Only 150ms have elapsed since the last change - still debouncing.
    expect(screen.getByTestId('debounced').props.children).toBe('a')

    await advance(150)
    expect(screen.getByTestId('debounced').props.children).toBe('abc')
  })
})
