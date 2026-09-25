import { fireEvent, renderWithProviders, screen } from '@/test/render'

import { JoinLobbyScreen } from '../join-lobby-screen'
import type { LobbyPreview } from '@/features/game/types/game.types'

const PREVIEW: LobbyPreview = {
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
}

function baseProps(overrides: Partial<React.ComponentProps<typeof JoinLobbyScreen>> = {}) {
  return {
    onBack: jest.fn(),
    code: '',
    onChangeCode: jest.fn(),
    previewLoading: false,
    joining: false,
    onSubmit: jest.fn(),
    ...overrides,
  }
}

describe('JoinLobbyScreen', () => {
  it('submits on Enter in the code field once a preview is ready', async () => {
    const onSubmit = jest.fn()
    await renderWithProviders(
      <JoinLobbyScreen {...baseProps({ code: 'ABCDEF', preview: PREVIEW, onSubmit })} />
    )

    fireEvent(screen.getByLabelText('Join code'), 'submitEditing')
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })

  it('does not submit on Enter before the code is complete / preview loads', async () => {
    const onSubmit = jest.fn()
    await renderWithProviders(<JoinLobbyScreen {...baseProps({ code: 'AB', onSubmit })} />)

    fireEvent(screen.getByLabelText('Join code'), 'submitEditing')
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('does not submit on Enter while already joining', async () => {
    const onSubmit = jest.fn()
    await renderWithProviders(
      <JoinLobbyScreen {...baseProps({ code: 'ABCDEF', preview: PREVIEW, joining: true, onSubmit })} />
    )

    fireEvent(screen.getByLabelText('Join code'), 'submitEditing')
    expect(onSubmit).not.toHaveBeenCalled()
  })
})
