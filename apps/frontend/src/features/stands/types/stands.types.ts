import { z } from 'zod'

import { raritySchema, standStatSchema, type PictureStatus, type Rarity, type StandStat } from '@/shared/lib/zod'

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
// update (PUT).
export type StandInput = {
  name: string
  description: string
  rarity: Rarity
  skills: string[]
  attackPower: StandStat
  speed: StandStat
  attackRange: StandStat
  endurance: StandStat
  precision: StandStat
  potential: StandStat
  evolvesFromId: string | null
}

export const standFormSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name is too long'),
  description: z.string().min(1, 'Description is required').max(1000, 'Description is too long'),
  rarity: raritySchema,
  skills: z.array(z.string().min(1)).min(1, 'At least one skill is required'),
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
