import { secureStorage } from '../secure-storage'

// This suite runs under the "logic" jest project (jest-expo/web, jsdom),
// where `react-native` is mapped to `react-native-web` and Platform.OS is
// 'web' - exactly the branch this file exists to pin: web must never fall
// back to persisting auth state (it used to write to localStorage before
// the refresh-token rework), it must throw loudly instead. The native
// branch is exercised indirectly by every other suite that touches this
// module, backed by jest.setup.ts's global `expo-secure-store` mock.
describe('secureStorage on web', () => {
  it('getItem throws', async () => {
    await expect(secureStorage.getItem('jops.refresh')).rejects.toThrow(
      'secureStorage is native-only; web must not persist auth state'
    )
  })

  it('setItem throws', async () => {
    await expect(secureStorage.setItem('jops.refresh', 'value')).rejects.toThrow(
      'secureStorage is native-only; web must not persist auth state'
    )
  })

  it('removeItem throws', async () => {
    await expect(secureStorage.removeItem('jops.refresh')).rejects.toThrow(
      'secureStorage is native-only; web must not persist auth state'
    )
  })
})
