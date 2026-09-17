import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { ShareInviteSheet } from '../share-invite-sheet'

function baseProps(overrides: Partial<React.ComponentProps<typeof ShareInviteSheet>> = {}) {
  return {
    visible: true,
    onClose: jest.fn(),
    url: 'https://example.com/join/abc123',
    message: 'Join my lobby',
    onCopy: jest.fn().mockResolvedValue(true),
    ...overrides,
  }
}

describe('ShareInviteSheet', () => {
  it('renders a labelled row for each share destination', async () => {
    await renderWithProviders(<ShareInviteSheet {...baseProps()} />)

    expect(screen.getByLabelText('Copy link')).toBeTruthy()
    expect(screen.getByLabelText('Show QR code')).toBeTruthy()
    expect(screen.getByLabelText('WhatsApp')).toBeTruthy()
    expect(screen.getByLabelText('Telegram')).toBeTruthy()
    expect(screen.getByLabelText('X')).toBeTruthy()
    expect(screen.getByLabelText('Email')).toBeTruthy()
  })

  it('calls onCopy when the copy-link row is pressed', async () => {
    const onCopy = jest.fn().mockResolvedValue(true)
    await renderWithProviders(<ShareInviteSheet {...baseProps({ onCopy })} />)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('Copy link'))
    })

    expect(onCopy).toHaveBeenCalledTimes(1)
  })

  it('reveals the QR code when its row is pressed', async () => {
    await renderWithProviders(<ShareInviteSheet {...baseProps()} />)

    expect(screen.queryByText('Scan with another device to open this link.')).toBeNull()
    await act(async () => {
      fireEvent.press(screen.getByLabelText('Show QR code'))
    })
    expect(screen.getByText('Scan with another device to open this link.')).toBeTruthy()
  })
})
