import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { JoinInviteScreen } from '../join-invite-screen'

function baseProps(overrides: Partial<React.ComponentProps<typeof JoinInviteScreen>> = {}) {
  return {
    status: 'checking' as const,
    onJoin: jest.fn(),
    onBack: jest.fn(),
    ...overrides,
  }
}

describe('JoinInviteScreen', () => {
  it('shows a checking indicator while status is "checking"', async () => {
    await renderWithProviders(<JoinInviteScreen {...baseProps({ status: 'checking' })} />)

    expect(screen.getByText('Checking your invite…')).toBeTruthy()
  })

  it.each([
    ['game.invite.error.expired', 'This invite link has expired.'],
    ['game.invite.error.revoked', "This invite link no longer works - the host generated a new join code."],
    ['game.invite.error.started', 'This game has already started.'],
    ['game.invite.error.full', 'This lobby is full.'],
    ['game.invite.error.locked', 'This lobby is locked.'],
    ['game.invite.error.gone', "This lobby couldn't be found."],
    ['game.invite.error.generic', "This invite link couldn't be used."],
  ])('renders the message for %s and a back button', async (errorKey, expectedMessage) => {
    await renderWithProviders(<JoinInviteScreen {...baseProps({ status: 'error', errorKey })} />)

    expect(screen.getByText(expectedMessage)).toBeTruthy()
    expect(screen.getByLabelText('Back to games')).toBeTruthy()
  })

  it('back button calls onBack from the error screen', async () => {
    const onBack = jest.fn()
    await renderWithProviders(
      <JoinInviteScreen {...baseProps({ status: 'error', errorKey: 'game.invite.error.expired', onBack })} />
    )

    fireEvent.press(screen.getByLabelText('Back to games'))
    expect(onBack).toHaveBeenCalledTimes(1)
  })

  it('renders the preview and a working join button once a preview is available', async () => {
    const onJoin = jest.fn()
    await renderWithProviders(
      <JoinInviteScreen
        {...baseProps({
          status: 'preview',
          onJoin,
          preview: {
            gameId: 'g1',
            mode: 'GAUNTLET',
            hostDisplayName: 'host',
            abilitySource: 'RANDOM',
            mangas: ['JOJO'],
            playerCount: 1,
            maxPlayers: 5,
            votingWindowSeconds: 30,
            allowBots: false,
            locked: false,
            visibility: 'PRIVATE',
          },
        })}
      />
    )

    const joinButton = screen.getByLabelText('Join lobby')
    expect(joinButton).toBeTruthy()
    fireEvent.press(joinButton)
    expect(onJoin).toHaveBeenCalledTimes(1)
  })

  it('disables the join button while status is "joining"', async () => {
    await renderWithProviders(
      <JoinInviteScreen
        {...baseProps({
          status: 'joining',
          preview: {
            gameId: 'g1',
            mode: 'GAUNTLET',
            hostDisplayName: 'host',
            abilitySource: 'RANDOM',
            mangas: ['JOJO'],
            playerCount: 1,
            maxPlayers: 5,
            votingWindowSeconds: 30,
            allowBots: false,
            locked: false,
            visibility: 'PRIVATE',
          },
        })}
      />
    )

    expect(screen.getByText('Joining…')).toBeTruthy()
  })
})
