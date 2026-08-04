import { ScrollView, YStack } from 'tamagui'
import { useSafeAreaInsets } from 'react-native-safe-area-context'

import {
  columnMaxWidth,
  topClearance,
  bottomClearance,
  desktopBottomClearance,
} from '@/shared/lib/layout'

import { AquaBackground } from './aqua-background'

type PageShellProps = {
  children: React.ReactNode
  align?: 'center' | 'top'
  maxWidth?: number
  scroll?: boolean
  /** Adds clearance for the floating top/bottom channel bars — use on
   * screens that render inside the authenticated app shell. */
  navPadding?: boolean
  /** Static backdrop only, no animated bubbles. */
  plain?: boolean
  /** Renders its own AquaBackground. Turn off on screens that already sit
   * inside AppShell, which mounts one for every authenticated route — two
   * stacked backdrops otherwise render (and animate) on top of each other. */
  backdrop?: boolean
}

// Kills the 5x duplicated centering-wrapper-over-gradient recipe that used
// to live inline in every screen. Safe-area aware, responsive maxWidth.
export function PageShell({
  children,
  align = 'center',
  maxWidth = 480,
  scroll = false,
  navPadding = false,
  plain = false,
  backdrop = !navPadding,
}: PageShellProps) {
  const insets = useSafeAreaInsets()

  const content = (
    <YStack
      flex={1}
      width="100%"
      items="center"
      justify={align === 'center' ? 'center' : 'flex-start'}
      p="$4"
      pt={topClearance(insets, navPadding)}
      pb={bottomClearance(insets, navPadding)}
      gap="$4"
      $md={{ p: '$6', pb: desktopBottomClearance(insets) }}
    >
      <YStack
        width="100%"
        maxW={maxWidth}
        $md={{ maxW: columnMaxWidth(maxWidth, 'md') }}
        $lg={{ maxW: columnMaxWidth(maxWidth, 'lg') }}
        $xl={{ maxW: columnMaxWidth(maxWidth, 'xl') }}
        items="center"
        gap="$4"
      >
        {children}
      </YStack>
    </YStack>
  )

  return (
    <YStack flex={1}>
      {backdrop ? <AquaBackground plain={plain} /> : null}
      {scroll ? (
        <ScrollView flex={1} contentContainerStyle={{ flexGrow: 1 } as object}>
          {content}
        </ScrollView>
      ) : (
        content
      )}
    </YStack>
  )
}
