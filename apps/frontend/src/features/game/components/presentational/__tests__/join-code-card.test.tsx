import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { JoinCodeCard } from '../join-code-card'

function baseProps(overrides: Partial<React.ComponentProps<typeof JoinCodeCard>> = {}) {
  return {
    code: 'K7F2QX',
    isPublic: false,
    onCopy: jest.fn().mockResolvedValue('copied' as const),
    onShare: jest.fn().mockResolvedValue('shared' as const),
    ...overrides,
  }
}

describe('JoinCodeCard', () => {
  it('renders both copy and share buttons with distinct tooltip labels', async () => {
    await renderWithProviders(<JoinCodeCard {...baseProps()} />)

    expect(screen.getByLabelText('Copy join code')).toBeTruthy()
    expect(screen.getByLabelText('Share invite link')).toBeTruthy()
  })

  it('copy button calls onCopy, share button calls onShare (never the other)', async () => {
    const onCopy = jest.fn().mockResolvedValue('copied' as const)
    const onShare = jest.fn().mockResolvedValue('shared' as const)
    await renderWithProviders(<JoinCodeCard {...baseProps({ onCopy, onShare })} />)

    // The copy handler is async (awaits onCopy, then setState) - flushed
    // through `act` so its state update settles here rather than leaking
    // into whichever test runs next.
    await act(async () => {
      fireEvent.press(screen.getByLabelText('Copy join code'))
    })
    expect(onCopy).toHaveBeenCalledTimes(1)
    expect(onShare).not.toHaveBeenCalled()

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Share invite link'))
    })
    expect(onShare).toHaveBeenCalledTimes(1)
    expect(onCopy).toHaveBeenCalledTimes(1)
  })

  it('renders no regenerate button when onRegenerate is omitted (non-host)', async () => {
    await renderWithProviders(<JoinCodeCard {...baseProps()} />)

    expect(screen.queryByLabelText('Get a new code')).toBeNull()
  })

  it('renders a regenerate button and calls onRegenerate when pressed (host)', async () => {
    const onRegenerate = jest.fn()
    await renderWithProviders(<JoinCodeCard {...baseProps({ onRegenerate })} />)

    const button = screen.getByLabelText('Get a new code')
    fireEvent.press(button)

    expect(onRegenerate).toHaveBeenCalledTimes(1)
  })
})
