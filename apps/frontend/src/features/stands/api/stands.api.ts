import { Platform } from 'react-native'

import { apiClient } from '@/shared/api/client'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import type { TranslationFormValues } from '@/shared/lib/power-translations'
import type { StandFilters, StandInput, StandResponse } from '@/features/stands/types/stands.types'

export async function getStands(filters?: StandFilters): Promise<StandResponse[]> {
  const response = await apiClient.get<StandResponse[]>('/stands', { params: filters })
  return response.data
}

export async function getStand(id: string): Promise<StandResponse> {
  const response = await apiClient.get<StandResponse>(`/stands/${id}`)
  return response.data
}

// Admin-only: every locale's content at once, for the edit form's
// LocaleTabs. Mirrors dto.PowerTranslationsResponse.
export async function getStandTranslations(
  id: string
): Promise<Partial<Record<string, TranslationFormValues>>> {
  const response = await apiClient.get<{ translations: Partial<Record<string, TranslationFormValues>> }>(
    `/stands/${id}/translations`
  )
  return response.data.translations
}

export async function createStand(input: StandInput): Promise<StandResponse> {
  const response = await apiClient.post<StandResponse>('/stands', input)
  return response.data
}

export async function updateStand(id: string, input: StandInput): Promise<StandResponse> {
  const response = await apiClient.put<StandResponse>(`/stands/${id}`, input)
  return response.data
}

// Same web/native FormData branching as profile.api.ts's uploadAvatar — see
// that file for the full rationale (react-native-web's FormData needs a
// real Blob, native's needs the { uri, name, type } bridge shape, and axios
// must be told to drop its default JSON Content-Type or it stringifies the
// FormData body instead of sending multipart/form-data).
export async function uploadStandPicture(id: string, asset: PickedPicture): Promise<StandResponse> {
  const form = new FormData()

  if (Platform.OS === 'web') {
    const blob = await (await fetch(asset.uri)).blob()
    form.append('picture', blob, asset.fileName)
  } else {
    form.append('picture', {
      uri: asset.uri,
      name: asset.fileName,
      type: asset.mimeType,
    } as unknown as Blob)
  }

  const response = await apiClient.patch<StandResponse>(`/stands/${id}/picture`, form, {
    headers: { 'Content-Type': undefined },
  })
  return response.data
}

export async function deleteStand(id: string): Promise<void> {
  await apiClient.delete(`/stands/${id}`)
}
