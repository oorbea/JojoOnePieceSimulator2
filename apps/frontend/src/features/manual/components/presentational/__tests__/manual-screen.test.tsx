import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { MANUAL_RULES } from '@/shared/contracts/rules'

import { MANUAL_SECTIONS } from '../../../lib/sections'
import { ManualScreen } from '../manual-screen'

async function renderScreen(overrides: Partial<React.ComponentProps<typeof ManualScreen>> = {}) {
  const props: React.ComponentProps<typeof ManualScreen> = {
    rules: MANUAL_RULES,
    active: null,
    onSelectSection: jest.fn(),
    onMeasureRow: jest.fn(),
    onMeasureColumn: jest.fn(),
    onMeasureSection: jest.fn(),
    ...overrides,
  }
  await renderWithProviders(<ManualScreen {...props} />)
  return { props }
}

describe('ManualScreen', () => {
  it('renders every section from the generated rules', async () => {
    await renderScreen()

    for (const id of MANUAL_SECTIONS) {
      expect(screen.getByTestId(`manual-section-${id}`)).toBeTruthy()
    }
  })

  it('quotes the numbers straight from the game limits', async () => {
    await renderScreen()

    const { votingDefaultSeconds, votingMinSeconds, votingMaxSeconds, votingExtensionSeconds } =
      MANUAL_RULES.limits
    expect(
      screen.getByText(
        new RegExp(
          `within ${votingDefaultSeconds} seconds by default \\(the host can set anything from ${votingMinSeconds} to ${votingMaxSeconds}\\)`
        )
      )
    ).toBeTruthy()
    expect(screen.getByText(new RegExp(`add ${votingExtensionSeconds} seconds`))).toBeTruthy()
  })

  it('shows the evolution outcomes the real resolver computed', async () => {
    await renderScreen()

    // Tusk: Act 1 + Golden Spin always lands on Act 3 (the tie goes to the more evolved stage).
    expect(screen.getAllByText('Spin Golden: evolves into Tusk: Act 3').length).toBeGreaterThan(0)
    expect(
      screen.getAllByText('Fruit Mastery Awakened: evolves into Hito Hito no mi: Model Nika').length
    ).toBeGreaterThan(0)
  })

  it('renders each combat convention, flagging the ones for the voters to judge', async () => {
    await renderScreen()

    for (const c of MANUAL_RULES.conventions) {
      expect(screen.getByTestId(`manual-convention-${c.id}`)).toBeTruthy()
    }
    expect(screen.getAllByText('For the voters to judge').length).toBe(
      MANUAL_RULES.conventions.filter((c) => c.judgement).length
    )
  })

  it('tells the table of contents which section was picked', async () => {
    const { props } = await renderScreen()

    fireEvent.press(screen.getByLabelText('Draw odds'))

    expect(props.onSelectSection).toHaveBeenCalledWith('odds')
  })
})
