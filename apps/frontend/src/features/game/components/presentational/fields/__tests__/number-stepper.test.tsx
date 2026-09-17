import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { NumberStepper } from '../number-stepper'

async function renderStepper(overrides: Partial<React.ComponentProps<typeof NumberStepper>> = {}) {
  const onChange = jest.fn()
  const props = {
    label: 'Voting window (seconds)',
    value: 30,
    min: 5,
    max: 180,
    onChange,
    ...overrides,
  }
  await renderWithProviders(<NumberStepper {...props} />)
  return { onChange }
}

describe('NumberStepper', () => {
  it('increments and decrements via the -/+ buttons', async () => {
    const { onChange } = await renderStepper({ value: 30 })

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Voting window (seconds) +'))
    })
    expect(onChange).toHaveBeenCalledWith(31)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Voting window (seconds) -'))
    })
    expect(onChange).toHaveBeenCalledWith(29)
  })

  // Disabled Tamagui Buttons render `aria-disabled` (not RN's
  // accessibilityState) and `fireEvent.press` bypasses real touch gating in
  // RNTL - assert the disabled flag directly instead of firing press and
  // expecting no call (same pattern as lobby-config-panel.test.tsx's Save
  // button guards).
  it('marks the - button disabled at min', async () => {
    await renderStepper({ value: 5, min: 5, max: 180 })
    expect(screen.getByLabelText('Voting window (seconds) -').props['aria-disabled']).toBe(true)
    expect(screen.getByLabelText('Voting window (seconds) +').props['aria-disabled']).not.toBe(true)
  })

  it('marks the + button disabled at max', async () => {
    await renderStepper({ value: 180, min: 5, max: 180 })
    expect(screen.getByLabelText('Voting window (seconds) +').props['aria-disabled']).toBe(true)
    expect(screen.getByLabelText('Voting window (seconds) -').props['aria-disabled']).not.toBe(true)
  })

  it('opens a digits-only input when the value is pressed', async () => {
    await renderStepper({ value: 30 })

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Voting window (seconds): 30'))
    })

    const input = screen.getByDisplayValue('30')
    expect(input).toBeTruthy()

    await act(async () => {
      fireEvent.changeText(input, 'ab12cd')
    })
    expect(screen.getByDisplayValue('12')).toBeTruthy()
  })

  it('clamps a typed value above max on commit', async () => {
    const { onChange } = await renderStepper({ value: 30, min: 5, max: 180 })

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Voting window (seconds): 30'))
    })
    const input = screen.getByDisplayValue('30')
    await act(async () => {
      fireEvent.changeText(input, '500')
    })
    await act(async () => {
      fireEvent(input, 'submitEditing')
    })

    expect(onChange).toHaveBeenCalledWith(180)
  })

  it('clamps a typed value below min on commit', async () => {
    const { onChange } = await renderStepper({ value: 30, min: 5, max: 180 })

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Voting window (seconds): 30'))
    })
    const input = screen.getByDisplayValue('30')
    await act(async () => {
      fireEvent.changeText(input, '0')
    })
    await act(async () => {
      fireEvent(input, 'submitEditing')
    })

    expect(onChange).toHaveBeenCalledWith(5)
  })

  it('reverts to the previous value when the draft is emptied', async () => {
    const { onChange } = await renderStepper({ value: 30 })

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Voting window (seconds): 30'))
    })
    const input = screen.getByDisplayValue('30')
    await act(async () => {
      fireEvent.changeText(input, '')
    })
    await act(async () => {
      fireEvent(input, 'submitEditing')
    })

    expect(onChange).not.toHaveBeenCalled()
    expect(screen.getByLabelText('Voting window (seconds): 30')).toBeTruthy()
  })

  it('closes without committing on blur after Escape', async () => {
    const { onChange } = await renderStepper({ value: 30 })

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Voting window (seconds): 30'))
    })
    const input = screen.getByDisplayValue('30')
    await act(async () => {
      fireEvent.changeText(input, '99')
    })
    await act(async () => {
      fireEvent(input, 'keyPress', { nativeEvent: { key: 'Escape' } })
    })

    expect(onChange).not.toHaveBeenCalled()
    expect(screen.getByLabelText('Voting window (seconds): 30')).toBeTruthy()
  })
})
