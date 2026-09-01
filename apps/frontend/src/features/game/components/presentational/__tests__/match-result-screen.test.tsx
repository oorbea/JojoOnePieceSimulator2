import type {
  GameParticipant,
  GameRound,
  GameSnapshot,
  GameTeam,
  GameViewer,
} from '@/features/game/types/game.types'
import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { MatchResultScreen } from '../match/match-result-screen'

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

function snapshot(overrides: Partial<GameSnapshot> = {}): GameSnapshot {
  return {
    id: 'g1',
    code: 'ABC123',
    state: 'FINISHED',
    mode: 'GAUNTLET',
    hostId: 'p1',
    locked: false,
    config: {
      stageMangas: ['JOJO'],
      powerMangas: ['JOJO'],
      abilitySource: 'RANDOM',
      teamSize: 4,
      allowBots: false,
      visibility: 'PRIVATE',
      votingWindowSeconds: 30,
      revealSpeed: 'NORMAL',
      summaryDurationSeconds: 60,
      poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
    },
    teams: [team()],
    participants: [participant()],
    rounds: [
      round({ votes: { p1: 'SURVIVE' }, result: { winner: 'SURVIVE', decidedByCoinFlip: false } }),
    ],
    result: { mode: 'GAUNTLET', winner: 'SURVIVE', roundsPlayed: 1, aborted: false },
    ...overrides,
  }
}

function viewer(overrides: Partial<GameViewer> = {}): GameViewer {
  return { participantId: 'p1', teamId: 't1', isHost: false, hasVoted: false, ...overrides }
}

function baseProps(overrides: Partial<React.ComponentProps<typeof MatchResultScreen>> = {}) {
  return {
    snapshot: snapshot(),
    you: viewer(),
    isHost: false,
    onBackToLobbies: jest.fn(),
    onRematch: jest.fn(),
    ...overrides,
  }
}

describe('MatchResultScreen', () => {
  describe('GAUNTLET', () => {
    it('renders the collective survived verdict and the round recap', async () => {
      await renderWithProviders(<MatchResultScreen {...baseProps()} />)

      expect(screen.getByText('The squad survived')).toBeTruthy()
      expect(screen.getByText('Round by round')).toBeTruthy()
      expect(screen.getByText('Round 1 - Phantom Blood')).toBeTruthy()
    })

    it('renders the fallen verdict when the run ended on FALL', async () => {
      const snap = snapshot({
        rounds: [round({ result: { winner: 'FALL', decidedByCoinFlip: false } })],
        result: { mode: 'GAUNTLET', winner: 'FALL', roundsPlayed: 1, aborted: false },
      })
      await renderWithProviders(<MatchResultScreen {...baseProps({ snapshot: snap })} />)

      expect(screen.getByText('The squad fell')).toBeTruthy()
    })

    it('shows a coin-flip badge on a round decided by one', async () => {
      const snap = snapshot({
        rounds: [round({ result: { winner: 'SURVIVE', decidedByCoinFlip: true } })],
      })
      await renderWithProviders(<MatchResultScreen {...baseProps({ snapshot: snap })} />)

      expect(screen.getByText('Decided by a coin flip')).toBeTruthy()
    })
  })

  describe('VERSUS', () => {
    const versus = snapshot({
      mode: 'VERSUS',
      teams: [team({ id: 'team-a', name: 'Straw Hats' }), team({ id: 'team-b', name: 'Baroque Works' })],
      participants: [
        participant({ id: 'p1', teamId: 'team-a' }),
        participant({ id: 'p2', displayName: 'crocodile', teamId: 'team-b' }),
      ],
      rounds: [
        round({
          options: ['team-a', 'team-b'],
          votes: { p1: 'team-a' },
          result: { winner: 'team-a', decidedByCoinFlip: false },
        }),
      ],
      result: { mode: 'VERSUS', winner: 'team-a', roundsPlayed: 1, aborted: false },
    })

    it('names the winning team and tells you whether you won', async () => {
      await renderWithProviders(
        <MatchResultScreen
          {...baseProps({ snapshot: versus, you: viewer({ teamId: 'team-a' }) })}
        />
      )

      expect(screen.getByText('Straw Hats wins')).toBeTruthy()
      expect(screen.getByText('Your team won.')).toBeTruthy()
    })

    it('tells a losing player they lost', async () => {
      await renderWithProviders(
        <MatchResultScreen
          {...baseProps({
            snapshot: versus,
            you: viewer({ participantId: 'p2', teamId: 'team-b' }),
          })}
        />
      )

      expect(screen.getByText('Your team lost.')).toBeTruthy()
    })
  })

  describe('aborted games', () => {
    it('shows the aborted notice instead of a winner, and still recaps the rounds played', async () => {
      const snap = snapshot({
        state: 'ABORTED',
        rounds: [round({ votes: { p1: 'SURVIVE' } })],
        result: { mode: 'GAUNTLET', winner: '', roundsPlayed: 1, aborted: true },
      })
      await renderWithProviders(<MatchResultScreen {...baseProps({ snapshot: snap })} />)

      expect(screen.getByText('Game cancelled')).toBeTruthy()
      expect(screen.getByText('The host cancelled this game before it finished.')).toBeTruthy()
      expect(screen.queryByText('The squad survived')).toBeNull()
      expect(screen.getByText('Round 1 - Phantom Blood')).toBeTruthy()
    })

    it('omits the recap entirely when no round was ever played', async () => {
      const snap = snapshot({
        state: 'ABORTED',
        rounds: [],
        result: { mode: 'GAUNTLET', winner: '', roundsPlayed: 0, aborted: true },
      })
      await renderWithProviders(<MatchResultScreen {...baseProps({ snapshot: snap })} />)

      expect(screen.queryByText('Round by round')).toBeNull()
    })
  })

  describe('actions', () => {
    it('offers only "back" to a non-host', async () => {
      await renderWithProviders(<MatchResultScreen {...baseProps({ isHost: false })} />)

      expect(screen.getByLabelText('Leave this game and go back to the lobby list')).toBeTruthy()
      expect(screen.queryByLabelText('Start a new lobby with the same players')).toBeNull()
    })

    it('offers rematch to the host', async () => {
      await renderWithProviders(<MatchResultScreen {...baseProps({ isHost: true })} />)

      expect(screen.getByLabelText('Start a new lobby with the same players')).toBeTruthy()
    })

    it('fires its callbacks on press', async () => {
      const onBackToLobbies = jest.fn()
      const onRematch = jest.fn()
      await renderWithProviders(
        <MatchResultScreen {...baseProps({ isHost: true, onBackToLobbies, onRematch })} />
      )

      // Each press gets its own awaited act() - fireEvent.press is async in
      // RNTL 14, and two bare calls back to back leave overlapping act
      // scopes (the button's pressStyle spring resolves inside them
      // asynchronously) that corrupt React's act-nesting for whatever test
      // renders next in this file. Same guard the *-form-modal tests use.
      await act(async () => {
        fireEvent.press(screen.getByLabelText('Leave this game and go back to the lobby list'))
      })
      expect(onBackToLobbies).toHaveBeenCalledTimes(1)

      await act(async () => {
        fireEvent.press(screen.getByLabelText('Start a new lobby with the same players'))
      })
      expect(onRematch).toHaveBeenCalledTimes(1)
    })

    it('surfaces a rematch error inline', async () => {
      await renderWithProviders(
        <MatchResultScreen
          {...baseProps({ isHost: true, rematchError: 'Could not start a rematch. Try again.' })}
        />
      )

      expect(screen.getByText('Could not start a rematch. Try again.')).toBeTruthy()
    })

    it('labels every interactive control for screen readers', async () => {
      await renderWithProviders(<MatchResultScreen {...baseProps({ isHost: true })} />)

      // Both actions carry an accessibilityLabel (project norm - every
      // button also gets a tooltip built from the same string).
      for (const label of [
        'Leave this game and go back to the lobby list',
        'Start a new lobby with the same players',
      ]) {
        expect(screen.getByLabelText(label)).toBeTruthy()
      }
    })
  })
})
