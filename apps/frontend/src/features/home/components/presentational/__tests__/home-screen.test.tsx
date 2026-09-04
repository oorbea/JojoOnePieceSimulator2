import { fireEvent, renderWithProviders, screen } from '@/test/render'
import enGB from '@/shared/i18n/locales/en-GB.json'
import type { SessionUser } from '@/shared/stores/session.store'

import { HomeScreen } from '../home-screen'

const USER: SessionUser = {
  id: 'user-1',
  email: 'jotaro@example.com',
  username: 'OriolO',
  completeName: 'Jotaro Kujo',
  picture: null,
  role: 'REGULAR',
  language: 'en-GB',
}

describe('HomeScreen', () => {
  it("shows the user's username in place of the old greeting sentence", async () => {
    await renderWithProviders(<HomeScreen user={USER} onNavigate={jest.fn()} />)

    expect(screen.getByText('OriolO')).toBeTruthy()
    expect(screen.queryByText(/ready when you are/i)).toBeNull()
  })

  it('shows the email and role pills', async () => {
    await renderWithProviders(<HomeScreen user={USER} onNavigate={jest.fn()} />)

    expect(screen.getByText(USER.email)).toBeTruthy()
    expect(screen.getByText(enGB.enums.role[USER.role])).toBeTruthy()
  })

  // AppShell's top bar already has a logout button on every authenticated
  // route — a second one here would just duplicate it.
  it('has no logout control of its own', async () => {
    await renderWithProviders(<HomeScreen user={USER} onNavigate={jest.fn()} />)

    expect(screen.queryByLabelText(/log out/i)).toBeNull()
  })

  it('navigates to the right route for every channel - none are locked anymore', async () => {
    const onNavigate = jest.fn()
    await renderWithProviders(<HomeScreen user={USER} onNavigate={onNavigate} />)

    fireEvent.press(screen.getByLabelText('Play'))
    fireEvent.press(screen.getByLabelText('Profile'))
    fireEvent.press(screen.getByLabelText('Stands'))
    fireEvent.press(screen.getByLabelText('Devil Fruits'))
    fireEvent.press(screen.getByLabelText('Stages'))

    expect(onNavigate).toHaveBeenNthCalledWith(1, '/play')
    expect(onNavigate).toHaveBeenNthCalledWith(2, '/profile')
    expect(onNavigate).toHaveBeenNthCalledWith(3, '/catalog/stands')
    expect(onNavigate).toHaveBeenNthCalledWith(4, '/catalog/devil-fruits')
    expect(onNavigate).toHaveBeenNthCalledWith(5, '/catalog/stages')
    expect(screen.queryByText(/coming soon/i)).toBeNull()
  })
})
