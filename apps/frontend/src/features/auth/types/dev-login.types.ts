import { z } from 'zod'

// Mirrors the backend's devLoginNamePattern (apps/backend's
// internal/infrastructure/api/dto/auth_request.go) client-side, so a typo
// surfaces as an inline field error instead of a round-trip 400. Message is
// an i18n key, not a display string - see usernameFormSchema's same
// decision in profile.types.ts; resolve with t(errors.name?.message) at the
// render site.
export const devLoginFormSchema = z.object({
  name: z
    .string()
    .min(1, 'devLogin.invalidName')
    .max(24, 'devLogin.invalidName')
    .regex(/^[a-z0-9_]+$/, 'devLogin.invalidName'),
  admin: z.boolean(),
})

export type DevLoginFormValues = z.infer<typeof devLoginFormSchema>
