import { Platform } from 'react-native'
import * as SecureStore from 'expo-secure-store'

// expo-secure-store throws on web (no Keychain/Keystore there). This module
// is the single place that branches on platform so every consumer — the
// session store especially — can stay platform-agnostic.
export const secureStorage = {
  async getItem(key: string): Promise<string | null> {
    if (Platform.OS === 'web') {
      return typeof window === 'undefined' ? null : window.localStorage.getItem(key)
    }
    return SecureStore.getItemAsync(key)
  },

  async setItem(key: string, value: string): Promise<void> {
    if (Platform.OS === 'web') {
      window.localStorage.setItem(key, value)
      return
    }
    await SecureStore.setItemAsync(key, value)
  },

  async removeItem(key: string): Promise<void> {
    if (Platform.OS === 'web') {
      window.localStorage.removeItem(key)
      return
    }
    await SecureStore.deleteItemAsync(key)
  },
}
