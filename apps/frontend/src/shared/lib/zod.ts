import { z } from 'zod'

// Backend enums (apps/backend/internal/domain/enums) mirrored here so every
// feature validates against the same source of truth instead of redeclaring
// string unions ad hoc.
export const raritySchema = z.enum(['COMMON', 'RARE', 'EPIC', 'LEGENDARY'])
export const standStatSchema = z.enum(['E', 'D', 'C', 'B', 'A', 'INFINITE', 'NULL'])
export const fruitTypeSchema = z.enum([
  'PARAMECIA',
  'ZOAN',
  'LOGIA',
  'SPECIAL_PARAMECIA',
  'ANCIENT_ZOAN',
  'MYTHICAL_ZOAN',
])
export const roleSchema = z.enum(['REGULAR', 'ADMIN'])
export const pictureStatusSchema = z.enum(['NONE', 'PENDING', 'READY', 'FAILED'])
// Mirrors the backend's enums.Locale (apps/backend .../domain/enums/locale.go).
export const localeSchema = z.enum(['en-GB', 'es-ES', 'ca-ES'])
// Mirrors the backend's enums.Manga (apps/backend .../domain/enums/manga.go).
export const mangaSchema = z.enum(['JOJO', 'ONE_PIECE'])
// Mirrors the backend's enums.GameModeKind.
export const gameModeSchema = z.enum(['GAUNTLET', 'VERSUS'])
// Mirrors the backend's enums.AbilitySource. INVENTORY is parseable but
// rejected server-side (game.ErrInventoryNotSupported) - kept here so a
// stale/legacy value still round-trips instead of failing validation.
export const abilitySourceSchema = z.enum(['RANDOM', 'INVENTORY'])
// Mirrors the backend's enums.GameState.
export const gameStateSchema = z.enum([
  'LOBBY',
  'ASSIGNING',
  'VOTING',
  'TIEBREAK',
  'RESOLVING',
  'FINISHED',
  'ABORTED',
])
// Mirrors the backend's enums.ParticipantKind.
export const participantKindSchema = z.enum(['HUMAN', 'BOT'])
// Mirrors the backend's enums.LobbyVisibility.
export const lobbyVisibilitySchema = z.enum(['PUBLIC', 'PRIVATE'])

export const errorResponseSchema = z.object({
  error: z.string(),
  code: z.string().optional(),
  details: z.array(z.string()).optional(),
})

export type Rarity = z.infer<typeof raritySchema>
export type StandStat = z.infer<typeof standStatSchema>
export type FruitType = z.infer<typeof fruitTypeSchema>
export type Role = z.infer<typeof roleSchema>
export type PictureStatus = z.infer<typeof pictureStatusSchema>
export type Locale = z.infer<typeof localeSchema>
export type Manga = z.infer<typeof mangaSchema>
export type GameMode = z.infer<typeof gameModeSchema>
export type AbilitySource = z.infer<typeof abilitySourceSchema>
export type GameState = z.infer<typeof gameStateSchema>
export type ParticipantKind = z.infer<typeof participantKindSchema>
export type LobbyVisibility = z.infer<typeof lobbyVisibilitySchema>
export type ErrorResponse = z.infer<typeof errorResponseSchema>
