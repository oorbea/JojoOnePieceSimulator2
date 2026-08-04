import { ScrollView, YStack } from 'tamagui'
import { useSafeAreaInsets } from 'react-native-safe-area-context'

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
}: PageShellProps) {
  const insets = useSafeAreaInsets()

  const content = (
    <YStack
      flex={1}
      width="100%"
      items="center"
      justify={align === 'center' ? 'center' : 'flex-start'}
      p="$4"
      pt={navPadding ? insets.top + 88 : insets.top + 16}
      pb={navPadding ? insets.bottom + 96 : insets.bottom + 16}
      gap="$4"
      $md={{ p: '$6' }}
    >
      <YStack width="100%" maxW={maxWidth} $md={{ maxW: maxWidth * 1.18 }} items="center" gap="$4">
        {children}
      </YStack>
    </YStack>
  )

  return (
    <YStack flex={1}>
      <AquaBackground plain={plain} />
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
