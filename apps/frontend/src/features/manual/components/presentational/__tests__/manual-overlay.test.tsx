import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { MANUAL_RULES } from '@/shared/contracts/rules'
import { MANUAL_SECTIONS } from '@/shared/lib/manual-sections'

import { ManualOverlay } from '../manual-overlay'

async function renderOverlay(
  section: Parameters<typeof ManualOverlay>[0]['section'],
  overrides: Partial<Parameters<typeof ManualOverlay>[0]> = {}
) {
  const props = {
    section,
    rules: MANUAL_RULES,
    onClose: jest.fn(),
    onOpenFull: jest.fn(),
    ...overrides,
  }
  await renderWithProviders(<ManualOverlay {...props} />)
  return props
}

describe('ManualOverlay', () => {
  it('shows nothing while no section is open', async () => {
    await renderOverlay(null)

    expect(screen.queryByText('Combat conventions')).toBeNull()
    expect(screen.queryByText('Open the full manual')).toBeNull()
  })

  it('shows one section under its title, from the same generated rules as the page', async () => {
    await renderOverlay('conventions')

    expect(screen.getByText('Combat conventions')).toBeTruthy()
    for (const c of MANUAL_RULES.conventions) {
      expect(screen.getByTestId(`manual-convention-${c.id}`)).toBeTruthy()
    }
    // The modal provides the title: the section must not draw its own card heading again.
    expect(screen.getAllByText('Combat conventions')).toHaveLength(1)
  })

  it('can open every section of the manual', async () => {
    for (const id of MANUAL_SECTIONS) {
      const { unmount } = await renderWithProviders(
        <ManualOverlay
          section={id}
          rules={MANUAL_RULES}
          onClose={jest.fn()}
          onOpenFull={jest.fn()}
        />
      )
      expect(screen.getByTestId(`manual-section-${id}`)).toBeTruthy()
      await unmount()
    }
  })

  it('offers the full manual at the open section', async () => {
    const props = await renderOverlay('odds')

    fireEvent.press(screen.getByLabelText('Open the full manual'))

    expect(props.onOpenFull).toHaveBeenCalledWith('odds')
  })
})
