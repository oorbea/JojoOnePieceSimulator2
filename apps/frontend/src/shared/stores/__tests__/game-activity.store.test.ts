import { activityForState, useGameActivityStore } from '../game-activity.store'

describe('activityForState', () => {
  it('maps the lobby to "lobby"', () => {
    expect(activityForState('LOBBY')).toBe('lobby')
  })

  it.each(['ASSIGNING', 'SUMMARY', 'VOTING', 'TIEBREAK', 'RESOLVING'] as const)(
    'maps %s to "playing"',
    (state) => {
      expect(activityForState(state)).toBe('playing')
    }
  )

  it.each(['FINISHED', 'ABORTED'] as const)('maps %s to "none"', (state) => {
    expect(activityForState(state)).toBe('none')
  })

  it('maps a missing snapshot to "none"', () => {
    expect(activityForState(undefined)).toBe('none')
    expect(activityForState(null)).toBe('none')
  })
})

describe('useGameActivityStore', () => {
  it('starts at "none" and updates through setActivity', () => {
    useGameActivityStore.setState({ activity: 'none' })
    useGameActivityStore.getState().setActivity('playing')
    expect(useGameActivityStore.getState().activity).toBe('playing')
  })
})
