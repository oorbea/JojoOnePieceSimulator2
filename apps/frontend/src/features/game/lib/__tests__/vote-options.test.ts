import { voteOptions } from '@/features/game/lib/vote-options'
import type { GameRound, GameSnapshot, GameTeam, GameViewer } from '@/features/game/types/game.types'

function snapshot(overrides: Partial<GameSnapshot> = {}): GameSnapshot {
  return {
    id: 'g1',
    code: 'ABC123',
    state: 'VOTING',
    mode: 'GAUNTLET',
    hostId: 'p1',
    locked: false,
    config: {
      mangas: ['ONE_PIECE'],
      abilitySource: 'RANDOM',
      teamSize: 4,
      allowBots: false,
      visibility: 'PRIVATE',
      votingWindowSeconds: 30,
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

describe('voteOptions', () => {
  it('returns nothing when there is no current round', () => {
    expect(voteOptions(snapshot({ rounds: [] }), viewer())).toEqual([])
  })

  it('Gauntlet: maps SURVIVE/FALL onto their i18n keys and tones, never a team lookup', () => {
    const snap = snapshot({ mode: 'GAUNTLET', rounds: [round()] })
    const options = voteOptions(snap, viewer())
    expect(options).toEqual([
      { id: 'SURVIVE', labelKey: 'game.vote.gauntlet.SURVIVE', tone: 'green', isOwnTeam: false },
      { id: 'FALL', labelKey: 'game.vote.gauntlet.FALL', tone: 'red', isOwnTeam: false },
    ])
  })

  it('Versus: maps raw team-id options onto the team name (never translated) and tone', () => {
    const teamA = team({ id: 'team-a', name: 'Straw Hats' })
    const teamB = team({ id: 'team-b', name: 'Baroque Works' })
    const snap = snapshot({
      mode: 'VERSUS',
      teams: [teamA, teamB],
      rounds: [round({ options: ['team-a', 'team-b'] })],
    })
    const options = voteOptions(snap, viewer({ teamId: 'team-a' }))
    expect(options).toEqual([
      { id: 'team-a', label: 'Straw Hats', tone: 'blue', isOwnTeam: true },
      { id: 'team-b', label: 'Baroque Works', tone: 'red', isOwnTeam: false },
    ])
  })

  it('Versus: falls back to the raw id if a round option matches no team', () => {
    const teamA = team({ id: 'team-a', name: 'Straw Hats' })
    const snap = snapshot({
      mode: 'VERSUS',
      teams: [teamA],
      rounds: [round({ options: ['team-a', 'missing-team'] })],
    })
    const options = voteOptions(snap, viewer({ teamId: 'team-a' }))
    expect(options[1]).toEqual({ id: 'missing-team', label: 'missing-team', tone: 'glass', isOwnTeam: false })
  })
})
