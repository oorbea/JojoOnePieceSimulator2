import { Contrast, Moon, Sun } from '@tamagui/lucide-icons-2'

import { a11yProps } from '@/shared/lib/a11y'
import type { ThemeMode } from '@/shared/stores/theme.store'

import { ChannelBarItem } from './channel-bar'

const ICON: Record<ThemeMode, typeof Moon> = {
  system: Contrast,
  light: Sun,
  dark: Moon,
}

const LABEL: Record<ThemeMode, string> = {
  system: 'Match system theme',
  light: 'Light theme',
  dark: 'Dark theme',
}

type ThemeToggleProps = {
  mode: ThemeMode
  onCycle: () => void
}

// Cycles system -> light -> dark -> system. Sits in the channel bar next to
// the other nav items, styled identically so it doesn't read as a special
// control bolted on.
export function ThemeToggle({ mode, onCycle }: ThemeToggleProps) {
  const Icon = ICON[mode]

  return (
    <ChannelBarItem
      iconOnly
      onPress={onCycle}
      pressStyle={{ scale: 0.9, rotate: '-12deg' }}
      {...a11yProps(LABEL[mode], 'button')}
      hitSlop={8}
    >
      <Icon size={22} color="$panelText" strokeWidth={2.5} />
    </ChannelBarItem>
  )
}
