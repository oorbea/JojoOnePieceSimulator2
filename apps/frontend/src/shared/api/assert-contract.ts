import type { z } from 'zod'

// Dev-only proof that a REST response still matches the generated contract
// (@/shared/contracts/dto). Never `.parse()`s in production: a 200 that
// fails validation would throw inside a TanStack Query queryFn, which
// retries by default and then surfaces an error boundary - turning a
// harmless backend field removal into a broken screen, three times over.
// The load-bearing check against drift is CI's `contracts` job (which
// catches it at the source, before a deploy can ever reach a client) - this
// is a second, local signal for whoever's driving the app in dev.
//
// Call this AFTER using the response, never in place of the real return
// value - it never throws and never changes behavior, only logs.
//
// Guard the call itself with `if (__DEV__)`, not just this function's body:
// the schema argument at a call site is a live expression, so an
// unconditional call would keep '@/shared/contracts/dto' in the release
// bundle's module graph regardless of what happens inside here.
export function assertContract<T>(schema: z.ZodType<T>, data: unknown, label: string): void {
  const result = schema.safeParse(data)
  if (!result.success) {
    // eslint-disable-next-line no-console
    console.error(`[contract] ${label} no longer matches the generated shape:`, result.error)
  }
}
