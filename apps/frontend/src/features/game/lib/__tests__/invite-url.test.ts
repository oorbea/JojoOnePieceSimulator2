const mockEnv: { EXPO_PUBLIC_WEB_ORIGIN?: string } = {}

jest.mock('@/shared/config/env', () => ({
  get env() {
    return mockEnv
  },
}))

import { buildInviteUrl, inviteOrigin } from '@/features/game/lib/invite-url'

describe('inviteOrigin / buildInviteUrl', () => {
  const originalLocation = window.location

  afterEach(() => {
    delete mockEnv.EXPO_PUBLIC_WEB_ORIGIN
    Object.defineProperty(window, 'location', { value: originalLocation, writable: true })
  })

  it('prefers EXPO_PUBLIC_WEB_ORIGIN when configured', () => {
    mockEnv.EXPO_PUBLIC_WEB_ORIGIN = 'https://jojoonepiece.example'

    expect(inviteOrigin()).toBe('https://jojoonepiece.example')
    expect(buildInviteUrl('abc123')).toBe('https://jojoonepiece.example/join/abc123')
  })

  it('falls back to window.location.origin on web when unconfigured', () => {
    Object.defineProperty(window, 'location', {
      value: { ...originalLocation, origin: 'https://example.com' },
      writable: true,
    })

    expect(inviteOrigin()).toBe('https://example.com')
    expect(buildInviteUrl('tok')).toBe('https://example.com/join/tok')
  })

  it('rejects a non-https configured origin that is not localhost', () => {
    mockEnv.EXPO_PUBLIC_WEB_ORIGIN = 'http://evil.example'

    expect(inviteOrigin()).toBeNull()
    expect(buildInviteUrl('tok')).toBeNull()
  })

  it('accepts http://localhost as a dev exception', () => {
    mockEnv.EXPO_PUBLIC_WEB_ORIGIN = 'http://localhost:3000'

    expect(inviteOrigin()).toBe('http://localhost:3000')
  })

  it('accepts http://127.0.0.1 as a dev exception', () => {
    mockEnv.EXPO_PUBLIC_WEB_ORIGIN = 'http://127.0.0.1:3000'

    expect(inviteOrigin()).toBe('http://127.0.0.1:3000')
  })

  it('URL-encodes the token', () => {
    mockEnv.EXPO_PUBLIC_WEB_ORIGIN = 'https://example.com'

    expect(buildInviteUrl('a b/c')).toBe('https://example.com/join/a%20b%2Fc')
  })

  it('returns null for a malformed configured origin', () => {
    mockEnv.EXPO_PUBLIC_WEB_ORIGIN = 'not a url'

    expect(inviteOrigin()).toBeNull()
  })
})
