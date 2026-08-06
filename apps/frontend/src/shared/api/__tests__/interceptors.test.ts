import axios from 'axios'

import { registerInterceptors } from '../interceptors'
import { getKnownEtag, rememberEtag } from '../etag'
import { useSessionStore } from '@/shared/stores/session.store'

const SESSION = {
  accessToken: 'token-123',
  expiresAt: new Date(0).toISOString(),
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
  beforeEach(async () => {
    await useSessionStore.getState().clearSession()
  })

  it("attaches the session token as a Bearer header when there's a session", async () => {
    await useSessionStore.getState().setSession(SESSION)
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

  it('attaches If-None-Match on GET when a prior ETag is known for that URL', async () => {
    rememberEtag('/profile/me', 'W/"abc123"')
    const client = axios.create()
    registerInterceptors(client)

    const config = await requestHandler(client).fulfilled({
      method: 'get',
      url: '/profile/me',
      headers: {},
    })
    expect(config.headers['If-None-Match']).toBe('W/"abc123"')
  })

  it('remembers the ETag from a successful GET response', async () => {
    const client = axios.create()
    registerInterceptors(client)

    responseHandler(client).fulfilled({
      config: { method: 'get', url: '/profile/other' },
      headers: { etag: 'W/"fresh"' },
    })

    expect(getKnownEtag('/profile/other')).toBe('W/"fresh"')
  })

  // The backend issues a plain bearer JWT with no refresh token (see
  // session.store.ts's own comment) — a 401 means the token is dead for
  // good, and the only recovery is a fresh sign-in, so the interceptor's
  // whole job on a 401 is to drop the stale session.
  it('clears the session on a 401 response', async () => {
    await useSessionStore.getState().setSession(SESSION)
    const client = axios.create()
    registerInterceptors(client)

    await expect(
      responseHandler(client).rejected!({ response: { status: 401 }, message: 'Unauthorized' })
    ).rejects.toBeTruthy()

    expect(useSessionStore.getState().session).toBeNull()
  })

  it('leaves the session alone on a non-401 error', async () => {
    await useSessionStore.getState().setSession(SESSION)
    const client = axios.create()
    registerInterceptors(client)

    await expect(
      responseHandler(client).rejected!({ response: { status: 500 }, message: 'Server error' })
    ).rejects.toBeTruthy()

    expect(useSessionStore.getState().session).toEqual(SESSION)
  })
})
