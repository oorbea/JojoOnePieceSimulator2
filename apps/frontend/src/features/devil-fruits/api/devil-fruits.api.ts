import { Platform } from 'react-native'

import { apiClient } from '@/shared/api/client'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import type { TranslationFormValues } from '@/shared/lib/power-translations'
import type {
  DevilFruitFilters,
  DevilFruitInput,
  DevilFruitResponse,
} from '@/features/devil-fruits/types/devil-fruits.types'

export async function getDevilFruits(filters?: DevilFruitFilters): Promise<DevilFruitResponse[]> {
  const response = await apiClient.get<DevilFruitResponse[]>('/devil-fruits', { params: filters })
  return response.data
}

export async function getDevilFruit(id: string): Promise<DevilFruitResponse> {
  const response = await apiClient.get<DevilFruitResponse>(`/devil-fruits/${id}`)
  return response.data
}

// Admin-only: every locale's content at once, for the edit form's
// LocaleTabs. Mirrors dto.PowerTranslationsResponse.
export async function getDevilFruitTranslations(
  id: string
): Promise<Partial<Record<string, TranslationFormValues>>> {
  const response = await apiClient.get<{ translations: Partial<Record<string, TranslationFormValues>> }>(
    `/devil-fruits/${id}/translations`
  )
  return response.data.translations
}

export async function createDevilFruit(input: DevilFruitInput): Promise<DevilFruitResponse> {
  const response = await apiClient.post<DevilFruitResponse>('/devil-fruits', input)
  return response.data
}

export async function updateDevilFruit(id: string, input: DevilFruitInput): Promise<DevilFruitResponse> {
  const response = await apiClient.put<DevilFruitResponse>(`/devil-fruits/${id}`, input)
  return response.data
}

// See stands.api.ts's uploadStandPicture for the full web/native FormData
// rationale — identical shape, different resource.
export async function uploadDevilFruitPicture(
  id: string,
  asset: PickedPicture
): Promise<DevilFruitResponse> {
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

  const response = await apiClient.patch<DevilFruitResponse>(`/devil-fruits/${id}/picture`, form, {
    headers: { 'Content-Type': undefined },
  })
  return response.data
}

export async function deleteDevilFruit(id: string): Promise<void> {
  await apiClient.delete(`/devil-fruits/${id}`)
}
