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
  // Origin lobby invite links are built under (buildInviteUrl,
  // features/game/lib/invite-url.ts). Optional: web falls back to
  // window.location.origin, native with nothing configured just can't share
  // a link (share.ts falls back to the raw code).
  EXPO_PUBLIC_WEB_ORIGIN: z.string().optional(),
  // Distinct per deploy (commit SHA in CI, "dev" locally) - used as the
  // React Query persister's buster (query-provider.tsx) so a new build
  // never rehydrates a persisted cache shaped for an older one.
  EXPO_PUBLIC_BUILD_ID: z.string().default('dev'),
  // Renders the /dev-login screen (see features/auth's dev-login-container)
  // - a local-only stand-in for Google sign-in, backed by the equally
  // dev-only POST /auth/dev-login (apps/backend's DEV_AUTH_BYPASS). Only
  // ever "true" in docker-compose.dev.yml's build args; unset/anything else
  // means the route redirects straight to /login. This flag only controls
  // whether the UI renders - the backend route being unreachable outside
  // dev is the real defense (see config.Load's boot guard).
  EXPO_PUBLIC_DEV_AUTH: z
    .string()
    .optional()
    .transform((raw) => raw === 'true'),
})

const parsed = envSchema.safeParse({
  EXPO_PUBLIC_API_URL: process.env.EXPO_PUBLIC_API_URL,
  EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID: process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID,
  EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID: process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
  EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID: process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  EXPO_PUBLIC_SOCKET_URL: process.env.EXPO_PUBLIC_SOCKET_URL,
  EXPO_PUBLIC_WEB_ORIGIN: process.env.EXPO_PUBLIC_WEB_ORIGIN,
  EXPO_PUBLIC_BUILD_ID: process.env.EXPO_PUBLIC_BUILD_ID,
  EXPO_PUBLIC_DEV_AUTH: process.env.EXPO_PUBLIC_DEV_AUTH,
})

if (!parsed.success) {
  throw new Error(`Invalid environment configuration: ${parsed.error.message}`)
}

export const env = parsed.data
