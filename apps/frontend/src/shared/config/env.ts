import { z } from 'zod'

// Parsed once at module load so a misconfigured Docker/CI environment fails
// loudly at boot with a readable message instead of silently producing
// `undefined` request URLs deep inside the app.
const envSchema = z.object({
  EXPO_PUBLIC_API_URL: z.string().url(),
  EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID: z.string().min(1),
  EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID: z.string().optional(),
  EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID: z.string().optional(),
  EXPO_PUBLIC_SOCKET_URL: z.string().optional(),
  // Distinct per deploy (commit SHA in CI, "dev" locally) - used as the
  // React Query persister's buster (query-provider.tsx) so a new build
  // never rehydrates a persisted cache shaped for an older one.
  EXPO_PUBLIC_BUILD_ID: z.string().default('dev'),
})

const parsed = envSchema.safeParse({
  EXPO_PUBLIC_API_URL: process.env.EXPO_PUBLIC_API_URL,
  EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID,
  EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
  EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  EXPO_PUBLIC_SOCKET_URL: process.env.EXPO_PUBLIC_SOCKET_URL,
  EXPO_PUBLIC_BUILD_ID: process.env.EXPO_PUBLIC_BUILD_ID,
})

if (!parsed.success) {
  throw new Error(`Invalid environment configuration: ${parsed.error.message}`)
}

export const env = parsed.data
