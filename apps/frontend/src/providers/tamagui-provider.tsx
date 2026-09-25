import { Bangers_400Regular } from '@expo-google-fonts/bangers'
import { DelaGothicOne_400Regular } from '@expo-google-fonts/dela-gothic-one'
import { Fredoka_500Medium, Fredoka_600SemiBold } from '@expo-google-fonts/fredoka'
import { Nunito_500Medium, Nunito_700Bold } from '@expo-google-fonts/nunito'
import { useFonts } from 'expo-font'
import { Platform, useColorScheme } from 'react-native'
import { TamaguiProvider as TamaguiRootProvider } from 'tamagui'

import { useThemeStore } from '@/shared/stores/theme.store'

import tamaguiConfig from '../../tamagui.config'

export function TamaguiProvider({ children }: { children: React.ReactNode }) {
  const scheme = useColorScheme()
  const mode = useThemeStore((state) => state.mode)
  const [fontsLoaded] = useFonts({
    Fredoka_500Medium,
    Fredoka_600SemiBold,
    Nunito_500Medium,
    Nunito_700Bold,
    // The manga JoJo × One Piece cinematics' own display face (owner
    // decision, 2026-09-25 playtest feedback) - see ObsidianVault/
    // game-victory-defeat-cinematic-2026-09-14.md. Bangers for verdict
    // headlines, Dela Gothic One for onomatopoeia/winner names (it also
    // covers katakana, unlike Bangers).
    Bangers_400Regular,
    DelaGothicOne_400Regular,
  })

  // Web fonts ship via the CSS emitted by @tamagui/metro-plugin
  // (tamagui-web.css, imported in app/_layout.tsx) instead of this JS
  // loader, so gating on `fontsLoaded` here only matters on native. On web
  // it never resolves synchronously during expo-router's static SSR pass,
  // which aborts that Suspense boundary and forces a client-only re-render
  // (surfaces as React error #419 / "Switched to client rendering").
  if (Platform.OS !== 'web' && !fontsLoaded) return null

  const resolvedTheme = mode === 'system' ? (scheme ?? 'light') : mode

  return (
    <TamaguiRootProvider config={tamaguiConfig} defaultTheme={resolvedTheme}>
      {children}
    </TamaguiRootProvider>
  )
}
