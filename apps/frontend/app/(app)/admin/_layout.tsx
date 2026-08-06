import { Redirect, Slot } from 'expo-router'

import { useSessionStore } from '@/shared/stores/session.store'

// Route-level guard for every screen under /admin — the nav item is already
// hidden from non-admins (see app-shell-container.tsx), but that alone
// doesn't stop someone from typing the URL directly, so this redirects any
// non-ADMIN session back home before Slot renders anything below it.
export default function AdminGroupLayout() {
  const session = useSessionStore((state) => state.session)

  if (!session) return <Redirect href="/login" />
  if (session.user.role !== 'ADMIN') return <Redirect href="/" />

  return <Slot />
}
