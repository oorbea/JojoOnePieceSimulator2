import { z } from 'zod'

import { powerTranslationsFormSchema } from '@/shared/lib/power-translations'
import { raritySchema, standStatSchema, type Rarity, type StandStat } from '@/shared/contracts/enums'

// StandResponse/StandInput are generated (dto.StandResponse/StandRequest) -
// Go decides the shape, this is a rename re-export. See
// ObsidianVault/contratos-tipos-generados.md.
export type { StandResponse, StandRequest as StandInput } from '@/shared/contracts/dto'

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

// Not generated: ports.StandFilters (the Go type this shape corresponds to)
// carries no json tags and is never itself marshaled - it's built
// server-side from query params, so this describes what the client SENDS,
// not a wire response shape. Client-only convenience type.
export type StandFilters = {
  rarity?: Rarity
  attackPower?: StandStat
  speed?: StandStat
  attackRange?: StandStat
  endurance?: StandStat
  precision?: StandStat
  potential?: StandStat
  evolvesFrom?: string
  q?: string
}
