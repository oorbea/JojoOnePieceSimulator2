import { apiClient } from '@/shared/api/client'
import type { ProfileUser } from '@/features/profile/types/profile.types'

export async function getMe(): Promise<ProfileUser> {
  const response = await apiClient.get<ProfileUser>('/users/me')
  return response.data
}

export async function updateUsername(username: string): Promise<ProfileUser> {
  const response = await apiClient.patch<ProfileUser>('/users/me', { username })
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
  // React Native's FormData accepts this { uri, name, type } shape directly;
  // on web the same shape works because apiClient/axios + RN Web's FormData
  // polyfill both forward it as-is to the underlying fetch/XHR.
  form.append('picture', {
    uri: asset.uri,
    name: asset.fileName,
    type: asset.mimeType,
  } as unknown as Blob)

  const response = await apiClient.patch<ProfileUser>('/users/me/picture', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
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
