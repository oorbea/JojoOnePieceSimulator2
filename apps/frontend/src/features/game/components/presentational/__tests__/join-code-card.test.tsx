import { fireEvent, renderWithProviders, screen } from '@/test/render'

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
