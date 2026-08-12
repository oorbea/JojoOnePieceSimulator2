import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { AdminHubScreen } from '../admin-hub-screen'

function renderScreen(overrides: Partial<Parameters<typeof AdminHubScreen>[0]> = {}) {
  return renderWithProviders(
    <AdminHubScreen
      onOpenStands={jest.fn()}
      onOpenDevilFruits={jest.fn()}
      onOpenStages={jest.fn()}
      {...overrides}
    />
  )
}

describe('AdminHubScreen', () => {
  it('renders every channel tile', async () => {
    await renderScreen()

    expect(screen.getByText('Stands')).toBeTruthy()
    expect(screen.getByText('Devil Fruits')).toBeTruthy()
    expect(screen.getByText('Stages')).toBeTruthy()
  })

  it('fires onOpenStands when the Stands tile is pressed', async () => {
    const onOpenStands = jest.fn()
    await renderScreen({ onOpenStands })

    fireEvent.press(screen.getByLabelText('Stands'))

    expect(onOpenStands).toHaveBeenCalledTimes(1)
  })

  it('fires onOpenDevilFruits when the Devil Fruits tile is pressed', async () => {
    const onOpenDevilFruits = jest.fn()
    await renderScreen({ onOpenDevilFruits })

    fireEvent.press(screen.getByLabelText('Devil Fruits'))

    expect(onOpenDevilFruits).toHaveBeenCalledTimes(1)
  })

  it('fires onOpenStages when the Stages tile is pressed', async () => {
    const onOpenStages = jest.fn()
    await renderScreen({ onOpenStages })

    fireEvent.press(screen.getByLabelText('Stages'))

    expect(onOpenStages).toHaveBeenCalledTimes(1)
  })
})
