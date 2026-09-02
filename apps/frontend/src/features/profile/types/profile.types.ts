import { z } from 'zod'

import type { Locale } from '@/shared/contracts/enums'
// ProfileUser was a byte-identical hand-mirror of dto.UserResponse; now a
// straight rename re-export of the generated type.
export type { UserResponse as ProfileUser } from '@/shared/contracts/dto'

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
