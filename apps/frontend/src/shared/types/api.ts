import type { FruitType, Locale, PictureStatus, Rarity, Role, StandStat } from '@/shared/contracts/enums'
import type { ErrorResponse } from '@/shared/contracts/errors'

// Shared DTO primitives mirroring apps/backend/internal/infrastructure/api/dto.
// Feature-specific response shapes (StandResponse, DevilFruitResponse, ...)
// belong in each feature's own types/ folder and compose these.
export type { ErrorResponse, FruitType, Locale, PictureStatus, Rarity, Role, StandStat }

export type UserResponse = {
  id: string
  email: string
  username: string
  completeName: string
  avatar: string
  avatarThumb: string
  avatarStatus: PictureStatus
  role: Role
  language: Locale
}
