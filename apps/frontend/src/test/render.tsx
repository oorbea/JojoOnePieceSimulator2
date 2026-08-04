import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, type RenderOptions } from '@testing-library/react-native'
import { useState, type ReactElement, type ReactNode } from 'react'
import { SafeAreaProvider } from 'react-native-safe-area-context'
import { TamaguiProvider } from 'tamagui'

import tamaguiConfig from '../../tamagui.config'

// Wraps a component under test with the real Tamagui config (not a mock —
// tokens, themes and breakpoints all need to resolve for real, since that's
// exactly what this suite is checking) plus the other providers most
// screens/containers assume exist somewhere above them. Mutations/queries
// never retry here: a broken mock should fail the test immediately, not
// after a retry backoff nobody's waiting for.
function AllProviders({ children }: { children: ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: { retry: false, staleTime: 0 },
          mutations: { retry: false },
        },
      })
  )

  return (
    <SafeAreaProvider>
      <TamaguiProvider config={tamaguiConfig} defaultTheme="light">
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
      </TamaguiProvider>
    </SafeAreaProvider>
  )
}

export function renderWithProviders(ui: ReactElement, options?: RenderOptions) {
  return render(ui, { wrapper: AllProviders, ...options })
}

export * from '@testing-library/react-native'
