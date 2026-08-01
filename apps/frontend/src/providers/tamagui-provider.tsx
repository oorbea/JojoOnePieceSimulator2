import { useFonts } from 'expo-font'
import { Platform, useColorScheme } from 'react-native'
import { TamaguiProvider as TamaguiRootProvider } from 'tamagui'

import tamaguiConfig from '../../tamagui.config'

export function TamaguiProvider({ children }: { children: React.ReactNode }) {
  const scheme = useColorScheme()
  const [fontsLoaded] = useFonts({
    Inter: require('@tamagui/font-inter/otf/Inter-Medium.otf'),
    InterBold: require('@tamagui/font-inter/otf/Inter-Bold.otf'),
  })

  // Web fonts ship via the CSS emitted by @tamagui/metro-plugin
  // (tamagui-web.css, imported in app/_layout.tsx) instead of this JS
  // loader, so gating on `fontsLoaded` here only matters on native. On web
  // it never resolves synchronously during expo-router's static SSR pass,
  // which aborts that Suspense boundary and forces a client-only re-render
  // (surfaces as React error #419 / "Switched to client rendering").
  if (Platform.OS !== 'web' && !fontsLoaded) return null

  return (
    <TamaguiRootProvider config={tamaguiConfig} defaultTheme={scheme ?? 'light'}>
      {children}
    </TamaguiRootProvider>
  )
}
