import { useGameInviteStore } from '@/features/game/stores/game-invite.store'

function futureIso(msFromNow: number): string {
  return new Date(Date.now() + msFromNow).toISOString()
}

describe('useGameInviteStore', () => {
  beforeEach(() => {
    useGameInviteStore.setState({ cached: null })
  })

  it('returns null when nothing is cached', () => {
    expect(useGameInviteStore.getState().get('g1', 'CODE01')).toBeNull()
  })

  it('reuses a cached token with more than 2 minutes left', () => {
    useGameInviteStore.getState().set('g1', 'CODE01', 'tok-abc', futureIso(5 * 60 * 1000))
    expect(useGameInviteStore.getState().get('g1', 'CODE01')).toBe('tok-abc')
  })

  it('does not reuse a cached token with 2 minutes or less left', () => {
    useGameInviteStore.getState().set('g1', 'CODE01', 'tok-abc', futureIso(90 * 1000))
    expect(useGameInviteStore.getState().get('g1', 'CODE01')).toBeNull()
  })

  it('drops the cache when the game id differs', () => {
    useGameInviteStore.getState().set('g1', 'CODE01', 'tok-abc', futureIso(5 * 60 * 1000))
    expect(useGameInviteStore.getState().get('g2', 'CODE01')).toBeNull()
  })

  it('drops the cache when the code differs (a rotation happened)', () => {
    useGameInviteStore.getState().set('g1', 'CODE01', 'tok-abc', futureIso(5 * 60 * 1000))
    expect(useGameInviteStore.getState().get('g1', 'CODE02')).toBeNull()
  })
})
