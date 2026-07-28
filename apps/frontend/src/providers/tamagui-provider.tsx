import { useFonts } from 'expo-font'
import { useColorScheme } from 'react-native'
import { TamaguiProvider as TamaguiRootProvider } from 'tamagui'

import tamaguiConfig from '../../tamagui.config'

export function TamaguiProvider({ children }: { children: React.ReactNode }) {
  const scheme = useColorScheme()
  const [fontsLoaded] = useFonts({
    Inter: require('@tamagui/font-inter/otf/Inter-Medium.otf'),
    InterBold: require('@tamagui/font-inter/otf/Inter-Bold.otf'),
  })

  if (!fontsLoaded) return null

  return (
    <TamaguiRootProvider config={tamaguiConfig} defaultTheme={scheme ?? 'light'}>
      {children}
    </TamaguiRootProvider>
  )
}
