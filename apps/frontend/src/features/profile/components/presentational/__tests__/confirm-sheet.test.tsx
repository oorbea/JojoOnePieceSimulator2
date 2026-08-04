import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { ConfirmSheet } from '../confirm-sheet'

function baseProps(overrides: Record<string, unknown> = {}) {
  return {
    visible: true,
    title: 'Remove avatar?',
    message: 'Your picture will revert to the one from your Google account.',
    confirmLabel: 'Remove avatar',
    onConfirm: jest.fn(),
    onCancel: jest.fn(),
    ...overrides,
  }
}

describe('ConfirmSheet', () => {
  it('renders nothing when not visible', async () => {
    await renderWithProviders(<ConfirmSheet {...baseProps({ visible: false })} />)

    expect(screen.queryByText('Remove avatar?')).toBeNull()
  })

  it('shows the title, message and both actions when visible', async () => {
    await renderWithProviders(<ConfirmSheet {...baseProps()} />)

    expect(screen.getByText('Remove avatar?')).toBeTruthy()
    expect(screen.getByText(baseProps().message as string)).toBeTruthy()
    expect(screen.getByLabelText('Remove avatar')).toBeTruthy()
    expect(screen.getByLabelText('Cancel')).toBeTruthy()
  })

  it('cancels when the dimmed backdrop is pressed', async () => {
    const onCancel = jest.fn()
    await renderWithProviders(<ConfirmSheet {...baseProps({ onCancel })} />)

    // The backdrop is the alert-role container wrapping the card — pressing
    // it (not the card, and not either button) must still cancel.
    fireEvent.press(screen.getByLabelText('Remove avatar?'))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })

  it('does not cancel when the card itself is pressed', async () => {
    const onCancel = jest.fn()
    await renderWithProviders(<ConfirmSheet {...baseProps({ onCancel })} />)

    fireEvent.press(screen.getByText('Remove avatar?'))
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('calls onConfirm and onCancel from their own buttons', async () => {
    const onConfirm = jest.fn()
    const onCancel = jest.fn()
    await renderWithProviders(<ConfirmSheet {...baseProps({ onConfirm, onCancel })} />)

    fireEvent.press(screen.getByLabelText('Remove avatar'))
    expect(onConfirm).toHaveBeenCalledTimes(1)
    expect(onCancel).not.toHaveBeenCalled()

    fireEvent.press(screen.getByLabelText('Cancel'))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })
})
