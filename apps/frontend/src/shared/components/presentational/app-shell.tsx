import { LogOut } from '@tamagui/lucide-icons-2'
import { useState } from 'react'
import type { LayoutChangeEvent } from 'react-native'
import { Image } from 'react-native'
import { useTranslation } from 'react-i18next'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { useMedia, XStack, YStack } from 'tamagui'

import { logoAsset } from '@/shared/assets'
import { a11yProps } from '@/shared/lib/a11y'
import { DOCK_BAR_HEIGHT, NAV_BAR_HEIGHT, navBottomInset, navTopInset } from '@/shared/lib/layout'
import { NavInsetsProvider } from '@/shared/lib/nav-insets'
import type { ThemeMode } from '@/shared/stores/theme.store'

import { AquaBackground } from './aqua-background'
import { ChannelBar, ChannelBarItem } from './channel-bar'
import type { IconComponent } from './channel-tile'
import { GlowText } from './glow-text'
import { ThemeToggle } from './theme-toggle'
import { WiiCard } from './wii-card'

export type AppShellNavItem = {
  href: string
  label: string
  icon: IconComponent
  active: boolean
}

type AppShellProps = {
  children: React.ReactNode
  items: AppShellNavItem[]
  onNavigate: (href: string) => void
  onLogout: () => void
  themeMode: ThemeMode
  onCycleTheme: () => void
}

// Two floating glass pills over the animated sky: a top bar with nav links
// on wide screens, OR a bottom dock on narrow ones — never both, never
// neither, decided by a single boolean below. Mounted once inside the auth
// guard in app/(app)/_layout.tsx, so every authenticated route inherits it.
//
// AppShell measures both bars' real rendered height via onLayout and
// publishes the resulting reservation through NavInsetsProvider, so
// PageShell's clearance always matches what's actually on screen — a bar
// that wraps onto two rows never ends up covering page content.
export function AppShell({
  children,
  items,
  onNavigate,
  onLogout,
  themeMode,
  onCycleTheme,
}: AppShellProps) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()
  const media = useMedia()
  const showTopLinks = media.md

  const [topBarHeight, setTopBarHeight] = useState<number | null>(null)
  const [dockHeight, setDockHeight] = useState<number | null>(null)

  const onTopBarLayout = (e: LayoutChangeEvent) => setTopBarHeight(e.nativeEvent.layout.height)
  const onDockLayout = (e: LayoutChangeEvent) => setDockHeight(e.nativeEvent.layout.height)

  const navInsets = {
    top: navTopInset(insets, topBarHeight ?? NAV_BAR_HEIGHT),
    bottom: navBottomInset(insets, showTopLinks ? null : dockHeight ?? DOCK_BAR_HEIGHT),
  }

  return (
    <YStack flex={1}>
      <AquaBackground />

      <ChannelBar
        dock="top"
        t={insets.top + 8}
        mx="$3"
        transition="bouncy"
        enterStyle={{ y: -24, opacity: 0 }}
        onLayout={onTopBarLayout}
      >
        <WiiCard aspect="square" width={44} height={44} interactive={false}>
          <Image
            source={logoAsset}
            style={{ width: '100%', height: '100%' }}
            resizeMode="contain"
          />
        </WiiCard>
        <GlowText level="heading" fontSize="$5" numberOfLines={1} display="none" $sm={{ display: 'flex' }}>
          JOPS
        </GlowText>

        <XStack flex={1} />

        {showTopLinks ? (
          <XStack gap="$2">
            {items.map((item) => (
              <ChannelBarItem
                key={item.href}
                active={item.active}
                onPress={() => onNavigate(item.href)}
                {...a11yProps(item.label, 'button')}
              >
                <item.icon size={20} color={item.active ? 'white' : '$panelText'} strokeWidth={2.5} />
                <GlowText level="label" tone={item.active ? 'onColor' : 'ink'} fontSize="$2">
                  {item.label}
                </GlowText>
              </ChannelBarItem>
            ))}
          </XStack>
        ) : null}

        <ThemeToggle mode={themeMode} onCycle={onCycleTheme} />

        <ChannelBarItem
          iconOnly
          onPress={onLogout}
          {...a11yProps(t('nav.logOut'), 'button')}
          hitSlop={8}
        >
          <LogOut size={20} color="$panelText" strokeWidth={2.5} />
        </ChannelBarItem>
      </ChannelBar>

      <NavInsetsProvider value={navInsets}>
        <YStack flex={1}>{children}</YStack>
      </NavInsetsProvider>

      {!showTopLinks ? (
        <ChannelBar
          dock="bottom"
          b={insets.bottom + 8}
          mx="$3"
          transition="bouncy"
          enterStyle={{ y: 24, opacity: 0 }}
          onLayout={onDockLayout}
        >
          {items.map((item) => (
            <YStack key={item.href} flex={1} items="center">
              <ChannelBarItem
                active={item.active}
                iconOnly
                onPress={() => onNavigate(item.href)}
                {...a11yProps(item.label, 'button')}
              >
                <item.icon size={22} color={item.active ? 'white' : '$panelText'} strokeWidth={2.5} />
              </ChannelBarItem>
            </YStack>
          ))}
        </ChannelBar>
      ) : null}
    </YStack>
  )
}
