import { matchRecap } from '@/features/game/lib/game-result'
import type {
  GameParticipant,
  GameRound,
  GameSnapshot,
  GameTeam,
  GameViewer,
} from '@/features/game/types/game.types'

function snapshot(overrides: Partial<GameSnapshot> = {}): GameSnapshot {
  return {
    id: 'g1',
    code: 'ABC123',
    state: 'FINISHED',
    mode: 'GAUNTLET',
    hostId: 'p1',
    locked: false,
    config: {
      stageMangas: ['ONE_PIECE'],
      powerMangas: ['ONE_PIECE'],
      abilitySource: 'RANDOM',
      teamSize: 4,
      allowBots: false,
      visibility: 'PRIVATE',
      votingWindowSeconds: 30,
      revealSpeed: 'NORMAL',
      summaryDurationSeconds: 60,
      poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
    },
    teams: [],
    participants: [],
    rounds: [],
    ...overrides,
  }
}

function round(overrides: Partial<GameRound> = {}): GameRound {
  return {
    index: 0,
    stage: {
      id: 's1',
      manga: 'JOJO',
      order: 0,
      name: 'Phantom Blood',
      description: '',
      picture: '',
      pictureThumb: '',
      pictureStatus: 'READY',
    },
    options: ['SURVIVE', 'FALL'],
    tiebreakUsed: false,
    votedParticipantIds: [],
    ...overrides,
  }
}

function viewer(overrides: Partial<GameViewer> = {}): GameViewer {
  return { participantId: 'p1', teamId: 't1', isHost: false, hasVoted: false, ...overrides }
}

function team(overrides: Partial<GameTeam> = {}): GameTeam {
  return { id: 't1', name: 'Squad', color: 0, memberIds: [], ...overrides }
}

function participant(overrides: Partial<GameParticipant> = {}): GameParticipant {
  return {
    id: 'p1',
    displayName: 'jotaro',
    teamId: 't1',
    kind: 'HUMAN',
    connected: true,
    avatarThumb: '',
    ...overrides,
  }
}

describe('matchRecap - GAUNTLET', () => {
  it('reports a collective survived verdict and no per-participant winners', () => {
    const snap = snapshot({
      mode: 'GAUNTLET',
      teams: [team()],
      participants: [participant(), participant({ id: 'p2', displayName: 'bot', kind: 'BOT' })],
      rounds: [round({ result: { winner: 'SURVIVE', decidedByCoinFlip: false } })],
      result: { mode: 'GAUNTLET', winner: 'SURVIVE', roundsPlayed: 1, aborted: false },
    })

    const recap = matchRecap(snap, viewer())

    expect(recap.squadSurvived).toBe(true)
    expect(recap.winnerOptionId).toBe('SURVIVE')
    // A gauntlet run has no team winner and no individual outcome - the
    // whole squad shares one verdict.
    expect(recap.winnerTeamName).toBeNull()
    expect(recap.outcomes.map((o) => o.won)).toEqual([null, null])
    expect(recap.roundsPlayed).toBe(1)
  })

  it('reports a fallen squad when the run ended on FALL', () => {
    const snap = snapshot({
      mode: 'GAUNTLET',
      rounds: [round({ result: { winner: 'FALL', decidedByCoinFlip: false } })],
      result: { mode: 'GAUNTLET', winner: 'FALL', roundsPlayed: 1, aborted: false },
    })

    expect(matchRecap(snap, viewer()).squadSurvived).toBe(false)
  })

  it('recaps every round with its stage, winner and vote breakdown', () => {
    const snap = snapshot({
      mode: 'GAUNTLET',
      participants: [participant(), participant({ id: 'p2', displayName: 'josuke' })],
      rounds: [
        round({
          index: 0,
          votes: { p1: 'SURVIVE', p2: 'SURVIVE' },
          result: { winner: 'SURVIVE', decidedByCoinFlip: false },
        }),
        round({
          index: 1,
          stage: { ...round().stage, id: 's2', name: 'Battle Tendency' },
          votes: { p1: 'FALL', p2: 'SURVIVE' },
          tiebreakUsed: true,
          result: { winner: 'FALL', decidedByCoinFlip: true },
        }),
      ],
      result: { mode: 'GAUNTLET', winner: 'FALL', roundsPlayed: 2, aborted: false },
    })

    const recap = matchRecap(snap, viewer())

    expect(recap.rounds).toHaveLength(2)
    expect(recap.rounds[0].stageName).toBe('Phantom Blood')
    expect(recap.rounds[0].winnerOptionId).toBe('SURVIVE')
    expect(recap.rounds[0].decidedByCoinFlip).toBe(false)
    expect(recap.rounds[0].unresolved).toBe(false)

    expect(recap.rounds[1].stageName).toBe('Battle Tendency')
    expect(recap.rounds[1].decidedByCoinFlip).toBe(true)

    // Counts come from voteTally, in the fixed option order the vote bar used.
    const first = recap.rounds[0].entries
    expect(first.map((e) => e.id)).toEqual(['SURVIVE', 'FALL'])
    expect(first.map((e) => e.count)).toEqual([2, 0])
    expect(first[0].voterIds).toEqual(['p1', 'p2'])
    expect(recap.rounds[0].maxCount).toBe(2)
  })
})

describe('matchRecap - VERSUS', () => {
  const teamA = team({ id: 'team-a', name: 'Straw Hats' })
  const teamB = team({ id: 'team-b', name: 'Baroque Works' })

  function versusSnapshot(overrides: Partial<GameSnapshot> = {}) {
    return snapshot({
      mode: 'VERSUS',
      teams: [teamA, teamB],
      participants: [
        participant({ id: 'p1', teamId: 'team-a' }),
        participant({ id: 'p2', displayName: 'crocodile', teamId: 'team-b' }),
      ],
      rounds: [
        round({
          options: ['team-a', 'team-b'],
          votes: { p1: 'team-a', p2: 'team-a' },
          result: { winner: 'team-a', decidedByCoinFlip: false },
        }),
      ],
      result: { mode: 'VERSUS', winner: 'team-a', roundsPlayed: 1, aborted: false },
      ...overrides,
    })
  }

  it('resolves the winning team name and marks per-participant outcomes', () => {
    const recap = matchRecap(versusSnapshot(), viewer({ teamId: 'team-a' }))

    expect(recap.winnerTeamName).toBe('Straw Hats')
    expect(recap.squadSurvived).toBeNull()
    expect(recap.outcomes).toEqual([
      {
        participantId: 'p1',
        displayName: 'jotaro',
        teamId: 'team-a',
        bot: false,
        isSelf: true,
        won: true,
      },
      {
        participantId: 'p2',
        displayName: 'crocodile',
        teamId: 'team-b',
        bot: false,
        isSelf: false,
        won: false,
      },
    ])
  })

  it("prefers the server's own end-of-game seat list over the live roster", () => {
    const snap = versusSnapshot({
      result: {
        mode: 'VERSUS',
        winner: 'team-a',
        roundsPlayed: 1,
        aborted: false,
        participants: [
          { participantId: 'px', displayName: 'left-early', teamId: 'team-b', bot: false },
          { participantId: 'pb', displayName: 'Bot 1', teamId: 'team-a', bot: true },
        ],
      },
    })

    const recap = matchRecap(snap, viewer({ teamId: 'team-a' }))

    expect(recap.outcomes.map((o) => o.participantId)).toEqual(['px', 'pb'])
    expect(recap.outcomes[0].won).toBe(false)
    expect(recap.outcomes[1]).toMatchObject({ bot: true, won: true })
  })

  it('falls back to the live roster when the result carries no seat list', () => {
    const recap = matchRecap(versusSnapshot(), viewer({ teamId: 'team-a' }))
    expect(recap.outcomes.map((o) => o.participantId)).toEqual(['p1', 'p2'])
  })
})

describe('matchRecap - aborted and edge cases', () => {
  it('marks an aborted game, keeps the rounds actually played, and has no winner', () => {
    const snap = snapshot({
      state: 'ABORTED',
      participants: [participant()],
      rounds: [round({ index: 0, votes: { p1: 'SURVIVE' } })],
      result: { mode: 'GAUNTLET', winner: '', roundsPlayed: 1, aborted: true },
    })

    const recap = matchRecap(snap, viewer())

    expect(recap.aborted).toBe(true)
    expect(recap.winnerOptionId).toBeNull()
    expect(recap.winnerTeamName).toBeNull()
    expect(recap.squadSurvived).toBeNull()
    expect(recap.rounds).toHaveLength(1)
    // The round never resolved - its votes are shown as they stood.
    expect(recap.rounds[0].unresolved).toBe(true)
    expect(recap.rounds[0].winnerOptionId).toBeNull()
    expect(recap.rounds[0].entries.find((e) => e.id === 'SURVIVE')?.count).toBe(1)
  })

  it('infers aborted from the state when no result object came through', () => {
    const snap = snapshot({ state: 'ABORTED', result: undefined })
    expect(matchRecap(snap, viewer()).aborted).toBe(true)
  })

  it('handles a game aborted before any round was played', () => {
    const snap = snapshot({
      state: 'ABORTED',
      participants: [participant()],
      rounds: [],
      result: { mode: 'GAUNTLET', winner: '', roundsPlayed: 0, aborted: true },
    })

    const recap = matchRecap(snap, viewer())

    expect(recap.rounds).toEqual([])
    expect(recap.roundsPlayed).toBe(0)
    expect(recap.outcomes).toHaveLength(1)
  })

  it('falls back to the round count when the result omits roundsPlayed', () => {
    const snap = snapshot({ rounds: [round(), round({ index: 1 })], result: undefined })
    expect(matchRecap(snap, viewer()).roundsPlayed).toBe(2)
  })

  it('uses the tied breakdown for a round cut short mid-tiebreak', () => {
    const snap = snapshot({
      state: 'ABORTED',
      participants: [participant(), participant({ id: 'p2' })],
      rounds: [round({ tiedVotes: { p1: 'SURVIVE', p2: 'FALL' }, tiebreakUsed: true })],
      result: { mode: 'GAUNTLET', winner: '', roundsPlayed: 1, aborted: true },
    })

    const recap = matchRecap(snap, viewer())

    expect(recap.rounds[0].entries.map((e) => e.count)).toEqual([1, 1])
    expect(recap.rounds[0].unresolved).toBe(true)
  })
})
