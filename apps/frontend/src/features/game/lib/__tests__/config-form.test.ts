import {
  applyModeChange,
  buildUpdateConfigPayload,
  clampTeamSize,
  configFormFromSnapshot,
  TEAM_SIZE_LIMITS,
  type ConfigFormState,
} from '@/features/game/lib/config-form'
import type { GameConfig } from '@/features/game/types/game.types'

function baseForm(overrides: Partial<ConfigFormState> = {}): ConfigFormState {
  return {
    mode: 'GAUNTLET',
    stageMangas: ['JOJO'],
    powerMangas: ['JOJO'],
    teamSize: 5,
    allowBots: false,
    visibility: 'PRIVATE',
    votingWindowSeconds: 30,
    summaryDurationSeconds: 60,
    revealSpeed: 'NORMAL',
    poolFilter: { standRarities: [], fruitRarities: [], fruitTypes: [], banned: [] },
    ...overrides,
  }
}

function baseGameConfig(overrides: Partial<GameConfig> = {}): GameConfig {
  return {
    stageMangas: ['JOJO'],
    powerMangas: ['JOJO', 'ONE_PIECE'],
    abilitySource: 'RANDOM',
    teamSize: 4,
    allowBots: true,
    visibility: 'PUBLIC',
    votingWindowSeconds: 45,
    poolFilter: { standRarities: ['COMMON'], fruitRarities: [], fruitTypes: [], banned: ['x'] },
    revealSpeed: 'SWIFT',
    summaryDurationSeconds: 90,
    ...overrides,
  }
}

describe('buildUpdateConfigPayload', () => {
  it('builds a full payload with every field from the form plus the given abilitySource', () => {
    const form = baseForm()
    const payload = buildUpdateConfigPayload(form, 'RANDOM')

    expect(payload).toEqual({
      mode: form.mode,
      stageMangas: form.stageMangas,
      powerMangas: form.powerMangas,
      abilitySource: 'RANDOM',
      teamSize: form.teamSize,
      allowBots: form.allowBots,
      visibility: form.visibility,
      votingWindowSeconds: form.votingWindowSeconds,
      summaryDurationSeconds: form.summaryDurationSeconds,
      revealSpeed: form.revealSpeed,
      poolFilter: form.poolFilter,
    })
  })

  it('always echoes the whole form, not a diff against any prior value', () => {
    // Only teamSize differs from the "original" baseForm() - every other
    // field should still be carried through unchanged, in full.
    const form = baseForm({ teamSize: 3 })
    const payload = buildUpdateConfigPayload(form, 'INVENTORY')

    expect(payload.teamSize).toBe(3)
    expect(payload.mode).toBe(form.mode)
    expect(payload.stageMangas).toBe(form.stageMangas)
    expect(payload.powerMangas).toBe(form.powerMangas)
    expect(payload.allowBots).toBe(form.allowBots)
    expect(payload.visibility).toBe(form.visibility)
    expect(payload.votingWindowSeconds).toBe(form.votingWindowSeconds)
    expect(payload.summaryDurationSeconds).toBe(form.summaryDurationSeconds)
    expect(payload.revealSpeed).toBe(form.revealSpeed)
    expect(payload.poolFilter).toBe(form.poolFilter)
    expect(payload.abilitySource).toBe('INVENTORY')
  })
})

describe('clampTeamSize', () => {
  it('caps VERSUS at TEAM_SIZE_LIMITS.VERSUS.max (5)', () => {
    expect(clampTeamSize('VERSUS', 8)).toBe(TEAM_SIZE_LIMITS.VERSUS.max)
    expect(clampTeamSize('VERSUS', 8)).toBe(5)
  })

  it('floors VERSUS at TEAM_SIZE_LIMITS.VERSUS.min (1)', () => {
    expect(clampTeamSize('VERSUS', 0)).toBe(TEAM_SIZE_LIMITS.VERSUS.min)
    expect(clampTeamSize('VERSUS', 0)).toBe(1)
  })

  it('caps GAUNTLET at TEAM_SIZE_LIMITS.GAUNTLET.max (10)', () => {
    expect(clampTeamSize('GAUNTLET', 15)).toBe(TEAM_SIZE_LIMITS.GAUNTLET.max)
    expect(clampTeamSize('GAUNTLET', 15)).toBe(10)
  })

  it('floors GAUNTLET at TEAM_SIZE_LIMITS.GAUNTLET.min (1)', () => {
    expect(clampTeamSize('GAUNTLET', 0)).toBe(TEAM_SIZE_LIMITS.GAUNTLET.min)
    expect(clampTeamSize('GAUNTLET', 0)).toBe(1)
  })

  it('passes a mid-range value through unchanged for both modes', () => {
    expect(clampTeamSize('VERSUS', 3)).toBe(3)
    expect(clampTeamSize('GAUNTLET', 6)).toBe(6)
  })
})

describe('applyModeChange', () => {
  it('switching VERSUS -> GAUNTLET keeps an in-range teamSize but forces allowBots off', () => {
    const form = baseForm({ mode: 'VERSUS', teamSize: 5, allowBots: true })
    const next = applyModeChange(form, 'GAUNTLET')

    expect(next.mode).toBe('GAUNTLET')
    expect(next.teamSize).toBe(5)
    expect(next.allowBots).toBe(false)
  })

  it('switching GAUNTLET -> VERSUS clamps an out-of-range teamSize down to VERSUS max', () => {
    const form = baseForm({ mode: 'GAUNTLET', teamSize: 8, allowBots: true })
    const next = applyModeChange(form, 'VERSUS')

    expect(next.mode).toBe('VERSUS')
    expect(next.teamSize).toBe(5)
    // allowBots is untouched by a non-GAUNTLET target mode.
    expect(next.allowBots).toBe(true)
  })

  it('switching GAUNTLET(teamSize=10) -> VERSUS clamps down to 5', () => {
    const form = baseForm({ mode: 'GAUNTLET', teamSize: 10 })
    const next = applyModeChange(form, 'VERSUS')

    expect(next.teamSize).toBe(5)
  })

  it('preserves every other field untouched across a mode switch', () => {
    const form = baseForm({ mode: 'VERSUS', teamSize: 5 })
    const next = applyModeChange(form, 'GAUNTLET')

    expect(next.stageMangas).toBe(form.stageMangas)
    expect(next.powerMangas).toBe(form.powerMangas)
    expect(next.visibility).toBe(form.visibility)
    expect(next.votingWindowSeconds).toBe(form.votingWindowSeconds)
    expect(next.summaryDurationSeconds).toBe(form.summaryDurationSeconds)
    expect(next.revealSpeed).toBe(form.revealSpeed)
    expect(next.poolFilter).toBe(form.poolFilter)
  })
})

describe('configFormFromSnapshot', () => {
  it('maps every GameConfig field through, plus the separately-passed mode', () => {
    const config = baseGameConfig()
    const form = configFormFromSnapshot('VERSUS', config)

    expect(form).toEqual({
      mode: 'VERSUS',
      stageMangas: config.stageMangas,
      powerMangas: config.powerMangas,
      teamSize: config.teamSize,
      allowBots: config.allowBots,
      visibility: config.visibility,
      votingWindowSeconds: config.votingWindowSeconds,
      summaryDurationSeconds: config.summaryDurationSeconds,
      revealSpeed: config.revealSpeed,
      poolFilter: config.poolFilter,
    })
  })

  it('does not read mode off the config object itself (mode lives outside GameConfig)', () => {
    const config = baseGameConfig()
    expect('mode' in config).toBe(false)

    const form = configFormFromSnapshot('GAUNTLET', config)
    expect(form.mode).toBe('GAUNTLET')
  })
})
