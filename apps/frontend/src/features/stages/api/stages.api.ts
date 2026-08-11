import { Platform } from 'react-native'

import { apiClient } from '@/shared/api/client'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import type { StageTranslationFormValues } from '@/shared/lib/stage-translations'
import type { Locale } from '@/shared/lib/zod'
import type { StageFilters, StageInput, StageResponse } from '@/features/stages/types/stages.types'

export async function getStages(filters?: StageFilters): Promise<StageResponse[]> {
  const response = await apiClient.get<StageResponse[]>('/stages', { params: filters })
  return response.data
}

export async function getStage(id: string): Promise<StageResponse> {
  const response = await apiClient.get<StageResponse>(`/stages/${id}`)
  return response.data
}

// Admin-only: every locale's content at once, for the edit form's
// LocaleTabs. Mirrors dto.StageTranslationsResponse.
export async function getStageTranslations(
  id: string
): Promise<Partial<Record<Locale, StageTranslationFormValues>>> {
  const response = await apiClient.get<{ translations: Partial<Record<Locale, StageTranslationFormValues>> }>(
    `/stages/${id}/translations`
  )
  return response.data.translations
}

export async function createStage(input: StageInput): Promise<StageResponse> {
  const response = await apiClient.post<StageResponse>('/stages', input)
  return response.data
}

export async function updateStage(id: string, input: StageInput): Promise<StageResponse> {
  const response = await apiClient.put<StageResponse>(`/stages/${id}`, input)
  return response.data
}

// Same web/native FormData branching as stands.api.ts's uploadStandPicture -
// see that file for the full rationale (react-native-web's FormData needs a
// real Blob, native's needs the { uri, name, type } bridge shape, and axios
// must be told to drop its default JSON Content-Type or it stringifies the
// FormData body instead of sending multipart/form-data).
export async function uploadStagePicture(id: string, asset: PickedPicture): Promise<StageResponse> {
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

  const response = await apiClient.patch<StageResponse>(`/stages/${id}/picture`, form, {
    headers: { 'Content-Type': undefined },
  })
  return response.data
}

export async function deleteStage(id: string): Promise<void> {
  await apiClient.delete(`/stages/${id}`)
}
