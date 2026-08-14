import enGB from '@/shared/i18n/locales/en-GB.json'
import { canKick, canSwitchTeam, canTransferHost, startGate } from '@/features/game/lib/lobby-rules'
import type { GameConfig, GameSnapshot, GameViewer } from '@/features/game/types/game.types'

function baseConfig(): GameConfig {
  return {
    mangas: ['JOJO'],
    abilitySource: 'RANDOM',
    teamSize: 5,
    allowBots: false,
    visibility: 'PRIVATE',
    votingWindowSeconds: 30,
    poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
  }
}

function baseSnapshot(overrides: Partial<GameSnapshot> = {}): GameSnapshot {
  return {
    id: 'g1',
    code: 'ABCDEF',
    state: 'LOBBY',
    mode: 'GAUNTLET',
    hostId: 'p1',
    locked: false,
    config: baseConfig(),
    teams: [{ id: 't1', name: 'Squad', color: 0, memberIds: ['p1'] }],
    participants: [{ id: 'p1', displayName: 'host', teamId: 't1', kind: 'HUMAN', connected: true }],
    rounds: [],
    ...overrides,
  }
}

function resolves(key: string): boolean {
  const parts = key.split('.')
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- walking a plain JSON tree
  let node: any = enGB
  for (const part of parts) {
    if (node == null || typeof node !== 'object' || !(part in node)) return false
    node = node[part]
  }
  return typeof node === 'string'
}

describe('startGate', () => {
  const host: GameViewer = { participantId: 'p1', teamId: 't1', isHost: true, hasVoted: false }

  it('is ready for a valid Gauntlet lobby', () => {
    const gate = startGate(baseSnapshot(), host)
    expect(gate.ok).toBe(true)
    expect(resolves(gate.reasonKey)).toBe(true)
  })

  it('rejects a non-host', () => {
    const notHost: GameViewer = { ...host, isHost: false }
    const gate = startGate(baseSnapshot(), notHost)
    expect(gate.ok).toBe(false)
    expect(gate.reasonKey).toBe('game.start.reasonNotHost')
    expect(resolves(gate.reasonKey)).toBe(true)
  })

  it('rejects a lobby already past LOBBY', () => {
    const gate = startGate(baseSnapshot({ state: 'VOTING' }), host)
    expect(gate.reasonKey).toBe('game.start.reasonAlreadyStarted')
    expect(resolves(gate.reasonKey)).toBe(true)
  })

  it('rejects an empty Versus team', () => {
    const snap = baseSnapshot({
      mode: 'VERSUS',
      teams: [
        { id: 'a', name: 'A', color: 0, memberIds: ['p1'] },
        { id: 'b', name: 'B', color: 0, memberIds: [] },
      ],
    })
    const gate = startGate(snap, host)
    expect(gate.reasonKey).toBe('game.start.reasonEmptyTeam')
    expect(resolves(gate.reasonKey)).toBe(true)
  })

  it('rejects unequal non-empty Versus teams with the count as params', () => {
    const snap = baseSnapshot({
      mode: 'VERSUS',
      teams: [
        { id: 'a', name: 'A', color: 0, memberIds: ['p1', 'p2'] },
        { id: 'b', name: 'B', color: 0, memberIds: ['p3'] },
      ],
    })
    const gate = startGate(snap, host)
    expect(gate.reasonKey).toBe('game.start.reasonUnequalTeams')
    expect(gate.params).toEqual({ a: 2, b: 1 })
    expect(resolves(gate.reasonKey)).toBe(true)
  })

  it('accepts equal non-full Versus teams', () => {
    const snap = baseSnapshot({
      mode: 'VERSUS',
      teams: [
        { id: 'a', name: 'A', color: 0, memberIds: ['p1', 'p2'] },
        { id: 'b', name: 'B', color: 0, memberIds: ['p3', 'p4'] },
      ],
    })
    expect(startGate(snap, host).ok).toBe(true)
  })

  it('rejects an empty Gauntlet lobby', () => {
    const gate = startGate(baseSnapshot({ participants: [] }), host)
    expect(gate.reasonKey).toBe('game.start.reasonNoPlayers')
    expect(resolves(gate.reasonKey)).toBe(true)
  })
})

describe('canSwitchTeam', () => {
  const you: GameViewer = { participantId: 'p1', teamId: 't1', isHost: false, hasVoted: false }

  it('allows a self-move to a non-full team', () => {
    const snap = baseSnapshot({
      teams: [
        { id: 't1', name: 'A', color: 0, memberIds: ['p1'] },
        { id: 't2', name: 'B', color: 0, memberIds: [] },
      ],
    })
    expect(canSwitchTeam(snap, you, 'p1', 't2').ok).toBe(true)
  })

  it('rejects a non-host moving someone else', () => {
    const snap = baseSnapshot({
      participants: [
        { id: 'p1', displayName: 'me', teamId: 't1', kind: 'HUMAN', connected: true },
        { id: 'p2', displayName: 'other', teamId: 't1', kind: 'HUMAN', connected: true },
      ],
    })
    const gate = canSwitchTeam(snap, you, 'p2', 't1')
    expect(gate.ok).toBe(false)
    expect(resolves(gate.reasonKey)).toBe(true)
  })

  it('rejects moving onto a full team', () => {
    const snap = baseSnapshot({
      config: { ...baseConfig(), teamSize: 1 },
      teams: [
        { id: 't1', name: 'A', color: 0, memberIds: ['p1'] },
        { id: 't2', name: 'B', color: 0, memberIds: ['p2'] },
      ],
    })
    const gate = canSwitchTeam(snap, you, 'p1', 't2')
    expect(gate.reasonKey).toBe('game.lobby.reasonTeamFull')
    expect(resolves(gate.reasonKey)).toBe(true)
  })
})

describe('canKick', () => {
  const host: GameViewer = { participantId: 'p1', teamId: 't1', isHost: true, hasVoted: false }

  it('rejects kicking yourself', () => {
    const gate = canKick(baseSnapshot(), host, 'p1')
    expect(gate.reasonKey).toBe('game.kick.reasonSelf')
    expect(resolves(gate.reasonKey)).toBe(true)
  })

  it('allows a host kicking someone else', () => {
    const snap = baseSnapshot({
      participants: [
        { id: 'p1', displayName: 'host', teamId: 't1', kind: 'HUMAN', connected: true },
        { id: 'p2', displayName: 'other', teamId: 't1', kind: 'HUMAN', connected: true },
      ],
    })
    expect(canKick(snap, host, 'p2').ok).toBe(true)
  })
})

describe('canTransferHost', () => {
  const host: GameViewer = { participantId: 'p1', teamId: 't1', isHost: true, hasVoted: false }

  it('rejects transferring to a bot', () => {
    const snap = baseSnapshot({
      participants: [
        { id: 'p1', displayName: 'host', teamId: 't1', kind: 'HUMAN', connected: true },
        { id: 'bot1', displayName: 'Bot 1', teamId: 't1', kind: 'BOT', connected: true },
      ],
    })
    const gate = canTransferHost(snap, host, 'bot1')
    expect(gate.reasonKey).toBe('game.transferHost.reasonBot')
    expect(resolves(gate.reasonKey)).toBe(true)
  })
})
