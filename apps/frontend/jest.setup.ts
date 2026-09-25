// Minimal mocks shared by both jest.config.js projects (web/native). Kept
// intentionally small — only what's needed for components under test to
// mount without crashing, not a full re-implementation of any of these
// modules.
// @testing-library/react-native >=v9 auto-extends `expect` with its own
// matchers (toBeVisible, toHaveTextContent, ...) on import — no separate
// jest-native setup import needed.

// @react-native-async-storage/async-storage has no native module in either
// jest-expo preset (unlike most Expo-owned modules, this is a community
// package with no built-in test mock) - without this, anything importing it
// transitively (theme.store.ts, language.store.ts, and now interceptors.ts
// via language.store.ts for the Accept-Language header) throws
// "NativeModule: AsyncStorage is null" the moment the module loads, in both
// the jsdom and native projects. The package ships this exact in-memory mock
// for tests.
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
)

// react-native-safe-area-context: fixed, controllable insets. A real device
// would give the top ChannelBar/bottom dock clearance math (see
// src/shared/lib/layout.ts) real safe-area values; tests use zero insets so
// the numbers under test are exactly the layout constants themselves, with
// no platform noise mixed in.
jest.mock('react-native-safe-area-context', () => {
  const insets = { top: 0, right: 0, bottom: 0, left: 0 }
  return {
    SafeAreaProvider: ({ children }: { children: React.ReactNode }) => children,
    useSafeAreaInsets: () => insets,
    useSafeAreaFrame: () => ({ x: 0, y: 0, width: 390, height: 844 }),
  }
})

// Reanimated 4's own `react-native-reanimated/mock` still pulls in
// react-native-worklets' native module init chain (loadUnpackers →
// NativeWorklets), which crashes outside a real native runtime. Hand-rolled
// instead, covering exactly the surface this repo's native-only files
// actually use (bubble-field.native.tsx, use-reduced-motion.ts): worklets
// run synchronously on the JS thread, shared values are plain refs, and
// Animated.* passes straight through to core RN components.
jest.mock('react-native-reanimated', () => {
  const React = require('react')
  const RN = require('react-native')

  function useSharedValue(initial: number) {
    const ref = React.useRef({ value: initial })
    return ref.current
  }

  function useAnimatedStyle(factory: () => Record<string, unknown>) {
    return factory()
  }

  const withTiming = (toValue: number) => toValue
  const withRepeat = (animation: number) => animation
  const withDelay = (_delay: number, animation: number) => animation
  // withSequence: real reanimated animates through every value in turn;
  // mocked the same way withTiming/withRepeat are - worklets run
  // synchronously, so a shared value just needs to end up at *a* concrete
  // number, not actually tween through the sequence. Last value wins,
  // matching where the animation would settle once real timers fired.
  const withSequence = (...animations: number[]) => animations[animations.length - 1]
  // Easing: only the identity fns components actually call are needed here
  // (see this mock's own doc) - modifiers just return their input curve
  // unevaluated, since withTiming/withSequence above never actually invoke
  // the easing function, only pass it through.
  const identity = (curve: unknown) => curve
  const easing = {
    linear: (t: number) => t,
    quad: (t: number) => t,
    cubic: (t: number) => t,
    ease: (t: number) => t,
    out: identity,
    in: identity,
    inOut: identity,
    back: (_s?: number) => (t: number) => t,
    bezier: () => (t: number) => t,
  }
  const interpolate = (value: number, input: number[], output: number[]) => {
    const [inMin, inMax] = [input[0], input[input.length - 1]]
    const [outMin, outMax] = [output[0], output[output.length - 1]]
    if (inMax === inMin) return outMin
    const t = (value - inMin) / (inMax - inMin)
    return outMin + t * (outMax - outMin)
  }

  return {
    __esModule: true,
    default: {
      View: RN.View,
      Text: RN.Text,
      ScrollView: RN.ScrollView,
      Image: RN.Image,
      createAnimatedComponent: (Component: unknown) => Component,
    },
    Easing: easing,
    useSharedValue,
    useAnimatedStyle,
    useReducedMotion: () => false,
    withTiming,
    withRepeat,
    withDelay,
    withSequence,
    interpolate,
    runOnJS: (fn: (...args: unknown[]) => unknown) => fn,
    cancelAnimation: () => undefined,
  }
})

jest.mock('@tamagui/linear-gradient', () => {
  const { View } = require('react-native')
  return { LinearGradient: View }
})

// jest-expo doesn't mock expo-image (it's a real native module, not part of
// the Expo SDK preset it ships a mock for). Map to RN's Image, forwarding
// only the props lazy-image.test.tsx/skeleton.test.tsx actually assert on -
// the rest (contentFit, transition, recyclingKey, placeholder) are real
// expo-image props RN's Image doesn't understand and would warn about.
jest.mock('expo-image', () => {
  const React = require('react')
  const { Image } = require('react-native')
  const MockImage = React.forwardRef((props: any, ref: any) =>
    React.createElement(Image, {
      ref,
      source: props.source,
      style: props.style,
      onLoad: props.onLoad,
      onError: props.onError,
      accessibilityLabel: props.accessibilityLabel,
      testID: props.testID,
    })
  )
  MockImage.displayName = 'MockExpoImage'
  return { Image: MockImage }
})

jest.mock('expo-image-picker', () => ({
  requestMediaLibraryPermissionsAsync: jest.fn().mockResolvedValue({ granted: true }),
  launchImageLibraryAsync: jest.fn().mockResolvedValue({ canceled: true, assets: null }),
  MediaTypeOptions: { Images: 'Images' },
}))

// jsdom (the web project's testEnvironment) has no matchMedia — Tamagui's
// web build (@tamagui/select, and anything reading window size) calls it
// eagerly at module-load time, before any component even mounts.
if (typeof window !== 'undefined' && !window.matchMedia) {
  window.matchMedia = jest.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  }))
}

jest.mock('expo-secure-store', () => ({
  getItemAsync: jest.fn().mockResolvedValue(null),
  setItemAsync: jest.fn().mockResolvedValue(undefined),
  deleteItemAsync: jest.fn().mockResolvedValue(undefined),
}))

// expo-localization pulls in expo-modules-core's native Platform.select
// bridge, which only exists under jest-expo's native preset - the `logic`
// project (jest-expo/web, jsdom) has no native module registry at all, so
// anything importing shared/i18n (directly or via power-translations.ts /
// language.store.ts) throws "Cannot read properties of undefined (reading
// 'select')" the moment the module loads, in both projects alike.
jest.mock('expo-localization', () => ({
  getLocales: () => [{ languageTag: 'en-GB', languageCode: 'en' }],
}))

