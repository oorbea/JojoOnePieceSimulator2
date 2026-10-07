import { cueForStateChange } from '../haptic-cues'

describe('cueForStateChange', () => {
  it('buzzes when voting or a tiebreak opens', () => {
    expect(cueForStateChange('SUMMARY', 'VOTING')).toBe('votingOpened')
    expect(cueForStateChange('VOTING', 'TIEBREAK')).toBe('votingOpened')
    expect(cueForStateChange('RESOLVING', 'VOTING')).toBe('votingOpened')
  })

  it('buzzes when the round resolves and when the game finishes', () => {
    expect(cueForStateChange('VOTING', 'RESOLVING')).toBe('roundResolved')
    expect(cueForStateChange('TIEBREAK', 'RESOLVING')).toBe('roundResolved')
    expect(cueForStateChange('RESOLVING', 'FINISHED')).toBe('gameFinished')
  })

  it('stays silent for the first state seen (load, reconnect, resume)', () => {
    expect(cueForStateChange(undefined, 'VOTING')).toBeNull()
    expect(cueForStateChange(undefined, 'FINISHED')).toBeNull()
  })

  it('stays silent when nothing changed', () => {
    expect(cueForStateChange('VOTING', 'VOTING')).toBeNull()
  })

  it('stays silent for states with no cue', () => {
    expect(cueForStateChange('LOBBY', 'ASSIGNING')).toBeNull()
    expect(cueForStateChange('ASSIGNING', 'SUMMARY')).toBeNull()
    expect(cueForStateChange('VOTING', 'ABORTED')).toBeNull()
  })
})
