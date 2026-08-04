import * as ImagePicker from 'expo-image-picker'
import { Platform } from 'react-native'

import type { PickedAvatar } from '@/features/profile/api/profile.api'

// Wraps expo-image-picker behind the plain { uri, fileName, mimeType } shape
// profile.api.ts's uploadAvatar builds a multipart part from. Native asks
// for media-library permission first (a no-op on web, where the browser's
// own file picker handles access). Returns null on permission-denied or a
// cancelled picker — callers just no-op in that case, no error to surface.
export function useAvatarPicker() {
  const pickAvatar = async (): Promise<PickedAvatar | null> => {
    if (Platform.OS !== 'web') {
      const permission = await ImagePicker.requestMediaLibraryPermissionsAsync()
      if (!permission.granted) return null
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      aspect: [1, 1],
      quality: 0.8,
    })

    if (result.canceled) return null

    const asset = result.assets[0]
    if (!asset) return null

    return {
      uri: asset.uri,
      fileName: asset.fileName ?? `avatar-${Date.now()}.jpg`,
      mimeType: asset.mimeType ?? 'image/jpeg',
    }
  }

  return { pickAvatar }
}
