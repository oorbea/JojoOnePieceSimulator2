import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { ManualOverlayContext } from '@/shared/lib/manual-overlay'

import { ManualHelpButton } from '../manual-help-button'

describe('ManualHelpButton', () => {
  it('opens the manual at its section, and names that section for assistive tech', async () => {
    const open = jest.fn()
    await renderWithProviders(
      <ManualOverlayContext.Provider value={open}>
        <ManualHelpButton section="effects" />
      </ManualOverlayContext.Provider>
    )

    fireEvent.press(screen.getByLabelText('Open the manual: Power effects'))

    expect(open).toHaveBeenCalledWith('effects')
  })

  it('does nothing outside the overlay provider (login, isolated screens)', async () => {
    await renderWithProviders(<ManualHelpButton section="how" />)

    expect(() =>
      fireEvent.press(screen.getByLabelText('Open the manual: How to play'))
    ).not.toThrow()
  })
})
