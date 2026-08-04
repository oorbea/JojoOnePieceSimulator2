// Minimal mocks shared by both jest.config.js projects (web/native). Kept
// intentionally small — only what's needed for components under test to
// mount without crashing, not a full re-implementation of any of these
// modules.
// @testing-library/react-native >=v9 auto-extends `expect` with its own
// matchers (toBeVisible, toHaveTextContent, ...) on import — no separate
// jest-native setup import needed.

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

jest.mock('react-native-reanimated', () => {
  const Reanimated = require('react-native-reanimated/mock')
  Reanimated.default.call = () => undefined
  return Reanimated
})

jest.mock('@tamagui/linear-gradient', () => {
  const { View } = require('react-native')
  return { LinearGradient: View }
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
