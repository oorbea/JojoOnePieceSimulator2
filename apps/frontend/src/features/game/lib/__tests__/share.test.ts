import { canSystemShare, copyToClipboard, shareInviteLink } from '@/features/game/lib/share'

describe('canSystemShare (web)', () => {
  afterEach(() => {
    // @ts-expect-error - test-only cleanup of a property we defined below
    delete navigator.share
  })

  it('is true when navigator.share exists', () => {
    navigator.share = jest.fn()
    expect(canSystemShare()).toBe(true)
  })

  it('is false when navigator.share is missing', () => {
    expect(canSystemShare()).toBe(false)
  })
})

describe('shareInviteLink (web)', () => {
  afterEach(() => {
    // @ts-expect-error - test-only cleanup
    delete navigator.share
  })

  it('resolves "shared" on success', async () => {
    navigator.share = jest.fn().mockResolvedValue(undefined)
    const result = await shareInviteLink('https://example.com/join/abc', 'Join me')
    expect(result).toBe('shared')
  })

  it('maps AbortError (share sheet dismissed) to "cancelled", not "failed"', async () => {
    const abortError = new Error('The user aborted a request.')
    abortError.name = 'AbortError'
    navigator.share = jest.fn().mockRejectedValue(abortError)

    const result = await shareInviteLink('https://example.com/join/abc', 'Join me')
    expect(result).toBe('cancelled')
  })

  it('maps any other rejection to "failed"', async () => {
    navigator.share = jest.fn().mockRejectedValue(new Error('boom'))
    const result = await shareInviteLink('https://example.com/join/abc', 'Join me')
    expect(result).toBe('failed')
  })

  it('resolves "unsupported" when navigator.share is unavailable', async () => {
    const result = await shareInviteLink('https://example.com/join/abc', 'Join me')
    expect(result).toBe('unsupported')
  })
})

describe('copyToClipboard (web)', () => {
  afterEach(() => {
    // @ts-expect-error - test-only cleanup
    delete navigator.clipboard
  })

  it('resolves true on success', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: jest.fn().mockResolvedValue(undefined) },
      configurable: true,
    })
    expect(await copyToClipboard('ABC123')).toBe(true)
  })

  it('resolves false when the clipboard write rejects', async () => {
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: jest.fn().mockRejectedValue(new Error('denied')) },
      configurable: true,
    })
    expect(await copyToClipboard('ABC123')).toBe(false)
  })
})
