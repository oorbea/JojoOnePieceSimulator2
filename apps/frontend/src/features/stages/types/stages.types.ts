import { z } from 'zod'

import {
  stageTranslationsFormSchema,
  type StageTranslationFormValues,
} from '@/shared/lib/stage-translations'
import { mangaSchema, type Locale, type Manga, type PictureStatus } from '@/shared/lib/zod'

// Mirrors the backend's dto.StageResponse (apps/backend .../dto/stage_response.go).
export type StageResponse = {
  id: string
  manga: Manga
  order: number
  name: string
  description: string
  picture: string
  pictureThumb: string
  pictureStatus: PictureStatus
}

// Mirrors dto.StageRequest - same body shape for both create (POST) and
// update (PUT). Unlike StandInput's translations (only en-GB mandatory),
// every locale is required here - see the vault's game-stage-content.md.
export type StageInput = {
  manga: Manga
  order: number
  name: string
  translations: Record<Locale, StageTranslationFormValues>
}

export const stageFormSchema = z.object({
  manga: mangaSchema,
  order: z.number().int().min(0, 'validation.orderNonNegative'),
  name: z.string().min(1, 'validation.nameRequired').max(100, 'validation.nameTooLong'),
  translations: stageTranslationsFormSchema,
})

export type StageFormValues = z.infer<typeof stageFormSchema>

export type StageFilters = {
  manga?: Manga
}
