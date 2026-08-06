import * as ImagePicker from 'expo-image-picker'
import { Platform } from 'react-native'

export type PickedPicture = {
  uri: string
  fileName: string
  mimeType: string
}

// Generalized version of features/profile/hooks/use-avatar-picker.ts for any
// Power picture (Stand/DevilFruit) — same { uri, fileName, mimeType } shape
// the picture-upload API functions build a multipart part from, but without
// the 1:1 aspect lock: avatars are cropped to a circle, Power art isn't.
export function usePicturePicker() {
  const pickPicture = async (): Promise<PickedPicture | null> => {
    if (Platform.OS !== 'web') {
      const permission = await ImagePicker.requestMediaLibraryPermissionsAsync()
      if (!permission.granted) return null
    }

    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ['images'],
      allowsEditing: true,
      quality: 0.8,
    })

    if (result.canceled) return null

    const asset = result.assets[0]
    if (!asset) return null

    return {
      uri: asset.uri,
      fileName: asset.fileName ?? `picture-${Date.now()}.jpg`,
      mimeType: asset.mimeType ?? 'image/jpeg',
    }
  }

  return { pickPicture }
}
