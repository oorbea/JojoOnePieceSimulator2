jest.mock('@/shared/config/env', () => ({
  env: {
    EXPO_PUBLIC_API_URL: 'http://localhost/api/v1',
    EXPO_PUBLIC_SOCKET_URL: 'ws://localhost/api/v1',
    EXPO_PUBLIC_BUILD_ID: 'test',
  },
}))

import { buildGameSocketUrl } from '@/features/game/lib/socket-url'

describe('buildGameSocketUrl', () => {
  it('builds the ws URL with the ticket in the query string, not "token"', () => {
    const url = buildGameSocketUrl('g1', 'abc123')
    expect(url).toBe('ws://localhost/api/v1/games/g1/ws?ticket=abc123')
  })

  it('URL-encodes the ticket', () => {
    const url = buildGameSocketUrl('g1', 'a+b/c=')
    expect(url).toBe(`ws://localhost/api/v1/games/g1/ws?ticket=${encodeURIComponent('a+b/c=')}`)
  })

  // The "EXPO_PUBLIC_SOCKET_URL unset" case needs the opposite env mock and
  // lives in its own file (socket-url.unconfigured.test.ts) - jest.mock's
  // module-level env stub above is shared by every test in this file, and
  // fighting that with isolateModules/doMock proved flakier than just
  // splitting the file.
})
