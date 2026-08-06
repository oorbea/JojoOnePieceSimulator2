import { z } from 'zod'

import { fruitTypeSchema, raritySchema, type FruitType, type PictureStatus, type Rarity } from '@/shared/lib/zod'

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

// Mirrors dto.DevilFruitRequest — same body for create (POST) and update (PUT).
export type DevilFruitInput = {
  name: string
  description: string
  rarity: Rarity
  skills: string[]
  fruitType: FruitType
}

export const devilFruitFormSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name is too long'),
  description: z.string().min(1, 'Description is required').max(1000, 'Description is too long'),
  rarity: raritySchema,
  skills: z.array(z.string().min(1)).min(1, 'At least one skill is required'),
  fruitType: fruitTypeSchema,
})

export type DevilFruitFormValues = z.infer<typeof devilFruitFormSchema>

export type DevilFruitFilters = {
  rarity?: Rarity
  fruitType?: FruitType
}
