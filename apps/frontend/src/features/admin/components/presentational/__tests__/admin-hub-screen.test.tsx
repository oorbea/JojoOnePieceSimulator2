import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { AdminHubScreen } from '../admin-hub-screen'

describe('AdminHubScreen', () => {
  it('renders both channel tiles', async () => {
    await renderWithProviders(<AdminHubScreen onOpenStands={jest.fn()} onOpenDevilFruits={jest.fn()} />)

    expect(screen.getByText('Stands')).toBeTruthy()
    expect(screen.getByText('Devil Fruits')).toBeTruthy()
  })

  it('fires onOpenStands when the Stands tile is pressed', async () => {
    const onOpenStands = jest.fn()
    await renderWithProviders(<AdminHubScreen onOpenStands={onOpenStands} onOpenDevilFruits={jest.fn()} />)

    fireEvent.press(screen.getByLabelText('Stands'))

    expect(onOpenStands).toHaveBeenCalledTimes(1)
  })

  it('fires onOpenDevilFruits when the Devil Fruits tile is pressed', async () => {
    const onOpenDevilFruits = jest.fn()
    await renderWithProviders(<AdminHubScreen onOpenStands={jest.fn()} onOpenDevilFruits={onOpenDevilFruits} />)

    fireEvent.press(screen.getByLabelText('Devil Fruits'))

    expect(onOpenDevilFruits).toHaveBeenCalledTimes(1)
  })
})
