import { configFormFromSnapshot, isConfigFormDirty } from '../config-form'
import type { GameConfig } from '@/features/game/types/game.types'

const config: GameConfig = {
  stageMangas: ['JOJO', 'ONE_PIECE'],
  powerMangas: ['JOJO', 'ONE_PIECE'],
  abilitySource: 'RANDOM',
  teamSize: 5,
  allowBots: false,
  visibility: 'PRIVATE',
  votingWindowSeconds: 30,
  poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
  revealSpeed: 'NORMAL',
  summaryDurationSeconds: 60,
} as GameConfig

describe('isConfigFormDirty', () => {
  it('is clean right after seeding from the snapshot', () => {
    const form = configFormFromSnapshot('GAUNTLET', config)
    expect(isConfigFormDirty(form, 'GAUNTLET', config)).toBe(false)
  })

  it('flags an edited scalar field', () => {
    const form = { ...configFormFromSnapshot('GAUNTLET', config), votingWindowSeconds: 31 }
    expect(isConfigFormDirty(form, 'GAUNTLET', config)).toBe(true)
  })

  it('flags an added ban', () => {
    const base = configFormFromSnapshot('GAUNTLET', config)
    const form = { ...base, poolFilter: { ...base.poolFilter, banned: ['abc'] } }
    expect(isConfigFormDirty(form, 'GAUNTLET', config)).toBe(true)
  })

  it('goes clean again once the snapshot catches up with the form', () => {
    const form = { ...configFormFromSnapshot('GAUNTLET', config), summaryDurationSeconds: 61 }
    const confirmed = { ...config, summaryDurationSeconds: 61 }
    expect(isConfigFormDirty(form, 'GAUNTLET', confirmed)).toBe(false)
  })
})
