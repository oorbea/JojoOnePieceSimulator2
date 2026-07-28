import { SafeAreaProvider } from 'react-native-safe-area-context'

import { QueryProvider } from '@/providers/query-provider'
import { TamaguiProvider } from '@/providers/tamagui-provider'

// Single composition point so app/_layout.tsx stays a thin route shell —
// add new app-wide providers here, not in the layout file.
export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <SafeAreaProvider>
      <TamaguiProvider>
        <QueryProvider>{children}</QueryProvider>
      </TamaguiProvider>
    </SafeAreaProvider>
  )
}
