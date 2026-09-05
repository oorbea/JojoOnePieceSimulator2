// Mocked before the import below so refresh.ts's own module-scope
// `axios.create(...)` call picks up the mock, not the real axios.
const mockPost = jest.fn()

// `post` is a thin function wrapper (not `post: mockPost` directly) so the
// lookup of `mockPost` happens lazily at CALL time rather than being frozen
// into the returned object the moment `axios.create()` runs - which, since
// this factory is hoisted above the `const mockPost = jest.fn()` below by
// babel-plugin-jest-hoist, would otherwise capture `undefined`.
jest.mock('axios', () => ({
  __esModule: true,
  default: {
    create: () => ({ post: (...args: unknown[]) => mockPost(...args) }),
  },
}))

jest.mock('@/shared/config/env', () => ({
  env: { EXPO_PUBLIC_API_URL: 'http://localhost/api/v1' },
}))

// This suite runs under the jest-expo "native" project (Platform.OS !==
// 'web'), so refresh.ts's doRefresh() reads a stored refresh token before it
// will even attempt the POST - give it one so the interesting behavior
// (single-flight, error handling) is what's actually under test here.
jest.mock('@/shared/lib/secure-storage', () => ({
  secureStorage: {
    getItem: jest.fn().mockResolvedValue('stored-refresh-token'),
    setItem: jest.fn().mockResolvedValue(undefined),
    removeItem: jest.fn().mockResolvedValue(undefined),
  },
}))

import { refreshSession } from '../refresh'

const USER_RESPONSE = {
  id: 'user-1',
  email: 'jotaro@example.com',
  username: 'OriolO',
  completeName: 'Jotaro Kujo',
  avatar: '',
  avatarThumb: '',
  avatarStatus: 'NONE',
  role: 'REGULAR',
  language: 'en-GB',
}

describe('refreshSession', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('resolves with the fresh access token and mapped user on success', async () => {
    mockPost.mockResolvedValue({
      data: { accessToken: 'fresh-token', tokenType: 'Bearer', expiresAt: '2026-09-05T00:00:00Z', user: USER_RESPONSE },
    })

    const result = await refreshSession()

    expect(result).toEqual({
      accessToken: 'fresh-token',
      user: {
        id: 'user-1',
        email: 'jotaro@example.com',
        username: 'OriolO',
        completeName: 'Jotaro Kujo',
        picture: null,
        role: 'REGULAR',
        language: 'en-GB',
      },
    })
  })

  it('collapses N concurrent calls while one is in flight into exactly one POST /auth/refresh', async () => {
    let resolvePost!: (value: unknown) => void
    mockPost.mockReturnValue(
      new Promise((resolve) => {
        resolvePost = resolve
      })
    )

    const calls = [refreshSession(), refreshSession(), refreshSession()]
    resolvePost({
      data: { accessToken: 'fresh-token', tokenType: 'Bearer', expiresAt: '2026-09-05T00:00:00Z', user: USER_RESPONSE },
    })
    const results = await Promise.all(calls)

    expect(mockPost).toHaveBeenCalledTimes(1)
    results.forEach((result) => expect(result?.accessToken).toBe('fresh-token'))
  })

  it('resolves to null (never throws) on a failed refresh (e.g. 401)', async () => {
    mockPost.mockRejectedValue({ response: { status: 401 } })

    await expect(refreshSession()).resolves.toBeNull()
  })

  it('resolves to null (never throws) on a network error', async () => {
    mockPost.mockRejectedValue(new Error('Network Error'))

    await expect(refreshSession()).resolves.toBeNull()
  })

  it('allows a fresh call after a prior one has settled', async () => {
    mockPost.mockRejectedValueOnce(new Error('Network Error'))
    await refreshSession()

    mockPost.mockResolvedValueOnce({
      data: { accessToken: 'second-token', tokenType: 'Bearer', expiresAt: '2026-09-05T00:00:00Z', user: USER_RESPONSE },
    })
    const result = await refreshSession()

    expect(mockPost).toHaveBeenCalledTimes(2)
    expect(result?.accessToken).toBe('second-token')
  })
})
