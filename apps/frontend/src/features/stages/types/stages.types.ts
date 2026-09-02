import { z } from 'zod'

import { stageTranslationsFormSchema } from '@/shared/lib/stage-translations'
import { mangaSchema, type Manga } from '@/shared/contracts/enums'

// StageResponse/StageInput are generated (dto.StageResponse/StageRequest) -
// Go decides the shape, this is a rename re-export. See
// ObsidianVault/contratos-tipos-generados.md.
export type { StageResponse, StageRequest as StageInput } from '@/shared/contracts/dto'

export const stageFormSchema = z.object({
  manga: mangaSchema,
  order: z.number().int().min(0, 'validation.orderNonNegative'),
  name: z.string().min(1, 'validation.nameRequired').max(100, 'validation.nameTooLong'),
  translations: stageTranslationsFormSchema,
})

export type StageFormValues = z.infer<typeof stageFormSchema>

// Not generated - same reasoning as stands.types.ts's StandFilters:
// ports.StageFilters is server-side only, never marshaled.
export type StageFilters = {
  manga?: Manga
  q?: string
}
