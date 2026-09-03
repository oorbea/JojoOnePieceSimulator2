import { env } from '@/shared/config/env'

// buildGameSocketUrl mirrors the backend's WS auth: a short-lived, single-use
// ticket in the query string, since neither browser WebSocket nor RN's can
// set headers (same reasoning as the SSE bridge's EventSource URL). The
// ticket is minted fresh per connect attempt via mintGameSocketTicket - see
// game-socket.store.ts. Returns null when EXPO_PUBLIC_SOCKET_URL isn't
// configured - the documented "realtime unavailable, fall back to REST
// polling" signal.
export function buildGameSocketUrl(gameId: string, ticket: string): string | null {
  if (!env.EXPO_PUBLIC_SOCKET_URL) return null
  return `${env.EXPO_PUBLIC_SOCKET_URL}/games/${gameId}/ws?ticket=${encodeURIComponent(ticket)}`
}
