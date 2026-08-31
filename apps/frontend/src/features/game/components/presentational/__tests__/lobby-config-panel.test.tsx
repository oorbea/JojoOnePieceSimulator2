import { fireEvent, renderWithProviders, screen } from '@/test/render'
import type { PoolFilter } from '@/features/game/types/game.types'

import { LobbyConfigPanel } from '../lobby-config-panel'

function baseProps(overrides: Partial<React.ComponentProps<typeof LobbyConfigPanel>> = {}) {
  const poolFilter: PoolFilter = { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] }

  return {
    isHost: true,
    mode: 'GAUNTLET' as const,
    onChangeMode: jest.fn(),
    stageMangas: ['JOJO'] as const,
    powerMangas: ['JOJO'] as const,
    teamSize: 5,
    teamSizeMin: 1,
    teamSizeMax: 10,
    onChangeTeamSize: jest.fn(),
    allowBots: false,
    onToggleAllowBots: jest.fn(),
    visibility: 'PRIVATE' as const,
    onToggleVisibility: jest.fn(),
    votingWindowSeconds: 30,
    onChangeVotingWindow: jest.fn(),
    summaryDurationSeconds: 60,
    onChangeSummaryDuration: jest.fn(),
    revealSpeed: 'NORMAL' as const,
    onCycleRevealSpeed: jest.fn(),
    poolFilter,
    poolActiveCount: 0,
    banlistItems: [],
    onAddBan: jest.fn(),
    onRemoveBan: jest.fn(),
    onBanMatching: jest.fn(),
    onClearPoolFilter: jest.fn(),
    saving: false,
    saved: false,
    error: undefined,
    onSubmit: jest.fn(),
    ...overrides,
  }
}

describe('LobbyConfigPanel', () => {
  describe('host vs non-host', () => {
    it('renders the Save button and interactive controls for a host', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ isHost: true })} />)

      expect(screen.getByLabelText('Save')).toBeTruthy()
    })

    it('renders no Save button and read-only copy for a non-host', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ isHost: false })} />)

      expect(screen.queryByLabelText('Save')).toBeNull()
      expect(screen.getByText('Read only')).toBeTruthy()
      expect(screen.getByText('Only the host can change these settings.')).toBeTruthy()
    })
  })

  describe('Save button guards', () => {
    // Disabled Tamagui Buttons render `aria-disabled` (not RN's
    // accessibilityState), and RNTL's label queries exclude elements it
    // considers inert - so query by the still-visible "Save" text instead
    // and check its disabled ancestor directly (same pattern as
    // confirm-sheet-confirming.test.tsx).
    it('disables Save while saving', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ saving: true })} />)

      expect(screen.getByText('Saving…').parent?.props['aria-disabled']).toBe(true)
    })

    it('disables Save when stageMangas is empty', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ stageMangas: [] })} />)

      expect(screen.getByText('Save').parent?.props['aria-disabled']).toBe(true)
    })

    it('disables Save when powerMangas is empty', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ powerMangas: [] })} />)

      expect(screen.getByText('Save').parent?.props['aria-disabled']).toBe(true)
    })

    it('enables Save when not saving and both manga axes are non-empty', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps()} />)

      const button = screen.getByLabelText('Save')
      expect(button.props['aria-disabled']).not.toBe(true)
    })

    it('fires onSubmit when the enabled Save button is pressed', async () => {
      const onSubmit = jest.fn()
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ onSubmit })} />)

      fireEvent.press(screen.getByLabelText('Save'))
      expect(onSubmit).toHaveBeenCalledTimes(1)
    })

    it('shows "Saving…" while saving', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ saving: true, saved: false })} />)
      expect(screen.getByText('Saving…')).toBeTruthy()
    })

    it('shows "Saved" once saved', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ saving: false, saved: true })} />)
      expect(screen.getByText('Saved')).toBeTruthy()
    })

    it('shows "Save" otherwise', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ saving: false, saved: false })} />)
      expect(screen.getByText('Save')).toBeTruthy()
    })
  })

  describe('error display', () => {
    it('renders a non-empty error string somewhere in the tree', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ error: 'Something went wrong' })} />)

      expect(screen.getByText('Something went wrong')).toBeTruthy()
    })

    it('renders nothing extra when error is undefined', async () => {
      await renderWithProviders(<LobbyConfigPanel {...baseProps({ error: undefined })} />)

      expect(screen.queryByText('Something went wrong')).toBeNull()
    })
  })
})
