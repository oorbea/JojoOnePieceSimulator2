import { Home, Shield, User } from '@tamagui/lucide-icons-2'
import { usePathname, useRouter } from 'expo-router'

import { AppShell, type AppShellNavItem } from '@/shared/components/presentational/app-shell'
import { useSessionStore } from '@/shared/stores/session.store'
import { useThemeStore } from '@/shared/stores/theme.store'

// `experiments.typedRoutes` is on, so a `href` to anything else is a compile
// error. This array is the single extension point for future nav items —
// widen the union and add an entry here once a new route exists.
const NAV_ITEMS: { href: '/' | '/profile' | '/admin'; label: string; icon: AppShellNavItem['icon'] }[] = [
  { href: '/', label: 'Home', icon: Home },
  { href: '/profile', label: 'Profile', icon: User },
]

const ADMIN_NAV_ITEM = { href: '/admin' as const, label: 'Admin', icon: Shield }

export function AppShellContainer({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const clearSession = useSessionStore((state) => state.clearSession)
  const session = useSessionStore((state) => state.session)
  const themeMode = useThemeStore((state) => state.mode)
  const cycleTheme = useThemeStore((state) => state.cycle)

  // Admin nav item is hidden entirely for non-admins — the actual gate lives
  // server-side (RequireAdmin) plus the /admin route group's own guard, this
  // is purely about not advertising a channel a REGULAR user can't use.
  const navSource = session?.user.role === 'ADMIN' ? [...NAV_ITEMS, ADMIN_NAV_ITEM] : NAV_ITEMS

  const items: AppShellNavItem[] = navSource.map((item) => ({
    ...item,
    active: pathname === item.href || (item.href === '/admin' && pathname.startsWith('/admin')),
  }))

  return (
    <AppShell
      items={items}
      onNavigate={(href) => router.navigate(href as never)}
      onLogout={() => void clearSession()}
      themeMode={themeMode}
      onCycleTheme={() => void cycleTheme()}
    >
      {children}
    </AppShell>
  )
}
