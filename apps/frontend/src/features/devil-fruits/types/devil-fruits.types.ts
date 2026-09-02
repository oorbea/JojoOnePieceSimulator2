import { z } from 'zod'

import { powerTranslationsFormSchema } from '@/shared/lib/power-translations'
import { fruitTypeSchema, raritySchema, type FruitType, type Rarity } from '@/shared/contracts/enums'

// DevilFruitResponse/DevilFruitInput are generated
// (dto.DevilFruitResponse/DevilFruitRequest) - Go decides the shape, this
// is a rename re-export. See ObsidianVault/contratos-tipos-generados.md.
export type { DevilFruitResponse, DevilFruitRequest as DevilFruitInput } from '@/shared/contracts/dto'

export const devilFruitFormSchema = z.object({
  name: z.string().min(1, 'validation.nameRequired').max(100, 'validation.nameTooLong'),
  translations: powerTranslationsFormSchema,
  rarity: raritySchema,
  fruitType: fruitTypeSchema,
})

export type DevilFruitFormValues = z.infer<typeof devilFruitFormSchema>

// Not generated - same reasoning as StandFilters (stands.types.ts):
// ports.DevilFruitFilters is server-side only, never marshaled.
export type DevilFruitFilters = {
  rarity?: Rarity
  fruitType?: FruitType
  q?: string
}
