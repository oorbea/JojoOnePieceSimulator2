import { apiClient } from '@/shared/api/client'
import { assertContract } from '@/shared/api/assert-contract'
import { loginResponseSchema } from '@/shared/contracts/dto'
import type { AuthGoogleResponse } from '@/features/auth/types/auth.types'

export async function postGoogleAuth(idToken: string): Promise<AuthGoogleResponse> {
  const response = await apiClient.post<AuthGoogleResponse>('/auth/google', { idToken })
  if (__DEV__) assertContract(loginResponseSchema, response.data, 'POST /auth/google')
  return response.data
}
