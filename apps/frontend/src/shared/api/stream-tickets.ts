import { apiClient } from '@/shared/api/client'
import type { StreamTicketResponse } from '@/shared/contracts/dto'

// Both SSE (EventSource) and the game WebSocket authenticate a long-lived
// connection with a short-lived, single-use ticket instead of the access
// token itself - neither connection type can set an Authorization header,
// so the token would otherwise sit in the URL (and every access log) for
// its whole lifetime. The interceptors already attach the real
// Authorization header to these two POSTs. See
// ObsidianVault/stream-connection-tickets-2026-09-02.md.

export async function mintEventsTicket(): Promise<string> {
  const response = await apiClient.post<StreamTicketResponse>('/events/ticket')
  return response.data.ticket
}

export async function mintGameSocketTicket(gameId: string): Promise<string> {
  const response = await apiClient.post<StreamTicketResponse>(`/games/${gameId}/ws-ticket`)
  return response.data.ticket
}
