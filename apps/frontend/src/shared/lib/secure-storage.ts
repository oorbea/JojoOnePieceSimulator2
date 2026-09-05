import { Platform } from 'react-native'
import * as SecureStore from 'expo-secure-store'

// Native-only. Web must never persist auth state here (the refresh token
// lives in an HttpOnly cookie on web, not client-readable storage) - a
// caller that reaches this on web is a bug, so it throws loudly instead of
// silently falling back to localStorage (which is what happened before this
// module absorbed the platform branch: see session.store.ts's history).
function assertNative(): void {
  if (Platform.OS === 'web') {
    throw new Error('secureStorage is native-only; web must not persist auth state')
  }
}

export const secureStorage = {
  async getItem(key: string): Promise<string | null> {
    assertNative()
    return SecureStore.getItemAsync(key)
  },

  async setItem(key: string, value: string): Promise<void> {
    assertNative()
    // Never included in an iCloud/device-transfer backup - appropriate for a
    // credential like the refresh token, unlike the default WHEN_UNLOCKED.
    await SecureStore.setItemAsync(key, value, {
      keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
    })
  },

  async removeItem(key: string): Promise<void> {
    assertNative()
    await SecureStore.deleteItemAsync(key)
  },
}
