import type { Locale, Role } from '@/shared/contracts/enums'
import type { UserResponse } from '@/shared/contracts/dto'

export type SessionUser = {
  id: string
  email: string
  username: string
  completeName: string
  picture: string | null
  role: Role
  language: Locale
}

// Shared by both call sites that turn the backend's UserResponse into the
// in-memory SessionUser shape - the login POST (use-google-auth.ts) and the
// silent-refresh POST (shared/api/refresh.ts) - so they can never drift.
// Split into its own module (rather than living in session.store.ts) so
// refresh.ts can import it without a circular dependency back to the store.
export function fromUserResponse(user: UserResponse): SessionUser {
  return {
    id: user.id,
    email: user.email,
    username: user.username,
    completeName: user.completeName,
    picture: user.avatar || null,
    role: user.role,
    language: user.language,
  }
}
