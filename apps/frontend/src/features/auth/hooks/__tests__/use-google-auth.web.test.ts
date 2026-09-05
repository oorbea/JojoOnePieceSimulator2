/**
 * The web branch of useGoogleAuth: the state/nonce round trip across Google's
 * full-page redirect, the hand-rolled base64url JWT payload decoder, and the
 * authorize URL the hook navigates to.
 *
 * None of those three are exported, so everything here is driven through the
 * hook itself. The `.web.test.ts` suffix routes this file to the jsdom
 * "logic" jest project (see jest.config.js) - it needs a real `document`,
 * `sessionStorage`, `atob` and `Platform.OS === 'web'`, none of which the
 * native project provides.
 *
 * Why this is worth pinning: `state` is the only thing standing between a
 * user and an attacker-supplied id_token pasted into this app's redirect URL,
 * and `nonce` is what ties the returned token to the request this tab made.
 * Both are compared here in the client, not on the backend, so a regression
 * is invisible to every Go test in the repo.
 */
import { act, render } from '@testing-library/react-native'
import { createElement, useEffect } from 'react'

import { useGoogleAuth } from '@/features/auth/hooks/use-google-auth'

const WEB_STATE_KEY = 'jojo_google_auth_state'
const WEB_NONCE_KEY = 'jojo_google_auth_nonce'
const WEB_CLIENT_ID = 'test-web-client-id.apps.googleusercontent.com'

// Every name a jest.mock factory closes over has to start with `mock` -
// babel-plugin-jest-hoist lifts these calls above the imports above and
// rejects any other out-of-scope reference.
const mockPromptAsync = jest.fn()
const mockPostGoogleAuth = jest.fn()
const mockSetSession = jest.fn()

jest.mock('@/shared/config/env', () => ({
  env: {
    EXPO_PUBLIC_API_URL: 'http://localhost/api/v1',
    EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID: 'test-web-client-id.apps.googleusercontent.com',
    EXPO_PUBLIC_BUILD_ID: 'test',
  },
}))

// expo-auth-session's Google provider only drives the native branch; on web
// the hook must never call promptAsync. Stubbed rather than mocked through,
// so the real module (and its expo-modules-core native bridge) never loads.
jest.mock('expo-auth-session/providers/google', () => ({
  useAuthRequest: () => [{ url: 'unused' }, null, mockPromptAsync],
}))

jest.mock('expo-web-browser', () => ({ maybeCompleteAuthSession: jest.fn() }))

jest.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

jest.mock('@/features/auth/api/auth.api', () => ({
  postGoogleAuth: (idToken: string) => mockPostGoogleAuth(idToken),
}))

jest.mock('@/shared/stores/session.store', () => ({
  useSessionStore: (selector: (state: { setSession: unknown }) => unknown) =>
    selector({ setSession: mockSetSession }),
  // Real implementation, not a mock - it's a pure mapping function with no
  // dependency on store state, and the hook imports it from this same
  // (mocked) module.
  fromUserResponse: (user: {
    id: string
    email: string
    username: string
    completeName: string
    avatar: string
    role: string
    language: string
  }) => ({
    id: user.id,
    email: user.email,
    username: user.username,
    completeName: user.completeName,
    picture: user.avatar || null,
    role: user.role,
    language: user.language,
  }),
}))

type Api = ReturnType<typeof useGoogleAuth>

let captured: Api | null = null

function publish(next: Api) {
  captured = next
}

function api(): Api {
  if (!captured) throw new Error('Probe has not rendered yet')
  return captured
}

// Same probe shape use-roving-group.web.test.tsx established: render `null`
// (this project maps react-native -> react-native-web, so a real RN element
// would confuse the native test renderer) and hand the hook's value out from
// an effect with no dependency array, so it re-publishes after every commit.
function Probe() {
  const value = useGoogleAuth()
  useEffect(() => {
    publish(value)
  })
  return null
}

async function mount() {
  captured = null
  await act(async () => {
    render(createElement(Probe))
  })
}

// The hook defers its redirect handling into an async IIFE whose body awaits
// further promises (postGoogleAuth, then setSession). Yielding to a real
// timer rather than a single `await Promise.resolve()` drains that whole
// microtask chain in one go, instead of one link of it per call.
async function flush() {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0))
  })
}

// base64url-encodes a JSON payload the way Google does: '+'/'/' swapped for
// '-'/'_' and the '=' padding stripped. decodeJwtPayload has to undo both.
function base64UrlJson(payload: Record<string, unknown>): string {
  return btoa(JSON.stringify(payload)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

function idTokenWithNonce(nonce: string): string {
  return `header-segment.${base64UrlJson({ nonce, sub: '1234567890' })}.signature-segment`
}

function loginResponse(overrides: Record<string, unknown> = {}) {
  return {
    accessToken: 'backend-access-token',
    expiresAt: '2026-09-03T00:00:00Z',
    user: {
      id: 'user-1',
      email: 'jotaro@example.com',
      username: 'jotaro',
      completeName: 'Jotaro Kujo',
      avatar: 'https://r2.test/avatar.webp',
      role: 'REGULAR',
      language: 'en-GB',
    },
    ...overrides,
  }
}

// window.location is an own accessor property on jsdom's window and is
// configurable, so it can be swapped for a plain object. That is the only way
// to both read origin/pathname deterministically and observe assign(): jsdom's
// real Location refuses cross-origin navigation and cannot be spied on.
type FakeLocation = { origin: string; pathname: string; hash: string; assign: jest.Mock }

let fakeLocation: FakeLocation
let replaceState: jest.Mock

function setLocation(hash: string) {
  fakeLocation = { origin: 'https://app.test', pathname: '/login', hash, assign: jest.fn() }
  Object.defineProperty(window, 'location', {
    configurable: true,
    writable: true,
    value: fakeLocation,
  })
}

beforeEach(() => {
  jest.clearAllMocks()
  window.sessionStorage.clear()
  setLocation('')

  replaceState = jest.fn()
  Object.defineProperty(window, 'history', {
    configurable: true,
    writable: true,
    value: { replaceState },
  })

  // Deterministic state/nonce, so the URL assertions can name exact values.
  const randomUUID = jest
    .fn()
    .mockReturnValueOnce('state-uuid-0000')
    .mockReturnValueOnce('nonce-uuid-1111')
  Object.defineProperty(window, 'crypto', {
    configurable: true,
    writable: true,
    value: { randomUUID },
  })
})

describe('useGoogleAuth - signIn on web', () => {
  it('stores a fresh state and nonce before navigating away', async () => {
    await mount()

    await act(async () => {
      api().signIn()
    })

    expect(window.sessionStorage.getItem(WEB_STATE_KEY)).toBe('state-uuid-0000')
    expect(window.sessionStorage.getItem(WEB_NONCE_KEY)).toBe('nonce-uuid-1111')
    // Web never goes through expo-auth-session - that is the whole point of
    // the redirect flow (accounts.google.com's COOP header severs the popup).
    expect(mockPromptAsync).not.toHaveBeenCalled()
  })

  it('builds the Google authorize URL from the stored state and nonce', async () => {
    await mount()

    await act(async () => {
      api().signIn()
    })

    expect(fakeLocation.assign).toHaveBeenCalledTimes(1)
    const url = new URL(fakeLocation.assign.mock.calls[0][0] as string)

    expect(`${url.origin}${url.pathname}`).toBe('https://accounts.google.com/o/oauth2/v2/auth')
    expect(url.searchParams.get('client_id')).toBe(WEB_CLIENT_ID)
    // origin + pathname only: a redirect_uri carrying a query string or hash
    // would not match the one registered in Google Cloud Console.
    expect(url.searchParams.get('redirect_uri')).toBe('https://app.test/login')
    expect(url.searchParams.get('response_type')).toBe('id_token')
    expect(url.searchParams.get('scope')).toBe('openid email profile')
    expect(url.searchParams.get('prompt')).toBe('select_account')
    expect(url.searchParams.get('state')).toBe('state-uuid-0000')
    expect(url.searchParams.get('nonce')).toBe('nonce-uuid-1111')
  })

  it('reports ready immediately on web, without waiting for an auth request', async () => {
    await mount()

    expect(api().isReady).toBe(true)
  })
})

describe('useGoogleAuth - returning from the Google redirect', () => {
  // Puts the hook in the state it would be in right after signIn navigated
  // away: state/nonce in sessionStorage, Google's response in the URL hash.
  async function returnFromGoogle(hash: string, stored?: { state?: string; nonce?: string }) {
    if (stored?.state !== undefined) window.sessionStorage.setItem(WEB_STATE_KEY, stored.state)
    if (stored?.nonce !== undefined) window.sessionStorage.setItem(WEB_NONCE_KEY, stored.nonce)
    setLocation(hash)
    await mount()
    await flush()
  }

  const matching = { state: 'state-uuid-0000', nonce: 'nonce-uuid-1111' }

  it('completes sign-in when state and nonce both match', async () => {
    mockPostGoogleAuth.mockResolvedValue(loginResponse())
    const idToken = idTokenWithNonce('nonce-uuid-1111')

    await returnFromGoogle(`#id_token=${idToken}&state=state-uuid-0000`, matching)

    expect(mockPostGoogleAuth).toHaveBeenCalledWith(idToken)
    // setSession is now synchronous (no more expiresAt in the persisted
    // shape - the store only ever holds the access token + user in memory).
    expect(mockSetSession).toHaveBeenCalledWith({
      accessToken: 'backend-access-token',
      user: expect.objectContaining({ id: 'user-1', picture: 'https://r2.test/avatar.webp' }),
    })
    expect(api().error).toBeNull()
    // The id_token must not be left sitting in the address bar / history.
    expect(replaceState).toHaveBeenCalledWith(null, '', '/login')
  })

  it('does not persist any refresh token on web - the cookie is the only transport', async () => {
    mockPostGoogleAuth.mockResolvedValue(loginResponse())
    const idToken = idTokenWithNonce('nonce-uuid-1111')

    await returnFromGoogle(`#id_token=${idToken}&state=state-uuid-0000`, matching)

    // secureStorage is native-only (throws on web) - if the hook accidentally
    // tried to call it here for the web path, this suite would blow up.
    expect(mockSetSession).toHaveBeenCalled()
  })

  it('burns the stored state and nonce so a replayed URL cannot reuse them', async () => {
    mockPostGoogleAuth.mockResolvedValue(loginResponse())

    await returnFromGoogle(
      `#id_token=${idTokenWithNonce('nonce-uuid-1111')}&state=state-uuid-0000`,
      matching
    )

    expect(window.sessionStorage.getItem(WEB_STATE_KEY)).toBeNull()
    expect(window.sessionStorage.getItem(WEB_NONCE_KEY)).toBeNull()
  })

  it('rejects a mismatched state without calling the backend', async () => {
    await returnFromGoogle(
      `#id_token=${idTokenWithNonce('nonce-uuid-1111')}&state=attacker-state`,
      matching
    )

    expect(mockPostGoogleAuth).not.toHaveBeenCalled()
    expect(mockSetSession).not.toHaveBeenCalled()
    expect(api().error?.message).toBe('auth.googleSignInError')
  })

  it('rejects a response with no state at all', async () => {
    await returnFromGoogle(`#id_token=${idTokenWithNonce('nonce-uuid-1111')}`, matching)

    expect(mockPostGoogleAuth).not.toHaveBeenCalled()
    expect(api().error?.message).toBe('auth.googleSignInError')
  })

  // The no-stored-state case matters most: it is what an id_token pasted
  // straight into this app's URL by someone else looks like.
  it('rejects a response when this tab never started a sign-in', async () => {
    await returnFromGoogle(`#id_token=${idTokenWithNonce('nonce-uuid-1111')}&state=state-uuid-0000`)

    expect(mockPostGoogleAuth).not.toHaveBeenCalled()
    expect(api().error?.message).toBe('auth.googleSignInError')
  })

  it('rejects a token whose nonce does not match the one this tab sent', async () => {
    await returnFromGoogle(
      `#id_token=${idTokenWithNonce('someone-elses-nonce')}&state=state-uuid-0000`,
      matching
    )

    expect(mockPostGoogleAuth).not.toHaveBeenCalled()
    expect(api().error?.message).toBe('auth.googleSignInError')
  })

  it('rejects a token carrying no nonce claim', async () => {
    const noNonce = `header.${base64UrlJson({ sub: '1234567890' })}.signature`

    await returnFromGoogle(`#id_token=${noNonce}&state=state-uuid-0000`, matching)

    expect(mockPostGoogleAuth).not.toHaveBeenCalled()
    expect(api().error?.message).toBe('auth.googleSignInError')
  })

  it('ignores a hash that carries no id_token, leaving the URL alone', async () => {
    await returnFromGoogle('#error=access_denied', matching)

    expect(mockPostGoogleAuth).not.toHaveBeenCalled()
    expect(api().error).toBeNull()
    expect(replaceState).not.toHaveBeenCalled()
    // Nothing was consumed, so a real redirect arriving later still works.
    expect(window.sessionStorage.getItem(WEB_STATE_KEY)).toBe('state-uuid-0000')
  })

  it('does nothing at all on a plain page load with no hash', async () => {
    await returnFromGoogle('')

    expect(mockPostGoogleAuth).not.toHaveBeenCalled()
    expect(replaceState).not.toHaveBeenCalled()
    expect(api().error).toBeNull()
  })
})

describe('useGoogleAuth - decodeJwtPayload', () => {
  // Payload byte lengths 0..3 mod 4 all have to survive the manual '='
  // repadding, and a payload whose base64 contains '+' / '/' has to survive
  // the base64url -> base64 swap. Both are exercised through the nonce check,
  // the decoder's only observable consumer.
  const nonces = [
    'a', // shortest - forces the most padding
    'ab',
    'abc',
    'abcd',
    'ÿþýü', // base64-encodes with '+' and '/', so Google sends '-' and '_'
    'nonce-with-dashes-and_underscores-1234567890',
  ]

  it.each(nonces)('round-trips a nonce through the base64url decoder: %s', async (nonce) => {
    mockPostGoogleAuth.mockResolvedValue(loginResponse())
    window.sessionStorage.setItem(WEB_STATE_KEY, 'state-uuid-0000')
    window.sessionStorage.setItem(WEB_NONCE_KEY, nonce)
    setLocation(`#id_token=${idTokenWithNonce(nonce)}&state=state-uuid-0000`)

    await mount()
    await flush()

    expect(mockPostGoogleAuth).toHaveBeenCalledTimes(1)
  })
})
