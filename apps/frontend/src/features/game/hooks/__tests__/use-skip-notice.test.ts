import { skippedPhase } from '../use-skip-notice'

describe('skippedPhase', () => {
  const endsAt = 100_000

  it('flags a sorteo that ended well before its deadline', () => {
    expect(skippedPhase('ASSIGNING', 'SUMMARY', endsAt, endsAt - 20_000)).toBe('reveal')
    expect(skippedPhase('ASSIGNING', 'VOTING', endsAt, endsAt - 20_000)).toBe('reveal')
  })

  it('flags a summary that ended well before its deadline', () => {
    expect(skippedPhase('SUMMARY', 'VOTING', endsAt, endsAt - 20_000)).toBe('summary')
  })

  it('stays silent when the phase ended on its own timer', () => {
    expect(skippedPhase('ASSIGNING', 'SUMMARY', endsAt, endsAt - 200)).toBeNull()
    expect(skippedPhase('SUMMARY', 'VOTING', endsAt, endsAt + 500)).toBeNull()
  })

  it('stays silent without a known deadline or a state change', () => {
    expect(skippedPhase('ASSIGNING', 'SUMMARY', null, 0)).toBeNull()
    expect(skippedPhase('SUMMARY', 'SUMMARY', endsAt, 0)).toBeNull()
    expect(skippedPhase('VOTING', 'RESOLVING', endsAt, 0)).toBeNull()
  })
})
