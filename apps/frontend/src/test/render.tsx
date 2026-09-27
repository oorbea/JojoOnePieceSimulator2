import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, type RenderOptions } from '@testing-library/react-native'
import i18next from 'i18next'
import { useState, type ReactElement, type ReactNode } from 'react'
import { I18nextProvider, initReactI18next } from 'react-i18next'
import { SafeAreaProvider } from 'react-native-safe-area-context'
import { TamaguiProvider } from 'tamagui'

import enGB from '@/shared/i18n/locales/en-GB.json'

import tamaguiConfig from '../../tamagui.config'

// A dedicated instance, deliberately NOT the app's shared `i18n` singleton
// from shared/i18n - that one initializes at DEFAULT_LOCALE, which is a
// product decision (es-ES, as of 2026-09-27) independent of what language
// this test suite's hundreds of English-language assertions
// (getByLabelText('Cancel'), etc.) are written against. Hard-coding en-GB
// here keeps every test's expected strings correct regardless of
// DEFAULT_LOCALE - see i18n-keys.test.ts, which likewise treats en-GB.json
// as the reference catalog for its own, unrelated reason.
// eslint-disable-next-line import/no-named-as-default-member -- default import is correct; i18next's named `createInstance`/`use` exports are unrelated (same reasoning as shared/i18n/index.ts's own instance)
const testI18n = i18next.createInstance()
void testI18n.use(initReactI18next).init({
  resources: { 'en-GB': { translation: enGB } },
  lng: 'en-GB',
  fallbackLng: 'en-GB',
  interpolation: { escapeValue: false },
  returnNull: false,
})

// Wraps a component under test with the real Tamagui config (not a mock —
// tokens, themes and breakpoints all need to resolve for real, since that's
// exactly what this suite is checking) plus the other providers most
// screens/containers assume exist somewhere above them. Mutations/queries
// never retry here: a broken mock should fail the test immediately, not
// after a retry backoff nobody's waiting for.
// I18nextProvider with testI18n (not shared/i18n's own singleton) makes
// every test independent of import order AND of DEFAULT_LOCALE - any
// component using useTranslation() gets the real en-GB catalog whether or
// not it happens to import shared/i18n itself.
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
    <I18nextProvider i18n={testI18n}>
      <SafeAreaProvider>
        <TamaguiProvider config={tamaguiConfig} defaultTheme="light">
          <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
        </TamaguiProvider>
      </SafeAreaProvider>
    </I18nextProvider>
  )
}

// `render()` from this version of @testing-library/react-native is async —
// always `await renderWithProviders(...)`. Skipping the await doesn't throw;
// it just means `screen` hasn't registered the result yet, so every query
// right after fails with "render function has not been called" even though
// the component renders fine a tick later.
export function renderWithProviders(ui: ReactElement, options?: RenderOptions) {
  return render(ui, { wrapper: AllProviders, ...options })
}

export * from '@testing-library/react-native'
