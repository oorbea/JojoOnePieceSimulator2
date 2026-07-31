import { apiClient } from '@/shared/api/client'
import type { AuthGoogleResponse } from '@/features/auth/types/auth.types'

export async function postGoogleAuth(idToken: string): Promise<AuthGoogleResponse> {
  const response = await apiClient.post<AuthGoogleResponse>('/api/v1/auth/google', { idToken })
  return response.data
}
