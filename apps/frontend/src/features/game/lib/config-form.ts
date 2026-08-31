import type { GameConfig, PoolFilter, UpdateGameConfigInput } from '@/features/game/types/game.types'
import type {
  AbilitySource,
  GameMode,
  LobbyVisibility,
  Manga,
  RevealSpeed,
} from '@/shared/lib/zod'

// Local edit-form state for the config panel, seeded from snapshot.config
// (and snapshot.mode, which lives outside GameConfig) whenever the lobby
// hasn't been locally edited yet. Kept separate from the live snapshot so
// typing/toggling doesn't fight incoming STATE frames from other clients.
export type ConfigFormState = {
  mode: GameMode
  stageMangas: Manga[]
  powerMangas: Manga[]
  teamSize: number
  allowBots: boolean
  visibility: LobbyVisibility
  votingWindowSeconds: number
  summaryDurationSeconds: number
  revealSpeed: RevealSpeed
  poolFilter: PoolFilter
}

// Single source of truth for per-mode team size bounds - shared by the
// create-lobby form and the in-lobby config panel so they never drift apart.
export const TEAM_SIZE_LIMITS: Record<GameMode, { min: number; max: number }> = {
  GAUNTLET: { min: 1, max: 10 },
  VERSUS: { min: 1, max: 5 },
}

export function clampTeamSize(mode: GameMode, teamSize: number): number {
  const { min, max } = TEAM_SIZE_LIMITS[mode]
  return Math.min(Math.max(teamSize, min), max)
}

export function configFormFromSnapshot(mode: GameMode, config: GameConfig): ConfigFormState {
  return {
    mode,
    stageMangas: config.stageMangas,
    powerMangas: config.powerMangas,
    teamSize: config.teamSize,
    allowBots: config.allowBots,
    visibility: config.visibility,
    votingWindowSeconds: config.votingWindowSeconds,
    summaryDurationSeconds: config.summaryDurationSeconds,
    revealSpeed: config.revealSpeed,
    poolFilter: config.poolFilter,
  }
}

export function applyModeChange(form: ConfigFormState, mode: GameMode): ConfigFormState {
  if (mode === 'GAUNTLET') {
    return {
      ...form,
      mode,
      allowBots: false,
      teamSize: clampTeamSize(mode, form.teamSize),
    }
  }
  return { ...form, mode, teamSize: clampTeamSize(mode, form.teamSize) }
}

// UPDATE_CONFIG is a full replacement (mirrors CreateGameRequest, plus the
// fields only editable once a lobby exists) - always build the whole payload
// from current + edited fields, never a patch.
export function buildUpdateConfigPayload(
  form: ConfigFormState,
  abilitySource: AbilitySource
): UpdateGameConfigInput {
  return {
    mode: form.mode,
    stageMangas: form.stageMangas,
    powerMangas: form.powerMangas,
    abilitySource,
    teamSize: form.teamSize,
    allowBots: form.allowBots,
    visibility: form.visibility,
    votingWindowSeconds: form.votingWindowSeconds,
    summaryDurationSeconds: form.summaryDurationSeconds,
    revealSpeed: form.revealSpeed,
    poolFilter: form.poolFilter,
  }
}
