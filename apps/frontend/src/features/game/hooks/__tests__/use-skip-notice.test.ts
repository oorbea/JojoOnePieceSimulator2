import { skippedPhase, trackPhase } from '../use-skip-notice'

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

describe('trackPhase', () => {
  it('keeps the remembered deadline when it is cleared before the state flips', () => {
    // VOTING_OPENED nulls summaryEndsAt while the snapshot is still SUMMARY.
    const inSummary = trackPhase({ state: 'SUMMARY', endsAt: 5000 }, 'SUMMARY', null)
    expect(inSummary).toEqual({ state: 'SUMMARY', endsAt: 5000 })
    expect(skippedPhase(inSummary.state, 'VOTING', inSummary.endsAt, 1000)).toBe('summary')
  })

  it('adopts a new deadline and resets on a state change', () => {
    expect(trackPhase({ state: 'SUMMARY', endsAt: 5000 }, 'SUMMARY', 6000).endsAt).toBe(6000)
    expect(trackPhase({ state: 'SUMMARY', endsAt: 5000 }, 'VOTING', null)).toEqual({
      state: 'VOTING',
      endsAt: null,
    })
  })
})
