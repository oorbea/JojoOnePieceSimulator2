// EXPO_PUBLIC_SOCKET_URL is blank in the real .env (native realtime isn't
// deployed yet) - stub it here so buildGameSocketUrl doesn't short-circuit
// to null and skip opening a socket at all.
import { useGameSocketStore } from '@/features/game/stores/game-socket.store'
import { useSessionStore } from '@/shared/stores/session.store'
import { mintGameSocketTicket } from '@/shared/api/stream-tickets'
import { BASE_RECONNECT_MS } from '@/features/game/lib/backoff'

jest.mock('@/shared/config/env', () => ({
  env: {
    EXPO_PUBLIC_API_URL: 'http://localhost/api/v1',
    EXPO_PUBLIC_SOCKET_URL: 'ws://localhost/api/v1',
    EXPO_PUBLIC_BUILD_ID: 'test',
  },
}))

// connect() now mints a ticket before opening a socket - mocked here so
// every existing test (written for a synchronous connect()) keeps working
// once its attach()/retryNow() call is followed by a flush() (see below).
// Individual tests override this per-case (mockResolvedValueOnce/
// mockRejectedValueOnce) to exercise the mint-failure paths.
jest.mock('@/shared/api/stream-tickets', () => ({
  mintGameSocketTicket: jest.fn(),
}))
const mockMint = jest.mocked(mintGameSocketTicket)

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

// connect() has exactly one await point (mintGameSocketTicket) before it
// either opens a socket or bails out - two microtask ticks is enough margin
// for that plus zustand's own (synchronous) set() calls in between.
async function flush() {
  await Promise.resolve()
  await Promise.resolve()
}

// Thin wrapper so every existing test reads the same as before this tanda
// (attach then immediately use the socket), with the flush the new async
// connect() requires made explicit at the call site.
async function attach(gameId: string) {
  useGameSocketStore.getState().attach(gameId, factory)
  await flush()
}

// axios-shaped rejection toAppError (shared/api/errors.ts) recognizes -
// enough for its `.response?.status` extraction, nothing else.
function httpError(status: number) {
  return { response: { status }, message: `http ${status}` }
}

// The store now validates every frame against the generated
// serverFrameSchema (game-socket.store.ts), so a STATE payload must be a
// complete, valid GameSnapshotResponse/GameViewerResponse - a real backend
// never sends a partial one. Tests build on this base via a shallow
// `game`/`you` override so each case's `.toEqual` still holds (the store
// stores the payload wholesale, so actual and expected are built from the
// same merge).
function validGame(overrides: Record<string, unknown> = {}) {
  return {
    id: 'g1',
    code: 'ABCDEF',
    state: 'LOBBY',
    mode: 'GAUNTLET',
    hostId: 'p1',
    locked: false,
    config: {
      stageMangas: ['ONE_PIECE'],
      powerMangas: ['ONE_PIECE'],
      abilitySource: 'RANDOM',
      teamSize: 4,
      allowBots: true,
      visibility: 'PRIVATE',
      votingWindowSeconds: 30,
      poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
      revealSpeed: 'NORMAL',
      summaryDurationSeconds: 10,
    },
    teams: [],
    participants: [],
    rounds: [],
    ...overrides,
  }
}

function validYou(overrides: Record<string, unknown> = {}) {
  return { participantId: 'p1', teamId: 't1', isHost: true, hasVoted: false, ...overrides }
}

function validState(gameOverrides: Record<string, unknown> = {}, youOverrides: Record<string, unknown> = {}) {
  return { game: validGame(gameOverrides), you: validYou(youOverrides) }
}

beforeEach(() => {
  FakeWebSocket.instances = []
  mockMint.mockReset()
  mockMint.mockResolvedValue('test-ticket')
  useGameSocketStore.getState().reset()
  useSessionStore.setState({
    session: {
      accessToken: 'tok',
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
  it('replaces the snapshot wholesale on STATE and never merges', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    const snapshotA = validState({ locked: false })
    ws.receive({ type: 'STATE', payload: snapshotA })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshotA)

    const snapshotB = validState({ locked: true })
    ws.receive({ type: 'STATE', payload: snapshotB })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshotB)
  })

  it('does not touch the snapshot on a delta frame', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    const snapshot = validState()
    ws.receive({ type: 'STATE', payload: snapshot })

    ws.receive({ type: 'PLAYER_JOINED', payload: { participantId: 'p2' } })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshot)
    expect(useGameSocketStore.getState().feed.length).toBe(1)
  })

  it('VOTE_CAST never touches the snapshot', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    const snapshot = validState()
    ws.receive({ type: 'STATE', payload: snapshot })

    ws.receive({ type: 'VOTE_CAST', payload: { roundIndex: 0, votesCast: 1, voters: 3 } })
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshot)
    expect(useGameSocketStore.getState().feed.length).toBe(0)
  })

  it('sets terminal ABORTED on GAME_ABORTED and stops reconnecting', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'GAME_ABORTED', payload: { reason: 'host cancelled' } })
    expect(useGameSocketStore.getState().terminal).toEqual({
      kind: 'ABORTED',
      reason: 'host cancelled',
    })

    ws.close()
    expect(useGameSocketStore.getState().status).toBe('closed')
  })

  it('sets terminal KICKED only when the kicked participant is self', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'STATE', payload: validState() })

    ws.receive({ type: 'PLAYER_KICKED', payload: { participantId: 'p2' } })
    expect(useGameSocketStore.getState().terminal).toBeNull()

    ws.receive({ type: 'PLAYER_KICKED', payload: { participantId: 'p1' } })
    expect(useGameSocketStore.getState().terminal).toEqual({ kind: 'KICKED' })
  })

  it('records lastError from an ERROR frame with its requestId', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'ERROR', requestId: 'req-1', payload: { error: 'nope', code: 'NOT_HOST' } })
    expect(useGameSocketStore.getState().lastError).toEqual({
      message: 'nope',
      code: 'NOT_HOST',
      requestId: 'req-1',
    })
  })

  it('ignores a malformed frame instead of throwing', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    expect(() => ws.onmessage?.({ data: 'not json' })).not.toThrow()
  })

  it('reuses one socket for two attach calls with the same gameId', async () => {
    await attach('g1')
    await attach('g1')
    expect(FakeWebSocket.instances.length).toBe(1)
  })

  it('does not open a socket without a session token', async () => {
    useSessionStore.setState({ session: null, isHydrated: true })
    await attach('g1')
    expect(FakeWebSocket.instances.length).toBe(0)
  })

  it('LOADOUTS_ASSIGNED bumps assignmentSeq without touching snapshot/feed', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    const snapshot = validState()
    ws.receive({ type: 'STATE', payload: snapshot })

    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)
    expect(useGameSocketStore.getState().live.assignedRoundIndex).toBe(0)
    expect(useGameSocketStore.getState().snapshot).toEqual(snapshot)
    expect(useGameSocketStore.getState().feed.length).toBe(0)
  })

  it("LOADOUTS_ASSIGNED sets revealEndsAt from the frame's own closesAt (never Date.now()+duration)", async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:20.000Z' } })

    const live = useGameSocketStore.getState().live
    // Exact equality, not toBeGreaterThanOrEqual(Date.now()) - the whole
    // point of this frame's closesAt is that it is NOT relative to when
    // the frame happened to arrive.
    expect(live.revealEndsAt).toBe(Date.parse('2100-01-01T00:00:20.000Z'))
  })

  it('STATE adopts game.revealEndsAt when no reveal is already tracked (reconnect mid-sorteo)', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({
      type: 'STATE',
      payload: validState({ revealEndsAt: '2100-01-01T00:00:20.000Z' }),
    })

    const live = useGameSocketStore.getState().live
    expect(live.revealEndsAt).toBe(Date.parse('2100-01-01T00:00:20.000Z'))
  })

  it('STATE does not override an already-tracked reveal deadline from LOADOUTS_ASSIGNED', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })
    const revealEndsAtBefore = useGameSocketStore.getState().live.revealEndsAt

    ws.receive({
      type: 'STATE',
      payload: validState({ revealEndsAt: '2100-01-01T00:00:20.000Z' }),
    })

    expect(useGameSocketStore.getState().live.revealEndsAt).toBe(revealEndsAtBefore)
  })

  it('bumps assignmentSeq to 2 across two assignment frames', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 1, closesAt: '2100-01-01T00:00:18.000Z' } })

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(2)
    expect(useGameSocketStore.getState().live.assignedRoundIndex).toBe(1)
  })

  it('VOTING_OPENED populates live with roundIndex/closesAt, clears tiebreak, and clears the reveal deadline', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })

    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })

    const live = useGameSocketStore.getState().live
    expect(live.votingRoundIndex).toBe(0)
    expect(live.votingClosesAt).toBe(Date.parse('2100-01-01T00:00:10.000Z'))
    expect(live.tiebreak).toBe(false)
    expect(live.revealEndsAt).toBeNull()
  })

  it('TIEBREAK_OPENED populates live the same way with tiebreak true', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({
      type: 'TIEBREAK_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })

    const live = useGameSocketStore.getState().live
    expect(live.votingRoundIndex).toBe(0)
    expect(live.votingClosesAt).toBe(Date.parse('2100-01-01T00:00:10.000Z'))
    expect(live.tiebreak).toBe(true)
  })

  it('ROUND_RESOLVED clears the countdown and still pushes to feed', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })

    ws.receive({
      type: 'ROUND_RESOLVED',
      payload: { roundIndex: 0, winner: 'A', decidedByCoinFlip: false, closesAt: '2100-01-01T00:00:06.000Z' },
    })

    const live = useGameSocketStore.getState().live
    expect(live.votingRoundIndex).toBeNull()
    expect(live.votingClosesAt).toBeNull()
    expect(useGameSocketStore.getState().feed.length).toBe(1)
  })

  it('markAssignmentRevealed sets revealedAssignmentSeq to the current assignmentSeq', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })
    expect(useGameSocketStore.getState().live.assignmentSeq).toBeGreaterThan(
      useGameSocketStore.getState().live.revealedAssignmentSeq
    )

    useGameSocketStore.getState().markAssignmentRevealed()

    const live = useGameSocketStore.getState().live
    expect(live.revealedAssignmentSeq).toBe(live.assignmentSeq)
  })

  it('reset() clears live back to its zero/null initial value', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })

    useGameSocketStore.getState().reset()

    expect(useGameSocketStore.getState().live).toEqual({
      assignmentSeq: 0,
      revealedAssignmentSeq: 0,
      assignedRoundIndex: null,
      revealEndsAt: null,
      revealReadyCount: null,
      revealReadyTotal: null,
      summaryEndsAt: null,
      summaryReadyCount: null,
      summaryReadyTotal: null,
      votingRoundIndex: null,
      votingClosesAt: null,
      tiebreak: false,
      votesCast: null,
      voters: null,
      resultEndsAt: null,
      resultDismissed: false,
    })
  })

  it('a gameId change clears live', async () => {
    await attach('g1')
    const ws1 = FakeWebSocket.instances[0]
    ws1.open()
    ws1.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })
    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)

    await attach('g2')

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(0)
  })

  it('a reconnect for the same gameId preserves live', async () => {
    await attach('g1')
    const ws1 = FakeWebSocket.instances[0]
    ws1.open()
    ws1.receive({ type: 'LOADOUTS_ASSIGNED', payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:18.000Z' } })
    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)

    // retryNow() connects immediately (bypassing the real reconnect-delay
    // setTimeout) - same gameId, so this exercises the "reconnect, not a
    // fresh attach" path without needing fake timers.
    useGameSocketStore.getState().retryNow()
    await flush()
    const ws2 = FakeWebSocket.instances[1]
    ws2.open()

    expect(useGameSocketStore.getState().live.assignmentSeq).toBe(1)
  })

  it('VOTE_CAST writes absolute votesCast/voters when it matches the current voting round', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })

    ws.receive({ type: 'VOTE_CAST', payload: { roundIndex: 0, votesCast: 1, voters: 3 } })
    expect(useGameSocketStore.getState().live.votesCast).toBe(1)
    expect(useGameSocketStore.getState().live.voters).toBe(3)

    // A second frame overwrites, never accumulates - bots can emit several
    // in a row right after the window opens.
    ws.receive({ type: 'VOTE_CAST', payload: { roundIndex: 0, votesCast: 2, voters: 3 } })
    expect(useGameSocketStore.getState().live.votesCast).toBe(2)
  })

  it('VOTE_CAST for a round that is not the current voting round is ignored', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 1, closesAt: '2100-01-01T00:00:10.000Z' },
    })

    ws.receive({ type: 'VOTE_CAST', payload: { roundIndex: 0, votesCast: 1, voters: 3 } })

    expect(useGameSocketStore.getState().live.votesCast).toBe(0)
  })

  it('VOTING_OPENED/TIEBREAK_OPENED reset votesCast to 0 and voters to null', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })
    ws.receive({ type: 'VOTE_CAST', payload: { roundIndex: 0, votesCast: 2, voters: 3 } })

    ws.receive({
      type: 'TIEBREAK_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })

    const live = useGameSocketStore.getState().live
    expect(live.votesCast).toBe(0)
    expect(live.voters).toBeNull()
  })

  it('ROUND_RESOLVED clears votesCast/voters to null', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })
    ws.receive({ type: 'VOTE_CAST', payload: { roundIndex: 0, votesCast: 1, voters: 2 } })

    ws.receive({
      type: 'ROUND_RESOLVED',
      payload: { roundIndex: 0, winner: 'A', decidedByCoinFlip: false, closesAt: '2100-01-01T00:00:06.000Z' },
    })

    const live = useGameSocketStore.getState().live
    expect(live.votesCast).toBeNull()
    expect(live.voters).toBeNull()
  })

  it('STATE adopts game.votingEndsAt when no voting deadline is already tracked (reconnect mid-vote)', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    ws.receive({
      type: 'STATE',
      payload: validState({ votingEndsAt: '2100-01-01T00:00:10.000Z' }),
    })

    expect(useGameSocketStore.getState().live.votingClosesAt).toBe(
      Date.parse('2100-01-01T00:00:10.000Z')
    )
  })

  it('STATE does not override an already-tracked voting deadline from VOTING_OPENED', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })
    const votingClosesAtBefore = useGameSocketStore.getState().live.votingClosesAt

    ws.receive({
      type: 'STATE',
      payload: validState({ votingEndsAt: '2100-01-01T00:00:20.000Z' }),
    })

    expect(useGameSocketStore.getState().live.votingClosesAt).toBe(votingClosesAtBefore)
  })

  it('STATE adopts game.resultEndsAt when no result display is already tracked (reconnect mid-result)', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    // No preceding ROUND_RESOLVED - this is the reconnect path, where the
    // client never saw the frame at all (RESYNC -> a fresh STATE).
    ws.receive({
      type: 'STATE',
      payload: validState({ resultEndsAt: '2100-01-01T00:00:06.000Z' }),
    })

    expect(useGameSocketStore.getState().live.resultEndsAt).toBe(
      Date.parse('2100-01-01T00:00:06.000Z')
    )
  })

  it('STATE does not overwrite a resultEndsAt already set by ROUND_RESOLVED', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'ROUND_RESOLVED',
      payload: { roundIndex: 0, winner: 'A', decidedByCoinFlip: false, closesAt: '2100-01-01T00:00:06.000Z' },
    })
    const resultEndsAtBefore = useGameSocketStore.getState().live.resultEndsAt

    ws.receive({
      type: 'STATE',
      payload: validState({ resultEndsAt: '2100-01-01T00:00:20.000Z' }),
    })

    expect(useGameSocketStore.getState().live.resultEndsAt).toBe(resultEndsAtBefore)
  })

  it("ROUND_RESOLVED sets resultEndsAt from its own closesAt and clears resultDismissed", async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 0, closesAt: '2100-01-01T00:00:10.000Z' },
    })
    useGameSocketStore.getState().dismissResult()
    expect(useGameSocketStore.getState().live.resultDismissed).toBe(true)

    ws.receive({
      type: 'ROUND_RESOLVED',
      payload: { roundIndex: 0, winner: 'A', decidedByCoinFlip: false, closesAt: '2100-01-01T00:00:06.000Z' },
    })

    const live = useGameSocketStore.getState().live
    expect(live.resultEndsAt).toBe(Date.parse('2100-01-01T00:00:06.000Z'))
    expect(live.resultDismissed).toBe(false)
  })

  it('VOTING_OPENED/TIEBREAK_OPENED reset resultEndsAt/resultDismissed too', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    ws.receive({
      type: 'STATE',
      payload: validState({ resultEndsAt: '2100-01-01T00:00:06.000Z' }),
    })
    useGameSocketStore.getState().dismissResult()

    ws.receive({
      type: 'VOTING_OPENED',
      payload: { roundIndex: 1, closesAt: '2100-01-01T00:00:10.000Z' },
    })

    const live = useGameSocketStore.getState().live
    expect(live.resultEndsAt).toBeNull()
    expect(live.resultDismissed).toBe(false)
  })

  it('dismissResult sets resultDismissed', async () => {
    await attach('g1')
    expect(useGameSocketStore.getState().live.resultDismissed).toBe(false)

    useGameSocketStore.getState().dismissResult()

    expect(useGameSocketStore.getState().live.resultDismissed).toBe(true)
  })

  // Frames are now validated against the generated serverFrameSchema
  // (game-socket.store.ts) instead of blindly cast - these two prove the
  // fail-soft policy: a frame that fails validation degrades the same way
  // an unrecognized frame type already did (a feed entry, nothing else
  // touched), rather than throwing or corrupting state with a
  // partially-wrong payload.
  it('a STATE frame missing required fields keeps the previous snapshot instead of adopting a broken one', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()
    const snapshot = validState()
    ws.receive({ type: 'STATE', payload: snapshot })

    // Missing `config`, `teams`, `participants`, `rounds`, etc. - a real
    // backend never sends this, but this is exactly the shape a field
    // rename/removal would produce.
    ws.receive({ type: 'STATE', payload: { game: { id: 'g1' }, you: { participantId: 'p1' } } })

    expect(useGameSocketStore.getState().snapshot).toEqual(snapshot)
    expect(useGameSocketStore.getState().feed.length).toBe(1)
  })

  it('a frame whose type is unrecognized still lands in the feed and does not throw', async () => {
    await attach('g1')
    const ws = FakeWebSocket.instances[0]
    ws.open()

    expect(() => ws.receive({ type: 'SOME_FUTURE_FRAME', payload: { whatever: true } })).not.toThrow()
    expect(useGameSocketStore.getState().feed.length).toBe(1)
  })

  // --- connect() mints a ticket before opening a socket (2026-09-03) ---

  it('the socket URL carries the minted ticket', async () => {
    mockMint.mockResolvedValueOnce('minted-abc')
    await attach('g1')
    expect(FakeWebSocket.instances[0].url).toContain('ticket=minted-abc')
  })

  it('a 403 minting a ticket closes the store without opening a socket or scheduling a retry', async () => {
    mockMint.mockRejectedValueOnce(httpError(403))
    useGameSocketStore.getState().attach('g1', factory)
    await flush()

    expect(useGameSocketStore.getState().status).toBe('closed')
    expect(FakeWebSocket.instances.length).toBe(0)
    expect(useGameSocketStore.getState().nextRetryAt).toBeNull()
  })

  it('a 401 minting a ticket closes the store (the interceptor already cleared the session)', async () => {
    mockMint.mockRejectedValueOnce(httpError(401))
    useGameSocketStore.getState().attach('g1', factory)
    await flush()

    expect(useGameSocketStore.getState().status).toBe('closed')
    expect(FakeWebSocket.instances.length).toBe(0)
  })

  it('a network error minting a ticket schedules a reconnect that re-mints', async () => {
    jest.useFakeTimers()
    try {
      mockMint.mockRejectedValueOnce(new Error('network down'))
      useGameSocketStore.getState().attach('g1', factory)
      await flush()

      expect(useGameSocketStore.getState().status).toBe('reconnecting')
      expect(useGameSocketStore.getState().nextRetryAt).not.toBeNull()
      expect(mockMint).toHaveBeenCalledTimes(1)

      jest.advanceTimersByTime(BASE_RECONNECT_MS)
      await flush()

      expect(mockMint).toHaveBeenCalledTimes(2)
      expect(FakeWebSocket.instances.length).toBe(1)
    } finally {
      jest.useRealTimers()
    }
  })

  it("detach() during a pending mint opens no socket (the connectGeneration guard)", async () => {
    let resolveMint: (value: string) => void = () => {}
    mockMint.mockReturnValueOnce(
      new Promise<string>((resolve) => {
        resolveMint = resolve
      })
    )

    useGameSocketStore.getState().attach('g1', factory)
    // The mint above is still in flight - detach before it resolves, the
    // same race a component unmounting mid-connect would hit.
    useGameSocketStore.getState().detach()
    resolveMint('late-ticket')
    await flush()

    expect(FakeWebSocket.instances.length).toBe(0)
  })
})
