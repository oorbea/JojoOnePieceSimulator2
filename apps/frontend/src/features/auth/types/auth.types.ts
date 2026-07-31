import type { SessionUser } from '@/shared/stores/session.store'

export type AuthGoogleResponse = {
  accessToken: string
  tokenType: string
  expiresAt: string
  user: SessionUser
}
