import { LogOut } from '@tamagui/lucide-icons-2'
import { Image } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { XStack, YStack } from 'tamagui'

import { logoAsset } from '@/shared/assets'
import { a11yProps } from '@/shared/lib/a11y'
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

// Two floating glass pills over the animated sky: a top bar always, a
// bottom dock on mobile only. Mounted once inside the auth guard in
// app/(app)/_layout.tsx, so every authenticated route inherits it.
export function AppShell({
  children,
  items,
  onNavigate,
  onLogout,
  themeMode,
  onCycleTheme,
}: AppShellProps) {
  const insets = useSafeAreaInsets()

  return (
    <YStack flex={1}>
      <AquaBackground />

      <ChannelBar
        dock="top"
        t={insets.top + 8}
        mx="$3"
        transition="bouncy"
        enterStyle={{ y: -24, opacity: 0 }}
      >
        <WiiCard aspect="square" width={44} height={44} interactive={false}>
          <Image
            source={logoAsset}
            style={{ width: '100%', height: '100%' }}
            resizeMode="contain"
          />
        </WiiCard>
        <GlowText level="heading" fontSize="$5">
          JOPS
        </GlowText>

        <XStack flex={1} />

        <XStack display="none" $md={{ display: 'flex' }} gap="$2">
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

        <ThemeToggle mode={themeMode} onCycle={onCycleTheme} />

        <ChannelBarItem
          iconOnly
          onPress={onLogout}
          {...a11yProps('Log out', 'button')}
          hitSlop={8}
        >
          <LogOut size={20} color="$panelText" strokeWidth={2.5} />
        </ChannelBarItem>
      </ChannelBar>

      <YStack flex={1}>{children}</YStack>

      <ChannelBar
        dock="bottom"
        b={insets.bottom + 8}
        mx="$3"
        display="flex"
        $md={{ display: 'none' }}
        transition="bouncy"
        enterStyle={{ y: 24, opacity: 0 }}
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
    </YStack>
  )
}
