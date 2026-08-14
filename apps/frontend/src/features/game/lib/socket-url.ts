import { env } from '@/shared/config/env'

// buildGameSocketUrl mirrors the backend's WS auth (query-param token,
// since neither browser WebSocket nor RN's can set headers - same
// reasoning as the SSE bridge's EventSource URL). Returns null when
// EXPO_PUBLIC_SOCKET_URL isn't configured - the documented "realtime
// unavailable, fall back to REST polling" signal.
export function buildGameSocketUrl(gameId: string, token: string): string | null {
  if (!env.EXPO_PUBLIC_SOCKET_URL) return null
  return `${env.EXPO_PUBLIC_SOCKET_URL}/games/${gameId}/ws?token=${encodeURIComponent(token)}`
}
