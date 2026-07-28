import type {
  ErrorResponse,
  FruitType,
  PictureStatus,
  Rarity,
  Role,
  StandStat,
} from '@/shared/lib/zod'

// Shared DTO primitives mirroring apps/backend/internal/infrastructure/api/dto.
// Feature-specific response shapes (StandResponse, DevilFruitResponse, ...)
// belong in each feature's own types/ folder and compose these.
export type { ErrorResponse, FruitType, PictureStatus, Rarity, Role, StandStat }

export type UserResponse = {
  id: string
  email: string
  username: string
  completeName: string
  picture: string | null
  role: Role
}
