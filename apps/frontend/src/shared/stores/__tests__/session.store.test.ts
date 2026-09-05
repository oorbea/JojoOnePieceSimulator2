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
  user: USER,
}

const mockRefreshSession = jest.fn()
const mockPostLogout = jest.fn()

jest.mock('@/shared/api/refresh', () => ({
  refreshSession: () => mockRefreshSession(),
}))

jest.mock('@/features/auth/api/auth.api', () => ({
  postLogout: () => mockPostLogout(),
}))

// expo-secure-store is mocked in jest.setup.ts to resolve null/undefined for
// every call, which is enough here — this suite is about the store's own
// state transitions, not the storage backend itself.
describe('useSessionStore', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockPostLogout.mockResolvedValue(undefined)
    useSessionStore.setState({ session: null, isHydrated: false })
  })

  it('starts unhydrated with no session', () => {
    expect(useSessionStore.getState().isHydrated).toBe(false)
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('hydrate() with a successful refresh sets session and isHydrated', async () => {
    mockRefreshSession.mockResolvedValue({ accessToken: 'token-123', user: USER })

    await useSessionStore.getState().hydrate()

    expect(useSessionStore.getState().isHydrated).toBe(true)
    expect(useSessionStore.getState().session).toEqual(SESSION)
  })

  it('hydrate() with no valid refresh token settles to no session, but hydrated', async () => {
    mockRefreshSession.mockResolvedValue(null)

    await useSessionStore.getState().hydrate()

    expect(useSessionStore.getState().isHydrated).toBe(true)
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('hydrate() never throws even if refreshSession rejects unexpectedly, and still settles isHydrated', async () => {
    mockRefreshSession.mockRejectedValue(new Error('network exploded'))

    await expect(useSessionStore.getState().hydrate()).resolves.toBeUndefined()

    expect(useSessionStore.getState().isHydrated).toBe(true)
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('setSession() stores the session synchronously and getSessionToken() reflects it', () => {
    useSessionStore.getState().setSession(SESSION)

    expect(useSessionStore.getState().session).toEqual(SESSION)
    expect(getSessionToken()).toBe('token-123')
  })

  it('clearSession() calls postLogout and clears the session', async () => {
    useSessionStore.getState().setSession(SESSION)

    await useSessionStore.getState().clearSession()

    expect(mockPostLogout).toHaveBeenCalledTimes(1)
    expect(useSessionStore.getState().session).toBeNull()
    expect(getSessionToken()).toBeNull()
  })

  it('clearSession() clears local state even if postLogout rejects', async () => {
    useSessionStore.getState().setSession(SESSION)
    mockPostLogout.mockRejectedValue(new Error('network exploded'))

    await expect(useSessionStore.getState().clearSession()).resolves.toBeUndefined()

    expect(useSessionStore.getState().session).toBeNull()
  })
})
