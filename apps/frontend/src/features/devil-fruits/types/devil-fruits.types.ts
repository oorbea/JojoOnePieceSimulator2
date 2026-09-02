import { z } from 'zod'

import {
  powerTranslationsFormSchema,
  type TranslationFormValues,
} from '@/shared/lib/power-translations'
import {
  fruitTypeSchema,
  raritySchema,
  type FruitType,
  type Locale,
  type PictureStatus,
  type Rarity,
} from '@/shared/contracts/enums'

// Mirrors the backend's dto.DevilFruitResponse.
export type DevilFruitResponse = {
  id: string
  name: string
  description: string
  rarity: Rarity
  skills: string[]
  picture: string
  pictureThumb: string
  pictureStatus: PictureStatus
  fruitType: FruitType
}

// Mirrors dto.DevilFruitRequest — same body for create (POST) and update
// (PUT). Translations replaced the old flat description/skills - see
// stands.types.ts's StandInput for the identical reasoning.
export type DevilFruitInput = {
  name: string
  translations: Partial<Record<Locale, TranslationFormValues>>
  rarity: Rarity
  fruitType: FruitType
}

export const devilFruitFormSchema = z.object({
  name: z.string().min(1, 'validation.nameRequired').max(100, 'validation.nameTooLong'),
  translations: powerTranslationsFormSchema,
  rarity: raritySchema,
  fruitType: fruitTypeSchema,
})

export type DevilFruitFormValues = z.infer<typeof devilFruitFormSchema>

export type DevilFruitFilters = {
  rarity?: Rarity
  fruitType?: FruitType
  q?: string
}
