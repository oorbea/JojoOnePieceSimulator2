import { ScrollView, YStack } from 'tamagui'
import { useSafeAreaInsets } from 'react-native-safe-area-context'

import { columnMaxWidth } from '@/shared/lib/layout'
import { useNavInsets } from '@/shared/lib/nav-insets'

import { AquaBackground } from './aqua-background'

type PageShellProps = {
  children: React.ReactNode
  align?: 'center' | 'top'
  maxWidth?: number
  scroll?: boolean
  /** Static backdrop only, no animated bubbles. */
  plain?: boolean
  /** Renders its own AquaBackground. Defaults to on outside AppShell (login,
   * +not-found, LoadingScreen) and off inside it — AppShell already mounts
   * one for every authenticated route, so two stacked backdrops would
   * otherwise render (and animate) on top of each other. Screens no longer
   * need to opt into this by hand: it's derived from whether NavInsetsProvider
   * is actually reserving space above this page. */
  backdrop?: boolean
}

// Kills the 5x duplicated centering-wrapper-over-gradient recipe that used
// to live inline in every screen. Safe-area aware, responsive maxWidth.
//
// Nav clearance is no longer a prop a screen has to remember to pass: it's
// read from NavInsetsProvider (see nav-insets.tsx), which AppShell keeps in
// sync with the floating bars' real measured height. Outside the shell the
// context defaults to zero, so this falls back to plain breathing room.
export function PageShell({
  children,
  align = 'center',
  maxWidth = 480,
  scroll = false,
  plain = false,
  backdrop,
}: PageShellProps) {
  const insets = useSafeAreaInsets()
  const navInsets = useNavInsets()
  const insideShell = navInsets.top > 0 || navInsets.bottom > 0
  const showBackdrop = backdrop ?? !insideShell

  const content = (
    <YStack
      flex={1}
      width="100%"
      items="center"
      justify={align === 'center' ? 'center' : 'flex-start'}
      px="$4"
      pt={insideShell ? navInsets.top : insets.top + 16}
      pb={insideShell ? navInsets.bottom : insets.bottom + 16}
      gap="$4"
      $md={{ px: '$6' }}
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
      {showBackdrop ? <AquaBackground plain={plain} /> : null}
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
