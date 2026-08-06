import { fireEvent, renderWithProviders, screen } from '@/test/render'
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
    await renderWithProviders(<HomeScreen user={USER} onOpenProfile={jest.fn()} />)

    expect(screen.getByText('OriolO')).toBeTruthy()
    expect(screen.queryByText(/ready when you are/i)).toBeNull()
  })

  it('shows the email and role pills', async () => {
    await renderWithProviders(<HomeScreen user={USER} onOpenProfile={jest.fn()} />)

    expect(screen.getByText(USER.email)).toBeTruthy()
    expect(screen.getByText(USER.role)).toBeTruthy()
  })

  // AppShell's top bar already has a logout button on every authenticated
  // route — a second one here would just duplicate it.
  it('has no logout control of its own', async () => {
    await renderWithProviders(<HomeScreen user={USER} onOpenProfile={jest.fn()} />)

    expect(screen.queryByLabelText(/log out/i)).toBeNull()
  })

  it('opens the profile channel on press, and renders the rest locked', async () => {
    const onOpenProfile = jest.fn()
    await renderWithProviders(<HomeScreen user={USER} onOpenProfile={onOpenProfile} />)

    fireEvent.press(screen.getByLabelText('Profile'))
    expect(onOpenProfile).toHaveBeenCalledTimes(1)

    expect(screen.getByLabelText('Stands, coming soon')).toBeTruthy()
    expect(screen.getByLabelText('Devil Fruits, coming soon')).toBeTruthy()
    expect(screen.getByLabelText('Powers, coming soon')).toBeTruthy()
  })
})
