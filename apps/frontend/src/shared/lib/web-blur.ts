import { Platform, type ViewStyle } from 'react-native'

// Real backdrop-filter blur, web-only — RN has no native equivalent short of
// adding `expo-blur`, which this pass deliberately avoids. Computed once at
// module load (not per-render) and applied via a plain RN `style=` prop,
// never as a Tamagui style prop — the Tamagui optimizing compiler tries to
// statically extract known style props, and `backdropFilter` isn't one, so
// routing it through `style=` keeps it out of that pass entirely.
//
// Native has no blur at all, so callers using this constant should also
// resolve their glass fill token to the *Native variant (e.g. `$glassFill`
// on web, `$glassFillNative` on native) to compensate with opacity instead.
export const WEB_BLUR_STYLE: ViewStyle | undefined =
  Platform.OS === 'web'
    ? ({
        backdropFilter: 'blur(12px)',
        WebkitBackdropFilter: 'blur(12px)',
      } as unknown as ViewStyle)
    : undefined

export const isWeb = Platform.OS === 'web'
