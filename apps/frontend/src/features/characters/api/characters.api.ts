import { Platform } from 'react-native'

import { apiClient } from '@/shared/api/client'
import { assertContract } from '@/shared/api/assert-contract'
import {
  jojoCharacterPageResponseSchema,
  jojoCharacterResponseSchema,
  onePieceCharacterPageResponseSchema,
  onePieceCharacterResponseSchema,
} from '@/shared/contracts/dto'
import type { CataloguePage } from '@/shared/hooks/use-paginated-catalogue'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import type { CharacterTranslationFormValues } from '@/shared/lib/character-translations'
import type { Locale } from '@/shared/contracts/enums'
import type {
  JojoCharacterFilters,
  JojoCharacterInput,
  JojoCharacterResponse,
  OnePieceCharacterFilters,
  OnePieceCharacterInput,
  OnePieceCharacterResponse,
} from '@/features/characters/types/characters.types'

// Same web/native FormData branching every picture-upload call site shares -
// see stages.api.ts's uploadStagePicture for the full rationale.
async function buildPictureForm(asset: PickedPicture): Promise<FormData> {
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
  return form
}

export async function getJojoCharacters(
  filters?: JojoCharacterFilters
): Promise<JojoCharacterResponse[]> {
  const response = await apiClient.get<JojoCharacterResponse[]>('/jojo-characters', {
    params: filters,
  })
  if (__DEV__) {
    for (const c of response.data)
      assertContract(jojoCharacterResponseSchema, c, 'GET /jojo-characters[]')
  }
  return response.data
}

export async function getJojoCharactersPage(
  filters: JojoCharacterFilters | undefined,
  cursor: string | undefined,
  limit: number
): Promise<CataloguePage<JojoCharacterResponse>> {
  const response = await apiClient.get<CataloguePage<JojoCharacterResponse>>('/jojo-characters', {
    params: { ...filters, limit, cursor },
  })
  if (__DEV__)
    assertContract(jojoCharacterPageResponseSchema, response.data, 'GET /jojo-characters?limit=')
  return response.data
}

export async function getJojoCharacter(id: string): Promise<JojoCharacterResponse> {
  const response = await apiClient.get<JojoCharacterResponse>(`/jojo-characters/${id}`)
  if (__DEV__)
    assertContract(jojoCharacterResponseSchema, response.data, 'GET /jojo-characters/:id')
  return response.data
}

export async function getJojoCharacterTranslations(
  id: string
): Promise<Partial<Record<Locale, CharacterTranslationFormValues>>> {
  const response = await apiClient.get<{
    translations: Partial<Record<Locale, CharacterTranslationFormValues>>
  }>(`/jojo-characters/${id}/translations`)
  return response.data.translations
}

export async function createJojoCharacter(
  input: JojoCharacterInput
): Promise<JojoCharacterResponse> {
  const response = await apiClient.post<JojoCharacterResponse>('/jojo-characters', input)
  return response.data
}

export async function updateJojoCharacter(
  id: string,
  input: JojoCharacterInput
): Promise<JojoCharacterResponse> {
  const response = await apiClient.put<JojoCharacterResponse>(`/jojo-characters/${id}`, input)
  return response.data
}

export async function uploadJojoCharacterPicture(
  id: string,
  asset: PickedPicture
): Promise<JojoCharacterResponse> {
  const form = await buildPictureForm(asset)
  const response = await apiClient.patch<JojoCharacterResponse>(
    `/jojo-characters/${id}/picture`,
    form,
    {
      headers: { 'Content-Type': undefined },
    }
  )
  return response.data
}

export async function deleteJojoCharacter(id: string): Promise<void> {
  await apiClient.delete(`/jojo-characters/${id}`)
}

export async function getOnePieceCharacters(
  filters?: OnePieceCharacterFilters
): Promise<OnePieceCharacterResponse[]> {
  const response = await apiClient.get<OnePieceCharacterResponse[]>('/one-piece-characters', {
    params: filters,
  })
  if (__DEV__) {
    for (const c of response.data) {
      assertContract(onePieceCharacterResponseSchema, c, 'GET /one-piece-characters[]')
    }
  }
  return response.data
}

export async function getOnePieceCharactersPage(
  filters: OnePieceCharacterFilters | undefined,
  cursor: string | undefined,
  limit: number
): Promise<CataloguePage<OnePieceCharacterResponse>> {
  const response = await apiClient.get<CataloguePage<OnePieceCharacterResponse>>(
    '/one-piece-characters',
    {
      params: { ...filters, limit, cursor },
    }
  )
  if (__DEV__) {
    assertContract(
      onePieceCharacterPageResponseSchema,
      response.data,
      'GET /one-piece-characters?limit='
    )
  }
  return response.data
}

export async function getOnePieceCharacter(id: string): Promise<OnePieceCharacterResponse> {
  const response = await apiClient.get<OnePieceCharacterResponse>(`/one-piece-characters/${id}`)
  if (__DEV__)
    assertContract(onePieceCharacterResponseSchema, response.data, 'GET /one-piece-characters/:id')
  return response.data
}

export async function getOnePieceCharacterTranslations(
  id: string
): Promise<Partial<Record<Locale, CharacterTranslationFormValues>>> {
  const response = await apiClient.get<{
    translations: Partial<Record<Locale, CharacterTranslationFormValues>>
  }>(`/one-piece-characters/${id}/translations`)
  return response.data.translations
}

export async function createOnePieceCharacter(
  input: OnePieceCharacterInput
): Promise<OnePieceCharacterResponse> {
  const response = await apiClient.post<OnePieceCharacterResponse>('/one-piece-characters', input)
  return response.data
}

export async function updateOnePieceCharacter(
  id: string,
  input: OnePieceCharacterInput
): Promise<OnePieceCharacterResponse> {
  const response = await apiClient.put<OnePieceCharacterResponse>(
    `/one-piece-characters/${id}`,
    input
  )
  return response.data
}

export async function uploadOnePieceCharacterPicture(
  id: string,
  asset: PickedPicture
): Promise<OnePieceCharacterResponse> {
  const form = await buildPictureForm(asset)
  const response = await apiClient.patch<OnePieceCharacterResponse>(
    `/one-piece-characters/${id}/picture`,
    form,
    { headers: { 'Content-Type': undefined } }
  )
  return response.data
}

export async function deleteOnePieceCharacter(id: string): Promise<void> {
  await apiClient.delete(`/one-piece-characters/${id}`)
}
