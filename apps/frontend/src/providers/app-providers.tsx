import { I18nextProvider } from 'react-i18next'
import { SafeAreaProvider } from 'react-native-safe-area-context'

import { QueryProvider } from '@/providers/query-provider'
import { TamaguiProvider } from '@/providers/tamagui-provider'
import { ToasterMount } from '@/providers/toaster-mount'
import { ErrorBoundary } from '@/shared/components/containers/error-boundary'
import i18n from '@/shared/i18n'

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
          </QueryProvider>
        </TamaguiProvider>
      </SafeAreaProvider>
    </I18nextProvider>
  )
}
