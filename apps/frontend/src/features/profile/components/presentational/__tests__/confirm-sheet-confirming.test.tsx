import { renderWithProviders, screen } from '@/test/render'

import { ConfirmSheet } from '../confirm-sheet'

// Kept in its own file: RN's real Modal drives its native fade with the
// real `Animated` timing loop, and several ConfirmSheet renders in a single
// file (confirm-sheet.test.tsx covers those) leave earlier tests'
// still-settling animations on real timers when this one starts, corrupting
// its render. Separate test files get separate module/timer state, so this
// one test is isolated from that entirely.
describe('ConfirmSheet while confirming', () => {
  it('disables both actions and shows the working label', async () => {
    await renderWithProviders(
      <ConfirmSheet
        visible
        title="Remove avatar?"
        message="Your picture will revert to the one from your Google account."
        confirmLabel="Remove avatar"
        isConfirming
        onConfirm={jest.fn()}
        onCancel={jest.fn()}
      />
    )

    // Disabled Tamagui Buttons render `aria-disabled` (not RN's
    // accessibilityState), and RNTL's label queries exclude elements it
    // considers inert — so query by the still-visible text instead and
    // check its disabled ancestor directly.
    const confirmButton = screen.getByText('Working…').parent
    const cancelButton = screen.getByText('Cancel').parent
    expect(confirmButton?.props['aria-disabled']).toBe(true)
    expect(cancelButton?.props['aria-disabled']).toBe(true)
  })
})
