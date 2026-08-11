import { useRouter } from 'expo-router'

import { AdminHubScreen } from '@/features/admin/components/presentational/admin-hub-screen'

export function AdminHubContainer() {
  const router = useRouter()

  return (
    <AdminHubScreen
      onOpenStands={() => router.navigate('/admin/stands' as never)}
      onOpenDevilFruits={() => router.navigate('/admin/devil-fruits' as never)}
      onOpenStages={() => router.navigate('/admin/stages' as never)}
    />
  )
}
