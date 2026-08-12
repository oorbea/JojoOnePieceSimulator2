/** @type {import('jest').Config} */
module.exports = {
  projects: [
    {
      // Pure algorithm / static-analysis checks that don't render anything
      // (tamagui token ordering, layout math, the copy dash-guard). jsdom is
      // enough for these and lets tamagui.config.ts's real `createTamagui()`
      // call resolve without the native module stack.
      displayName: 'logic',
      preset: 'jest-expo/web',
      setupFiles: ['<rootDir>/jest.setup.ts'],
      testEnvironment: 'jsdom',
      moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
      },
      testMatch: [
        '<rootDir>/src/shared/lib/__tests__/**/*.test.ts',
        '<rootDir>/src/test/__tests__/**/*.test.ts',
        '<rootDir>/src/features/**/lib/__tests__/**/*.test.ts',
        '<rootDir>/src/features/**/stores/__tests__/**/*.test.ts',
      ],
      // jest-expo's own default only whitelists RN/Expo packages for
      // transform, keyed on the FIRST "node_modules/" segment in a path.
      // Under pnpm every package is nested two levels deep
      // (node_modules/.pnpm/<pkg>/node_modules/<realpkg>/...), so that first
      // segment is always ".pnpm" — the whitelist regex's negative lookahead
      // matches immediately there and the real package (e.g. @tamagui's
      // published ESM) never reaches transform, failing on its `import`
      // statements. Transforming everything is the simplest fix that works
      // regardless of package manager layout.
      transformIgnorePatterns: [],
      // The preset's own transform only matches `.[jt]sx?` — @tamagui ships
      // real `.mjs` files, which that pattern silently skips regardless of
      // transformIgnorePatterns above, hitting the same "Cannot use import
      // statement outside a module" error. Extend the pattern rather than
      // replace the preset's transform outright.
      transform: {
        '\\.[jt]sx?$': 'babel-jest',
        '\\.mjs$': 'babel-jest',
      },
    },
    {
      // Component/behavior tests: rendered through react-test-renderer via
      // jest-expo's default (native) preset, same as Expo Go/native builds
      // — accessibilityLabel/accessibilityRole are real props here, not
      // translated to aria-* the way a11yProps does for Platform.OS==='web'
      // (see src/shared/lib/a11y.ts), which keeps queries in these tests
      // simple and platform-consistent.
      displayName: 'native',
      preset: 'jest-expo',
      setupFiles: ['<rootDir>/jest.setup.ts'],
      moduleNameMapper: {
        '^@/(.*)$': '<rootDir>/src/$1',
      },
      testPathIgnorePatterns: [
        '/node_modules/',
        '<rootDir>/src/shared/lib/__tests__/',
        '<rootDir>/src/test/__tests__/',
        '<rootDir>/src/features/.*/lib/__tests__/',
        '<rootDir>/src/features/.*/stores/__tests__/',
      ],
      transformIgnorePatterns: [],
      transform: {
        '\\.[jt]sx?$': 'babel-jest',
        '\\.mjs$': 'babel-jest',
      },
    },
  ],
}
