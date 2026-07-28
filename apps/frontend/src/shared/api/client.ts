import axios from 'axios'

import { env } from '@/shared/config/env'
import { registerInterceptors } from '@/shared/api/interceptors'

// eslint-disable-next-line import/no-named-as-default-member -- default import is correct; axios's named `create` export is unrelated
export const apiClient = axios.create({
  baseURL: env.EXPO_PUBLIC_API_URL,
  timeout: 15_000,
  headers: {
    'Content-Type': 'application/json',
  },
  // A 304 is a normal, expected outcome of the conditional-GET flow (see
  // shared/api/etag.ts) — treat it as success instead of throwing.
  validateStatus: (status) => (status >= 200 && status < 300) || status === 304,
})

registerInterceptors(apiClient)
