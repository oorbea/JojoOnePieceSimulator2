// EXPO_PUBLIC_SOCKET_URL is blank in the real .env (native realtime isn't
// deployed yet) - stub it here so buildGameSocketUrl doesn't short-circuit
// to null and skip opening a socket at all.
jest.mock('@/shared/config/env', () => ({
  env: { EXPO_PUBLIC_API_URL: 'http://localhost/api/v1', EXPO_PUBLIC_SOCKET_URL: 'ws://localhost/api/v1', EXPO_PUBLIC_BUILD_ID: 'test' },
}))

import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import { useSessionStore } from '@/shared/stores/session.store'

// Minimal fake WebSocket: captures the last instance so the test can drive
// onopen/onmessage/onclose directly, mirroring the real socket's lifecycle
// without a network. Injected via attach()'s socketFactory param.
class FakeWebSocket {
  static instances: FakeWebSocket[] = []
  static readonly OPEN = 1
  readyState = 0
  onopen: (() => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onerror: (() => void) | null = null
  onclose: (() => void) | null = null
  sent: string[] = []
  url: string

  constructor(url: string) {
    this.url = url
    FakeWebSocket.instances.push(this)
  }

  send(data: string) {
    this.sent.push(data)
  }

  close() {
    this.readyState = 3
    this.onclose?.()
  }

  open() {
    this.readyState = FakeWebSocket.OPEN
    this.onopen?.()
  }

  receive(frame: unknown) {
    this.onmessage?.({ data: JSON.stringify(frame) })
  }
}

// WebSocket.OPEN is read by store.send() as a bare global - stub it so the
// fake's readyState comparison matches.
;(globalThis as { WebSocket?: unknown }).WebSocket = { OPEN: 1 }

function factory(url: string) {
  return new FakeWebSocket(url) as unknown as WebSocket
}

beforeEach(() => {
  FakeWebSocket.instances = []
  useGameSocketStore.getState().reset()
  useSessionStore.setState({
    session: {
      accessToken: 'tok',
      expiresAt: '2100-01-01',
      user: {
        id: 'u1',
        email: 'a@b.com',
        username: 'a',
        completeName: 'A',
        picture: null,
        role: 'REGULAR',
        language: 'en-GB',
      },
    },
    isHydrated: true,
  })
})

afterEach(() => {
  useGameSocketStore.getState().reset()
})

describe('useGameSocketStore', () => {
  it('replaces the snapshot wholesale on STATE and never merges', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()

    const snapshotA = { game: { id: 'g1', locked: false }, you: { participantId: 'p1' } }
    ws.receive({ type: 'STATE', payload: snapshotA })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshotA)

    const snapshotB = { game: { id: 'g1', locked: true }, you: { participantId: 'p1' } }
    ws.receive({ type: 'STATE', payload: snapshotB })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshotB)
  })

  it('does not touch the snapshot on a delta frame', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    const snapshot = { game: { id: 'g1' }, you: { participantId: 'p1' } }
    ws.receive({ type: 'STATE', payload: snapshot })

    ws.receive({ type: 'PLAYER_JOINED', payload: { participantId: 'p2' } })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshot)
    expect(useGameSocketStore.getState().feed.length).toBe(1)
  })

  it('VOTE_CAST never touches the snapshot', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    const snapshot = { game: { id: 'g1' }, you: { participantId: 'p1' } }
    ws.receive({ type: 'STATE', payload: snapshot })

    ws.receive({ type: 'VOTE_CAST', payload: { roundIndex: 0, votesCast: 1 } })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshot)
    expect(useGameSocketStore.getState().feed.length).toBe(0)
  })

  it('sets terminal ABORTED on GAME_ABORTED and stops reconnecting', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'GAME_ABORTED', payload: { reason: 'host cancelled' } })
    expect(useGameSocketStore.getState().terminal).toEqual({ kind: 'ABORTED', reason: 'host cancelled' })

    ws.close()
    expect(useGameSocketStore.getState().status).toBe('closed')
  })

  it('sets terminal KICKED only when the kicked participant is self', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'STATE', payload: { game: { id: 'g1' }, you: { participantId: 'p1' } } })

    ws.receive({ type: 'PLAYER_KICKED', payload: { participantId: 'p2' } })
    expect(useGameSocketStore.getState().terminal).toBeNull()

    ws.receive({ type: 'PLAYER_KICKED', payload: { participantId: 'p1' } })
    expect(useGameSocketStore.getState().terminal).toEqual({ kind: 'KICKED' })
  })

  it('records lastError from an ERROR frame with its requestId', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'ERROR', requestId: 'req-1', payload: { error: 'nope', code: 'NOT_HOST' } })
    expect(useGameSocketStore.getState().lastError).toEqual({
      message: 'nope',
      code: 'NOT_HOST',
      requestId: 'req-1',
    })
  })

  it('ignores a malformed frame instead of throwing', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    expect(() => ws.onmessage?.({ data: 'not json' })).not.toThrow()
  })

  it('reuses one socket for two attach calls with the same gameId', () => {
    useGameSocketStore.getState().attach('g1', factory)
    useGameSocketStore.getState().attach('g1', factory)
    expect(FakeWebSocket.instances.length).toBe(1)
  })

  it('does not open a socket without a session token', () => {
    useSessionStore.setState({ session: null, isHydrated: true })
    useGameSocketStore.getState().attach('g1', factory)
    expect(FakeWebSocket.instances.length).toBe(0)
  })

  it('LOADOUTS_ASSIGNED bumps assignmentSeq without touching snapshot/feed', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    const snapshot = { game: { id: 'g1' }, you: { participantId: 'p1' } }
    ws.receive({ type: 'STATE', payload: snapshot })

    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0 } })

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)
    expect(useGameSocketStore.getState().live.assignedRoundIndex).toBe(0)
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshot)
    expect(useGameSocketStore.getState().feed.length).toBe(0)
  })

  it('LOADOUTS_ASSIGNED sets revealMs/revealEndsAt from the frame payload', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, revealMs: 18350 } })

    const live = useGameSocketStore.getState().live
    expect(live.revealMs).toBe(18350)
    expect(live.revealEndsAt).not.toBeNull()
    expect(live.revealEndsAt).toBeGreaterThanOrEqual(Date.now())
  })

  it('STATE adopts game.revealEndsAt when no reveal is already tracked (reconnect mid-sorteo)', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({
      type: 'STATE',
      payload: { game: { id: 'g1', revealEndsAt: '2100-01-01T00:00:20.000Z' }, you: { participantId: 'p1' } },
    })

    const live = useGameSocketStore.getState().live
    expect(live.revealEndsAt).toBe(Date.parse('2100-01-01T00:00:20.000Z'))
    expect(live.revealMs).not.toBeNull()
  })

  it('STATE does not override an already-tracked reveal deadline from LOADOUTS_ASSIGNED', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, revealMs: 18350 } })
    const revealEndsAtBefore = useGameSocketStore.getState().live.revealEndsAt

    ws.receive({
      type: 'STATE',
      payload: { game: { id: 'g1', revealEndsAt: '2100-01-01T00:00:20.000Z' }, you: { participantId: 'p1' } },
    })

    expect(useGameSocketStore.getState().live.revealEndsAt).toBe(revealEndsAtBefore)
  })

  it('bumps assignmentSeq to 2 across two assignment frames', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0 } })
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 1 } })

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(2)
    expect(useGameSocketStore.getState().live.assignedRoundIndex).toBe(1)
  })

  it('VOTING_OPENED populates live with roundIndex/closesAt, clears tiebreak, and clears the reveal deadline', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, revealMs: 18350 } })

    ws.receive({ type: 'VOTING_OPENED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' } })

    const live = useGameSocketStore.getState().live
    expect(live.votingRoundIndex).toBe(0)
    expect(live.votingClosesAt).toBe(Date.parse('2100-01-01T00:00:10.000Z'))
    expect(live.tiebreak).toBe(false)
    expect(live.revealMs).toBeNull()
    expect(live.revealEndsAt).toBeNull()
  })

  it('TIEBREAK_OPENED populates live the same way with tiebreak true', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'TIEBREAK_OPENED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' } })

    const live = useGameSocketStore.getState().live
    expect(live.votingRoundIndex).toBe(0)
    expect(live.votingClosesAt).toBe(Date.parse('2100-01-01T00:00:10.000Z'))
    expect(live.tiebreak).toBe(true)
  })

  it('ROUND_RESOLVED clears the countdown and still pushes to feed', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'VOTING_OPENED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' } })

    ws.receive({ type: 'ROUND_RESOLVED', payload: { roundIndex: 0, winner: 'A', decidedByCoinFlip: false } })

    const live = useGameSocketStore.getState().live
    expect(live.votingRoundIndex).toBeNull()
    expect(live.votingClosesAt).toBeNull()
    expect(useGameSocketStore.getState().feed.length).toBe(1)
  })

  it('markAssignmentRevealed sets revealedAssignmentSeq to the current assignmentSeq', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0 } })
    expect(useGameSocketStore.getState().live.assignmentSeq).toBeGreaterThan(
      useGameSocketStore.getState().live.revealedAssignmentSeq
    )

    useGameSocketStore.getState().markAssignmentRevealed()

    const live = useGameSocketStore.getState().live
    expect(live.revealedAssignmentSeq).toBe(live.assignmentSeq)
  })

  it('reset() clears live back to its zero/null initial value', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0 } })

    useGameSocketStore.getState().reset()

    expect(useGameSocketStore.getState().live).toEqual({
      assignmentSeq: 0,
      revealedAssignmentSeq: 0,
      assignedRoundIndex: null,
      revealMs: null,
      revealEndsAt: null,
      votingRoundIndex: null,
      votingClosesAt: null,
      tiebreak: false,
    })
  })

  it('a gameId change clears live', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws1 = FakeWebSocket.instances[0]
    ws1.open()
    ws1.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0 } })
    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)

    useGameSocketStore.getState().attach('g2', factory)

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(0)
  })

  it('a reconnect for the same gameId preserves live', () => {
    useGameSocketStore.getState().attach('g1', factory)
    const ws1 = FakeWebSocket.instances[0]
    ws1.open()
    ws1.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0 } })
    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)

    // retryNow() connects immediately (bypassing the real reconnect-delay
    // setTimeout) - same gameId, so this exercises the "reconnect, not a
    // fresh attach" path without needing fake timers.
    useGameSocketStore.getState().retryNow()
    const ws2 = FakeWebSocket.instances[1]
    ws2.open()

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)
  })
})
