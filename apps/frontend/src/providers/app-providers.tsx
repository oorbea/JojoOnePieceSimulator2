import { SafeAreaProvider } from 'react-native-safe-area-context'

import { QueryProvider } from '@/providers/query-provider'
import { TamaguiProvider } from '@/providers/tamagui-provider'
import { ToasterMount } from '@/providers/toaster-mount'
import { ErrorBoundary } from '@/shared/components/containers/error-boundary'

// Single composition point so app/_layout.tsx stays a thin route shell —
// add new app-wide providers here, not in the layout file.
export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <SafeAreaProvider>
      <TamaguiProvider>
        <QueryProvider>
          <ErrorBoundary>{children}</ErrorBoundary>
          <ToasterMount />
        </QueryProvider>
      </TamaguiProvider>
    </SafeAreaProvider>
  )
}
