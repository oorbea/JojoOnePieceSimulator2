import type { StandResponse } from '@/features/stands'
import { standEvolutionChain, standEvolutionSteps } from '../stand-evolution'

function stand(id: string, evolvesFrom: StandResponse | null = null): StandResponse {
  return { id, name: id, evolvesFrom } as StandResponse
}

describe('standEvolutionChain', () => {
  it('returns a single-element chain for a stand with no evolvesFrom', () => {
    const base = stand('star-platinum')
    expect(standEvolutionChain(base)).toEqual([base])
  })

  it('returns the chain root-first for a one-step evolution', () => {
    const base = stand('silver-chariot')
    const evolved = stand('chariot-requiem', base)
    expect(standEvolutionChain(evolved).map((s) => s.id)).toEqual([
      'silver-chariot',
      'chariot-requiem',
    ])
  })

  it('returns the chain root-first for a multi-step evolution', () => {
    const egg = stand('echoes-huevo')
    const act1 = stand('echoes-act1', egg)
    const act2 = stand('echoes-act2', act1)
    const act3 = stand('echoes-act3', act2)
    expect(standEvolutionChain(act3).map((s) => s.id)).toEqual([
      'echoes-huevo',
      'echoes-act1',
      'echoes-act2',
      'echoes-act3',
    ])
  })

  it('breaks out of a malformed cycle instead of looping forever', () => {
    const a = stand('a')
    const b = stand('b', a)
    // Force a cycle a malformed/legacy row could produce - the backend
    // itself refuses to build or store one.
    ;(a as { evolvesFrom: StandResponse | null }).evolvesFrom = b
    expect(standEvolutionChain(b).map((s) => s.id)).toEqual(['a', 'b'])
  })
})

describe('standEvolutionSteps', () => {
  it('is 0 for null/undefined', () => {
    expect(standEvolutionSteps(null)).toBe(0)
    expect(standEvolutionSteps(undefined)).toBe(0)
  })

  it('is 0 for a stand with no evolvesFrom', () => {
    expect(standEvolutionSteps(stand('star-platinum'))).toBe(0)
  })

  it('is 1 for a one-step evolution', () => {
    const evolved = stand('chariot-requiem', stand('silver-chariot'))
    expect(standEvolutionSteps(evolved)).toBe(1)
  })

  it('is 3 for a four-stage chain', () => {
    const egg = stand('echoes-huevo')
    const act1 = stand('echoes-act1', egg)
    const act2 = stand('echoes-act2', act1)
    const act3 = stand('echoes-act3', act2)
    expect(standEvolutionSteps(act3)).toBe(3)
  })
})
