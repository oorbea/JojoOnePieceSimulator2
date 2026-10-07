import { assertContract } from '@/shared/api/assert-contract'
import { apiClient } from '@/shared/api/client'
import {
  pushConfigResponseSchema,
  type PushConfigResponse,
  type PushSubscribeRequest,
} from '@/shared/contracts/dto'

export async function getPushConfig(): Promise<PushConfigResponse> {
  const response = await apiClient.get<PushConfigResponse>('/users/me/push')
  if (__DEV__) assertContract(pushConfigResponseSchema, response.data, 'GET /users/me/push')
  return response.data
}

export async function registerPushSubscription(body: PushSubscribeRequest): Promise<void> {
  await apiClient.post('/users/me/push/subscriptions', body)
}

export async function removePushSubscription(endpoint: string): Promise<void> {
  await apiClient.delete('/users/me/push/subscriptions', { data: { endpoint } })
}
