import { z } from 'zod'

import type { Locale, PictureStatus, Role } from '@/shared/contracts/enums'

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
  language: Locale
}

export type UpdateUsernameInput = {
  username: string
}

export type UpdateLanguageInput = {
  language: Locale
}

// Mirrors the backend's username sanitizer client-side, so a typo surfaces
// as an inline field error instead of a round-trip 400.
// Messages are i18n keys, not display strings - see power-translations.ts's
// zod decision. Resolve with t(errors.username?.message) at the render site.
export const usernameFormSchema = z.object({
  username: z
    .string()
    .min(3, 'validation.usernameMinLength')
    .max(32, 'validation.usernameMaxLength')
    .regex(/^[a-z0-9_]+$/, 'validation.usernameFormat'),
})

export type UsernameFormValues = z.infer<typeof usernameFormSchema>
