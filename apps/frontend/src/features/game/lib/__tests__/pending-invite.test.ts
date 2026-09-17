import { clearPendingInvite, readPendingInvite, setPendingInvite } from '@/features/game/lib/pending-invite'

describe('pending-invite (web/sessionStorage)', () => {
  beforeEach(() => {
    window.sessionStorage.clear()
    jest.restoreAllMocks()
  })

  it('round-trips a stashed token', () => {
    setPendingInvite('abc123')
    expect(readPendingInvite()).toBe('abc123')
  })

  it('returns null when nothing was stashed', () => {
    expect(readPendingInvite()).toBeNull()
  })

  it('clears the stashed token', () => {
    setPendingInvite('abc123')
    clearPendingInvite()
    expect(readPendingInvite()).toBeNull()
  })

  it('drops a stash older than the 15-minute TTL', () => {
    jest.spyOn(Date, 'now').mockReturnValue(1_000_000)
    setPendingInvite('abc123')

    jest.spyOn(Date, 'now').mockReturnValue(1_000_000 + 16 * 60 * 1000)
    expect(readPendingInvite()).toBeNull()
  })

  it('keeps a stash still within the TTL', () => {
    jest.spyOn(Date, 'now').mockReturnValue(1_000_000)
    setPendingInvite('abc123')

    jest.spyOn(Date, 'now').mockReturnValue(1_000_000 + 10 * 60 * 1000)
    expect(readPendingInvite()).toBe('abc123')
  })

  it('survives malformed JSON in storage by returning null, not throwing', () => {
    window.sessionStorage.setItem('jojo.pendingInvite', 'not json')
    expect(() => readPendingInvite()).not.toThrow()
    expect(readPendingInvite()).toBeNull()
  })
})
