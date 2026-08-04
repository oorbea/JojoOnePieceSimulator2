import { z } from 'zod'

import type { PictureStatus, Role } from '@/shared/lib/zod'

// Mirrors the backend's dto.UserResponse (apps/backend .../dto/user_response.go)
// once the /users/me routes land. `avatar`/`avatarThumb` are presigned URLs
// (own upload if present, else the Google picture) — never object-storage
// keys, and never null: an unset avatar is "".
export type ProfileUser = {
  id: string
  email: string
  username: string
  completeName: string
  avatar: string
  avatarThumb: string
  avatarStatus: PictureStatus
  role: Role
}

export type UpdateUsernameInput = {
  username: string
}

// Mirrors the backend's username sanitizer client-side, so a typo surfaces
// as an inline field error instead of a round-trip 400.
export const usernameFormSchema = z.object({
  username: z
    .string()
    .min(3, 'Username must be at least 3 characters')
    .max(32, 'Username must be at most 32 characters')
    .regex(/^[a-z0-9_]+$/, 'Only lowercase letters, numbers, and underscores'),
})

export type UsernameFormValues = z.infer<typeof usernameFormSchema>
