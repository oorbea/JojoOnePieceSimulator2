// Backend enums and the error-response shape are now generated - Go is the
// single source of truth (apps/backend/cmd/typegen). This file is a thin
// re-export shim kept only so its ~49 existing importers don't all need to
// change import paths in the same commit; new code should import directly
// from '@/shared/contracts/enums' / '@/shared/contracts/errors' instead.
// See ObsidianVault/contratos-tipos-generados.md.
export {
  powerRaritySchema as raritySchema,
  standStatSchema,
  fruitTypeSchema,
  userRoleSchema as roleSchema,
  pictureStatusSchema,
  localeSchema,
  mangaSchema,
  gameModeKindSchema as gameModeSchema,
  abilitySourceSchema,
  gameStateSchema,
  participantKindSchema,
  lobbyVisibilitySchema,
  revealSpeedSchema,
  spinLevelSchema,
  hamonLevelSchema,
  fruitMasterySchema,
  hakiLevelSchema,
  physicalFormSchema,
} from '@/shared/contracts/enums'
export type {
  PowerRarity as Rarity,
  StandStat,
  FruitType,
  UserRole as Role,
  PictureStatus,
  Locale,
  Manga,
  GameModeKind as GameMode,
  AbilitySource,
  GameState,
  ParticipantKind,
  LobbyVisibility,
  RevealSpeed,
  SpinLevel,
  HamonLevel,
  FruitMastery,
  HakiLevel,
  PhysicalForm,
} from '@/shared/contracts/enums'
export { errorResponseSchema } from '@/shared/contracts/errors'
export type { ErrorResponse } from '@/shared/contracts/errors'
