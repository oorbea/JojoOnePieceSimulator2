import { z } from 'zod'

import { characterTranslationsFormSchema } from '@/shared/lib/character-translations'
import {
  fruitMasterySchema,
  hakiLevelSchema,
  hamonLevelSchema,
  physicalFormSchema,
  powerRaritySchema,
  spinLevelSchema,
} from '@/shared/contracts/enums'

// JojoCharacterResponse/Request and OnePieceCharacterResponse/Request are
// generated (dto.JojoCharacterResponse/Request, ...) - Go decides the
// shape, these are rename re-exports. See
// ObsidianVault/contratos-tipos-generados.md.
export type {
  JojoCharacterResponse,
  JojoCharacterRequest as JojoCharacterInput,
  OnePieceCharacterResponse,
  OnePieceCharacterRequest as OnePieceCharacterInput,
} from '@/shared/contracts/dto'

// A CharacterKind is not a generated enum (see battle-iq.ts's doc on
// enums.battleIQCategory for the same convention) - it only ever exists on
// the client, to pick which pair of endpoints/screens a manga-exclusive
// selection resolves to. 'JOJO'/'ONE_PIECE' match enums.Manga's own string
// values so components can share one selector with StageFormModal-style
// manga pickers.
export type CharacterKind = 'JOJO' | 'ONE_PIECE'

export const jojoCharacterFormSchema = z.object({
  name: z.string().min(1, 'validation.nameRequired').max(100, 'validation.nameTooLong'),
  rarity: powerRaritySchema,
  hamon: hamonLevelSchema,
  spin: spinLevelSchema,
  battleIq: z
    .number()
    .int()
    .min(0, 'validation.battleIqRange')
    .max(255, 'validation.battleIqRange'),
  translations: characterTranslationsFormSchema,
})
export type JojoCharacterFormValues = z.infer<typeof jojoCharacterFormSchema>

export const onePieceCharacterFormSchema = z.object({
  name: z.string().min(1, 'validation.nameRequired').max(100, 'validation.nameTooLong'),
  rarity: powerRaritySchema,
  physicalForm: physicalFormSchema,
  armamentHaki: hakiLevelSchema,
  observationHaki: hakiLevelSchema,
  conquerorHaki: hakiLevelSchema,
  fruitMastery: fruitMasterySchema,
  translations: characterTranslationsFormSchema,
})
export type OnePieceCharacterFormValues = z.infer<typeof onePieceCharacterFormSchema>

// Not generated - server-side-only filters, same reasoning as
// stages.types.ts's StageFilters.
export type JojoCharacterFilters = {
  rarity?: z.infer<typeof powerRaritySchema>
  q?: string
}
export type OnePieceCharacterFilters = JojoCharacterFilters
