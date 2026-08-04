import { Home, User } from '@tamagui/lucide-icons-2'
import { usePathname, useRouter } from 'expo-router'

import { AppShell, type AppShellNavItem } from '@/shared/components/presentational/app-shell'
import { useSessionStore } from '@/shared/stores/session.store'
import { useThemeStore } from '@/shared/stores/theme.store'

// `experiments.typedRoutes` is on, so a `href` to anything else is a compile
// error. This array is the single extension point for future nav items —
// widen the union and add an entry here once a new route exists.
const NAV_ITEMS: { href: '/' | '/profile'; label: string; icon: AppShellNavItem['icon'] }[] = [
  { href: '/', label: 'Home', icon: Home },
  { href: '/profile', label: 'Profile', icon: User },
]

export function AppShellContainer({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname()
  const clearSession = useSessionStore((state) => state.clearSession)
  const themeMode = useThemeStore((state) => state.mode)
  const cycleTheme = useThemeStore((state) => state.cycle)

  const items: AppShellNavItem[] = NAV_ITEMS.map((item) => ({
    ...item,
    active: pathname === item.href,
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
