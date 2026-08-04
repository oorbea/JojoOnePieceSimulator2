import type { UserResponse } from '@/shared/types/api'

export type AuthGoogleResponse = {
  accessToken: string
  tokenType: string
  expiresAt: string
  user: UserResponse
}
