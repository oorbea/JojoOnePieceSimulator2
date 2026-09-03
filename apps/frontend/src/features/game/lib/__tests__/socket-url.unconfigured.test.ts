jest.mock('@/shared/config/env', () => ({
  env: {
    EXPO_PUBLIC_API_URL: 'http://localhost/api/v1',
    EXPO_PUBLIC_SOCKET_URL: undefined,
    EXPO_PUBLIC_BUILD_ID: 'test',
  },
}))

import { buildGameSocketUrl } from '@/features/game/lib/socket-url'

// Split from socket-url.test.ts: that file's module-level env mock sets a
// real ws:// EXPO_PUBLIC_SOCKET_URL for its own tests, and jest.mock's env
// stub is shared by every test in a file - this one case needs the opposite.
describe('buildGameSocketUrl, EXPO_PUBLIC_SOCKET_URL unset', () => {
  it('returns null - the documented "realtime unavailable" signal', () => {
    expect(buildGameSocketUrl('g1', 'abc123')).toBeNull()
  })
})
