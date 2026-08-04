import { queryKeys } from '@/shared/api/query-keys'

export const profileKeys = {
  me: [...queryKeys.root, 'profile', 'me'] as const,
}
