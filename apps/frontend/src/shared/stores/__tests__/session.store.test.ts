import { getSessionToken, useSessionStore } from '../session.store'

const USER = {
  id: 'user-1',
  email: 'jotaro@example.com',
  username: 'OriolO',
  completeName: 'Jotaro Kujo',
  picture: null as string | null,
  role: 'REGULAR' as const,
  language: 'en-GB' as const,
}

const SESSION = {
  accessToken: 'token-123',
  expiresAt: new Date(0).toISOString(),
  user: USER,
}

// expo-secure-store is mocked in jest.setup.ts to resolve null/undefined for
// every call, which is enough here — this suite is about the store's own
// state transitions, not the storage backend itself.
describe('useSessionStore', () => {
  beforeEach(async () => {
    await useSessionStore.getState().clearSession()
    useSessionStore.setState({ isHydrated: false })
  })

  it('starts unhydrated with no session', () => {
    expect(useSessionStore.getState().isHydrated).toBe(false)
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('hydrate() with nothing stored settles to no session, but hydrated', async () => {
    await useSessionStore.getState().hydrate()

    expect(useSessionStore.getState().isHydrated).toBe(true)
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('setSession() stores the session and getSessionToken() reflects it', async () => {
    await useSessionStore.getState().setSession(SESSION)

    expect(useSessionStore.getState().session).toEqual(SESSION)
    expect(getSessionToken()).toBe('token-123')
  })

  it('clearSession() removes the session and getSessionToken() returns null', async () => {
    await useSessionStore.getState().setSession(SESSION)
    await useSessionStore.getState().clearSession()

    expect(useSessionStore.getState().session).toBeNull()
    expect(getSessionToken()).toBeNull()
  })
})
