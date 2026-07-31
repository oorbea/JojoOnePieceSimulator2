import { Redirect } from 'expo-router'

import { LoginContainer } from '@/features/auth'
import { useSessionStore } from '@/shared/stores/session.store'

export default function LoginRoute() {
  const session = useSessionStore((state) => state.session)
  const isHydrated = useSessionStore((state) => state.isHydrated)

  if (isHydrated && session) return <Redirect href="/" />

  return <LoginContainer />
}
