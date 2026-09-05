import axios from 'axios'

import { registerInterceptors } from '../interceptors'
import { etagCacheKey, getCachedResponse, rememberResponse, clearEtags } from '../etag'
import { useSessionStore } from '@/shared/stores/session.store'
import { useLanguageStore } from '@/shared/stores/language.store'

const SESSION = {
  accessToken: 'token-123',
  user: {
    id: 'user-1',
    email: 'jotaro@example.com',
    username: 'OriolO',
    completeName: 'Jotaro Kujo',
    picture: null,
    role: 'REGULAR' as const,
    language: 'en-GB' as const,
  },
}

const REFRESHED_USER = { ...SESSION.user, username: 'RefreshedOriolO' }

const mockRefreshSession = jest.fn()
const mockPostLogout = jest.fn()

jest.mock('@/shared/api/refresh', () => ({
  refreshSession: () => mockRefreshSession(),
}))

jest.mock('@/features/auth/api/auth.api', () => ({
  postLogout: () => mockPostLogout(),
}))

// axios doesn't expose a public way to invoke a registered interceptor in
// isolation — reaching into the manager's internal `handlers` array is the
// standard workaround for testing them without a real HTTP round trip.
type InterceptorHandler = { fulfilled: (v: any) => any; rejected?: (e: any) => any }

function requestHandler(client: ReturnType<typeof axios.create>): InterceptorHandler {
  return (client.interceptors.request as unknown as { handlers: InterceptorHandler[] }).handlers[0]
}

function responseHandler(client: ReturnType<typeof axios.create>): InterceptorHandler {
  return (client.interceptors.response as unknown as { handlers: InterceptorHandler[] }).handlers[0]
}

describe('registerInterceptors', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockPostLogout.mockResolvedValue(undefined)
    useSessionStore.setState({ session: null, isHydrated: false })
    clearEtags()
  })

  it("attaches the session token as a Bearer header when there's a session", async () => {
    useSessionStore.getState().setSession(SESSION)
    const client = axios.create()
    registerInterceptors(client)

    const config = await requestHandler(client).fulfilled({ headers: {} })
    expect(config.headers.Authorization).toBe('Bearer token-123')
  })

  it('sends no Authorization header when there is no session', async () => {
    const client = axios.create()
    registerInterceptors(client)

    const config = await requestHandler(client).fulfilled({ headers: {} })
    expect(config.headers.Authorization).toBeUndefined()
  })

  it('attaches If-None-Match on GET when a prior ETag+body is known for that URL', async () => {
    rememberResponse(etagCacheKey('/profile/me'), 'W/"abc123"', { id: 'me' })
    const client = axios.create()
    registerInterceptors(client)

    const config = await requestHandler(client).fulfilled({
      method: 'get',
      url: '/profile/me',
      headers: {},
    })
    expect(config.headers['If-None-Match']).toBe('W/"abc123"')
  })

  it('does not attach If-None-Match when only an ETag (no body) would be known', async () => {
    // Regression guard: an ETag must never be sent without the body it was
    // issued for - otherwise a 304 response has nothing to resolve to and
    // callers end up with an empty body instead of the real data.
    const client = axios.create()
    registerInterceptors(client)

    const config = await requestHandler(client).fulfilled({
      method: 'get',
      url: '/stands',
      headers: {},
    })
    expect(config.headers['If-None-Match']).toBeUndefined()
  })

  it('remembers the ETag and body from a successful GET response', async () => {
    const client = axios.create()
    registerInterceptors(client)

    const config = { method: 'get', url: '/profile/other', __etagCacheKey: etagCacheKey('/profile/other') }
    responseHandler(client).fulfilled({
      config,
      status: 200,
      headers: { etag: 'W/"fresh"' },
      data: { id: 'other' },
    })

    expect(getCachedResponse(etagCacheKey('/profile/other'))).toEqual({
      etag: 'W/"fresh"',
      data: { id: 'other' },
    })
  })

  it('fills in the cached body when the server answers 304', async () => {
    const client = axios.create()
    registerInterceptors(client)

    const key = etagCacheKey('/stands')
    rememberResponse(key, 'W/"list-1"', [{ id: 'stand-1' }])

    const requestConfig = await requestHandler(client).fulfilled({
      method: 'get',
      url: '/stands',
      headers: {},
    })
    expect(requestConfig.headers['If-None-Match']).toBe('W/"list-1"')

    const response = responseHandler(client).fulfilled({
      config: requestConfig,
      status: 304,
      headers: {},
      data: '',
    })

    expect(response.status).toBe(200)
    expect(response.data).toEqual([{ id: 'stand-1' }])
  })

  it('keys the ETag cache by params so a filtered request cannot reuse an unfiltered body', async () => {
    const client = axios.create()
    registerInterceptors(client)

    rememberResponse(etagCacheKey('/stands'), 'W/"all"', [{ id: 'a' }, { id: 'b' }])

    const filteredConfig = await requestHandler(client).fulfilled({
      method: 'get',
      url: '/stands',
      params: { rarity: 'RARE' },
      headers: {},
    })
    expect(filteredConfig.headers['If-None-Match']).toBeUndefined()
  })

  it('keys the ETag cache by locale so a language switch cannot reuse a stale body', async () => {
    const client = axios.create()
    registerInterceptors(client)
    const originalLocale = useLanguageStore.getState().locale

    useLanguageStore.setState({ locale: 'en-GB' })
    rememberResponse(etagCacheKey('/stands'), 'W/"en"', [{ id: 'a', description: 'english' }])

    useLanguageStore.setState({ locale: 'es-ES' })
    const config = await requestHandler(client).fulfilled({
      method: 'get',
      url: '/stands',
      headers: {},
    })
    expect(config.headers['If-None-Match']).toBeUndefined()

    useLanguageStore.setState({ locale: originalLocale })
  })

  it('clearEtags drops cached bodies as well as ETags', async () => {
    const key = etagCacheKey('/profile/me')
    rememberResponse(key, 'W/"abc123"', { id: 'me' })
    clearEtags()
    expect(getCachedResponse(key)).toBeUndefined()
  })

  it('on a 401, refreshes once and retries the original request on success', async () => {
    useSessionStore.getState().setSession(SESSION)
    mockRefreshSession.mockResolvedValue({ accessToken: 'token-456', user: REFRESHED_USER })
    const client = axios.create()
    const mockRetry = jest.fn().mockResolvedValue({ status: 200, data: 'ok' })
    const clientWithSpy = Object.assign(
      (config: unknown) => mockRetry(config),
      client
    ) as unknown as typeof client
    registerInterceptors(clientWithSpy)

    const config = { url: '/games/123', headers: {} }
    const result = await responseHandler(clientWithSpy).rejected!({
      response: { status: 401 },
      config,
      message: 'Unauthorized',
    })

    expect(mockRefreshSession).toHaveBeenCalledTimes(1)
    expect(mockRetry).toHaveBeenCalledTimes(1)
    expect(mockRetry).toHaveBeenCalledWith(expect.objectContaining({ url: '/games/123', __retried: true }))
    expect(result).toEqual({ status: 200, data: 'ok' })
    // Session was updated in place with the refreshed token, not cleared.
    expect(useSessionStore.getState().session?.accessToken).toBe('token-456')
    expect(mockPostLogout).not.toHaveBeenCalled()
  })

  it('a 401 on /auth/refresh itself does not trigger another refresh', async () => {
    useSessionStore.getState().setSession(SESSION)
    const client = axios.create()
    registerInterceptors(client)

    await expect(
      responseHandler(client).rejected!({
        response: { status: 401 },
        config: { url: '/auth/refresh', headers: {} },
        message: 'Unauthorized',
      })
    ).rejects.toBeTruthy()

    expect(mockRefreshSession).not.toHaveBeenCalled()
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('a second 401 on the retried request does not loop again', async () => {
    useSessionStore.getState().setSession(SESSION)
    mockRefreshSession.mockResolvedValue({ accessToken: 'token-456', user: REFRESHED_USER })
    const client = axios.create()
    registerInterceptors(client)

    const config = { url: '/games/123', headers: {}, __retried: true }
    await expect(
      responseHandler(client).rejected!({
        response: { status: 401 },
        config,
        message: 'Unauthorized',
      })
    ).rejects.toBeTruthy()

    expect(mockRefreshSession).not.toHaveBeenCalled()
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('clears the session on a 401 when refreshSession resolves to null', async () => {
    useSessionStore.getState().setSession(SESSION)
    mockRefreshSession.mockResolvedValue(null)
    const client = axios.create()
    registerInterceptors(client)

    await expect(
      responseHandler(client).rejected!({
        response: { status: 401 },
        config: { url: '/games/123', headers: {} },
        message: 'Unauthorized',
      })
    ).rejects.toBeTruthy()

    expect(mockRefreshSession).toHaveBeenCalledTimes(1)
    expect(useSessionStore.getState().session).toBeNull()
  })

  it('leaves the session alone on a non-401 error', async () => {
    useSessionStore.getState().setSession(SESSION)
    const client = axios.create()
    registerInterceptors(client)

    await expect(
      responseHandler(client).rejected!({ response: { status: 500 }, message: 'Server error' })
    ).rejects.toBeTruthy()

    expect(useSessionStore.getState().session).toEqual(SESSION)
  })
})
