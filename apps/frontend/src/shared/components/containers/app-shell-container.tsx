import { Home } from '@tamagui/lucide-icons-2'
import { usePathname, useRouter } from 'expo-router'

import { AppShell, type AppShellNavItem } from '@/shared/components/presentational/app-shell'
import { useSessionStore } from '@/shared/stores/session.store'
import { useThemeStore } from '@/shared/stores/theme.store'

// `experiments.typedRoutes` is on and the only route today is `/`, so a
// `href` to anything else is a compile error. This array is the single
// extension point for future nav items — add a route here once it exists.
const NAV_ITEMS: { href: '/'; label: string; icon: AppShellNavItem['icon'] }[] = [
  { href: '/', label: 'Home', icon: Home },
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
