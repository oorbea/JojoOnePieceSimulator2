import type { UserResponse } from '@/shared/contracts/dto'

export type AuthGoogleResponse = {
  accessToken: string
  tokenType: string
  expiresAt: string
  user: UserResponse
}
