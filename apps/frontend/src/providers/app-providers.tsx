import { I18nextProvider } from 'react-i18next'
import { Platform } from 'react-native'
import { SafeAreaProvider } from 'react-native-safe-area-context'

import { PictureEventsBridge } from '@/providers/picture-events-bridge'
import { QueryProvider } from '@/providers/query-provider'
import { TamaguiProvider } from '@/providers/tamagui-provider'
import { ToasterMount } from '@/providers/toaster-mount'
import { ErrorBoundary } from '@/shared/components/containers/error-boundary'
import i18n from '@/shared/i18n'
import { configureImageQueue } from '@/shared/lib/image-queue'

// Native's own loader/decoder does extra work per concurrent fetch on top
// of the network, and mobile radios fare worse under parallelism than web -
// so native gets a slightly lower cap than web. See image-queue.ts.
configureImageQueue({ maxConcurrent: Platform.OS === 'web' ? 4 : 3 })

// Single composition point so app/_layout.tsx stays a thin route shell —
// add new app-wide providers here, not in the layout file.
export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <I18nextProvider i18n={i18n}>
      <SafeAreaProvider>
        <TamaguiProvider>
          <QueryProvider>
            <ErrorBoundary>{children}</ErrorBoundary>
            <ToasterMount />
            <PictureEventsBridge />
          </QueryProvider>
        </TamaguiProvider>
      </SafeAreaProvider>
    </I18nextProvider>
  )
}
