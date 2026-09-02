import { z } from 'zod'

// ProfileUser was a byte-identical hand-mirror of dto.UserResponse; now a
// straight rename re-export of the generated type.
export type { UserResponse as ProfileUser } from '@/shared/contracts/dto'

// UpdateUsernameInput/UpdateLanguageInput used to hand-mirror
// dto.UpdateProfileRequest here, but nothing actually imported them -
// profile.api.ts's updateUsername/updateLanguage build their request
// bodies as inline object literals. Removed rather than migrated to a
// generated re-export, per the "no wire type written by hand" rule: a
// consumer that needs the request shape should import
// UpdateProfileRequest from '@/shared/contracts/dto' directly.

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
