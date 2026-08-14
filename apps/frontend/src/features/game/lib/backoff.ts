// Same backoff curve as src/providers/picture-events-bridge.tsx's SSE
// reconnect (2s -> 4s -> 8s ... capped at 30s) - kept as its own function so
// both the socket store and its test read one definition.
export const BASE_RECONNECT_MS = 2_000
export const MAX_RECONNECT_MS = 30_000

export function reconnectDelay(attempt: number): number {
  return Math.min(BASE_RECONNECT_MS * 2 ** attempt, MAX_RECONNECT_MS)
}
