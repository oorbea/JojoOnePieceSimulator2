import { z } from 'zod'

import { powerTranslationsFormSchema, type TranslationFormValues } from '@/shared/lib/power-translations'
import { raritySchema, standStatSchema, type Locale, type PictureStatus, type Rarity, type StandStat } from '@/shared/lib/zod'

// Mirrors the backend's dto.StandResponse (apps/backend .../dto/stand_response.go).
// `evolvesFrom` nests recursively — the backend returns the full parent
// Stand, not just its id.
export type StandResponse = {
  id: string
  name: string
  description: string
  rarity: Rarity
  skills: string[]
  picture: string
  pictureThumb: string
  pictureStatus: PictureStatus
  attackPower: StandStat
  speed: StandStat
  attackRange: StandStat
  endurance: StandStat
  precision: StandStat
  potential: StandStat
  evolvesFrom: StandResponse | null
}

// Mirrors dto.StandRequest — same body shape for both create (POST) and
// update (PUT). Translations replaced the old flat description/skills once
// power_translations landed - see the vault's i18n-multi-language.md.
export type StandInput = {
  name: string
  translations: Partial<Record<Locale, TranslationFormValues>>
  rarity: Rarity
  attackPower: StandStat
  speed: StandStat
  attackRange: StandStat
  endurance: StandStat
  precision: StandStat
  potential: StandStat
  evolvesFromId: string | null
}

export const standFormSchema = z.object({
  name: z.string().min(1, 'validation.nameRequired').max(100, 'validation.nameTooLong'),
  translations: powerTranslationsFormSchema,
  rarity: raritySchema,
  attackPower: standStatSchema,
  speed: standStatSchema,
  attackRange: standStatSchema,
  endurance: standStatSchema,
  precision: standStatSchema,
  potential: standStatSchema,
  evolvesFromId: z.string().uuid().nullable(),
})

export type StandFormValues = z.infer<typeof standFormSchema>

export type StandFilters = {
  rarity?: Rarity
  attackPower?: StandStat
  speed?: StandStat
  attackRange?: StandStat
  endurance?: StandStat
  precision?: StandStat
  potential?: StandStat
  evolvesFrom?: string
}
