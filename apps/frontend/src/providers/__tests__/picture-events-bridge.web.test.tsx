/**
 * PictureEventsBridge mints a ticket before opening its EventSource, is
 * admin-gated, and must never retry forever against a permanent rejection
 * (403 - not an admin) the way the old ?token=<jwt> loop did.
 *
 * The `.web.test.tsx` suffix routes this to the jsdom "logic" jest project
 * (see jest.config.js) - it needs a real `document`/EventSource stand-in and
 * `Platform.OS === 'web'`, neither of which the native project provides.
 */
import { render } from '@testing-library/react-native'
import { createElement } from 'react'

import { PictureEventsBridge } from '@/providers/picture-events-bridge'
import { useSessionStore } from '@/shared/stores/session.store'

jest.mock('@/shared/config/env', () => ({
  env: {
    EXPO_PUBLIC_API_URL: 'http://localhost/api/v1',
    EXPO_PUBLIC_SOCKET_URL: 'ws://localhost/api/v1',
    EXPO_PUBLIC_BUILD_ID: 'test',
  },
}))

const mockInvalidateQueries = jest.fn()
// A stable object, not a fresh literal per render: the effect's own
// dependency array includes `queryClient`, and the real useQueryClient()
// always returns the same client instance - a mock returning a new object
// every render would restart the effect (and re-mint) on every re-render,
// which is not what production code does.
const mockQueryClient = { invalidateQueries: mockInvalidateQueries }
jest.mock('@tanstack/react-query', () => ({
  useQueryClient: () => mockQueryClient,
}))

jest.mock('@/shared/api/etag', () => ({ clearEtags: jest.fn() }))

const mockMintEventsTicket = jest.fn()
jest.mock('@/shared/api/stream-tickets', () => ({
  mintEventsTicket: (...args: unknown[]) => mockMintEventsTicket(...args),
}))

// Minimal fake EventSource: captures every instance so a test can drive
// onopen/onerror directly, mirroring FakeWebSocket in
// game-socket.store.test.ts. jsdom has no built-in EventSource.
class FakeEventSource {
  static instances: FakeEventSource[] = []
  url: string
  onopen: (() => void) | null = null
  onerror: (() => void) | null = null
  listeners: Record<string, ((event: { data: string }) => void)[]> = {}
  closed = false

  constructor(url: string) {
    this.url = url
    FakeEventSource.instances.push(this)
  }

  addEventListener(type: string, cb: (event: { data: string }) => void) {
    this.listeners[type] ??= []
    this.listeners[type].push(cb)
  }

  close() {
    this.closed = true
  }
}
;(globalThis as { EventSource?: unknown }).EventSource = FakeEventSource

function httpError(status: number) {
  return { response: { status }, message: `http ${status}` }
}

function adminSession() {
  return {
    accessToken: 'tok',
    expiresAt: '2100-01-01',
    user: {
      id: 'u1',
      email: 'a@b.com',
      username: 'a',
      completeName: 'A',
      picture: null,
      role: 'ADMIN' as const,
      language: 'en-GB' as const,
    },
  }
}

async function flush() {
  await Promise.resolve()
  await Promise.resolve()
}

beforeEach(() => {
  FakeEventSource.instances = []
  mockMintEventsTicket.mockReset()
  mockMintEventsTicket.mockResolvedValue('test-ticket')
  mockInvalidateQueries.mockClear()
  useSessionStore.setState({ session: null, isHydrated: true })
})

describe('PictureEventsBridge', () => {
  it('mints a ticket and opens the stream with it in the query string, not a JWT', async () => {
    useSessionStore.setState({ session: adminSession(), isHydrated: true })
    render(createElement(PictureEventsBridge))
    await flush()

    expect(mockMintEventsTicket).toHaveBeenCalledTimes(1)
    expect(FakeEventSource.instances).toHaveLength(1)
    expect(FakeEventSource.instances[0].url).toBe('http://localhost/api/v1/events?ticket=test-ticket')
  })

  it('does not mount for a non-admin session', async () => {
    useSessionStore.setState({
      session: { ...adminSession(), user: { ...adminSession().user, role: 'REGULAR' } },
      isHydrated: true,
    })
    render(createElement(PictureEventsBridge))
    await flush()

    expect(mockMintEventsTicket).not.toHaveBeenCalled()
    expect(FakeEventSource.instances).toHaveLength(0)
  })

  it('does not mint without a session', async () => {
    render(createElement(PictureEventsBridge))
    await flush()

    expect(mockMintEventsTicket).not.toHaveBeenCalled()
  })

  it('a 403 minting a ticket stops for good instead of re-minting forever', async () => {
    jest.useFakeTimers()
    try {
      mockMintEventsTicket.mockRejectedValueOnce(httpError(403))
      useSessionStore.setState({ session: adminSession(), isHydrated: true })
      render(createElement(PictureEventsBridge))
      await flush()

      expect(mockMintEventsTicket).toHaveBeenCalledTimes(1)
      expect(FakeEventSource.instances).toHaveLength(0)

      // The old ?token= loop re-minted every 2s->30s forever against a 403
      // stream - advancing well past that window must not trigger a retry.
      jest.advanceTimersByTime(60_000)
      await flush()
      expect(mockMintEventsTicket).toHaveBeenCalledTimes(1)
    } finally {
      jest.useRealTimers()
    }
  })

  it('a network error minting a ticket backs off and re-mints', async () => {
    jest.useFakeTimers()
    try {
      mockMintEventsTicket.mockRejectedValueOnce(new Error('network down'))
      useSessionStore.setState({ session: adminSession(), isHydrated: true })
      render(createElement(PictureEventsBridge))
      await flush()

      expect(mockMintEventsTicket).toHaveBeenCalledTimes(1)
      expect(FakeEventSource.instances).toHaveLength(0)

      jest.advanceTimersByTime(2_000)
      await flush()

      expect(mockMintEventsTicket).toHaveBeenCalledTimes(2)
      expect(FakeEventSource.instances).toHaveLength(1)
    } finally {
      jest.useRealTimers()
    }
  })

  it('a reconnect after a prior successful connection invalidates all three query keys', async () => {
    jest.useFakeTimers()
    try {
      useSessionStore.setState({ session: adminSession(), isHydrated: true })
      render(createElement(PictureEventsBridge))
      await flush()

      const first = FakeEventSource.instances[0]
      first.onopen?.()
      mockInvalidateQueries.mockClear()

      // onerror closes and schedules a re-mint/re-connect - the second
      // onopen is a reconnect (hasConnectedBeforeRef already true), which
      // must trigger the safety-net resync.
      first.onerror?.()
      jest.advanceTimersByTime(2_000)
      await flush()
      const second = FakeEventSource.instances[1]
      second.onopen?.()

      expect(mockInvalidateQueries).toHaveBeenCalled()
    } finally {
      jest.useRealTimers()
    }
  })
})
