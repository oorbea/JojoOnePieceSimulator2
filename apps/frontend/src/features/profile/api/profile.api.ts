import { Platform } from 'react-native'

import { apiClient } from '@/shared/api/client'
import type { ProfileUser } from '@/features/profile/types/profile.types'
import type { Locale } from '@/shared/contracts/enums'

export async function getMe(): Promise<ProfileUser> {
  const response = await apiClient.get<ProfileUser>('/users/me')
  return response.data
}

export async function updateUsername(username: string): Promise<ProfileUser> {
  const response = await apiClient.patch<ProfileUser>('/users/me', { username })
  return response.data
}

// language is optional in the backend's PATCH /users/me body, but username
// is always mandatory there (see dto.UpdateProfileRequest), so a language-only
// change still has to resend the caller's current username.
export async function updateLanguage(username: string, language: Locale): Promise<ProfileUser> {
  const response = await apiClient.patch<ProfileUser>('/users/me', { username, language })
  return response.data
}

// Picked image asset, platform-agnostic enough to build a FormData part from
// either an expo-image-picker native asset or a web File.
export type PickedAvatar = {
  uri: string
  fileName: string
  mimeType: string
}

export async function uploadAvatar(asset: PickedAvatar): Promise<ProfileUser> {
  const form = new FormData()

  if (Platform.OS === 'web') {
    // On web, react-native-web's FormData IS the browser's native FormData,
    // which only accepts a real Blob/File - handing it the { uri, name,
    // type } object (the shape React Native's bridge needs) silently
    // stringifies to "[object Object]" instead of a file part, so the
    // backend's r.FormFile("picture") finds nothing and 400s. asset.uri is a
    // blob:/data: URL here, so fetch it back into a real Blob first.
    const blob = await (await fetch(asset.uri)).blob()
    form.append('picture', blob, asset.fileName)
  } else {
    // React Native's FormData accepts this { uri, name, type } shape
    // directly - the native bridge reads the file from the local uri.
    form.append('picture', {
      uri: asset.uri,
      name: asset.fileName,
      type: asset.mimeType,
    } as unknown as Blob)
  }

  // apiClient defaults to Content-Type: application/json for every request
  // (shared/api/client.ts). Axios's default transformRequest checks that
  // resolved header, and when it sees "application/json" it JSON.stringifies
  // FormData bodies instead of sending them as multipart - so a FormData
  // upload silently turned into a JSON body of its own field list. Setting
  // Content-Type to undefined here clears the default so axios leaves the
  // FormData untouched and the platform (browser XHR / RN's networking
  // layer) sets the correct multipart/form-data header, boundary included -
  // "multipart/form-data" typed by hand would be missing that boundary and
  // the backend's ParseMultipartForm couldn't parse it either.
  const response = await apiClient.patch<ProfileUser>('/users/me/picture', form, {
    headers: { 'Content-Type': undefined },
  })
  return response.data
}

export async function deleteAvatar(): Promise<ProfileUser> {
  const response = await apiClient.delete<ProfileUser>('/users/me/picture')
  return response.data
}

export async function deleteAccount(): Promise<void> {
  await apiClient.delete('/users/me')
}
