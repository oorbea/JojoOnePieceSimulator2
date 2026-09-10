import { z } from 'zod'

import { SUPPORTED_LOCALES } from '@/shared/i18n'
import type { Locale } from '@/shared/contracts/enums'
import type { CharacterTranslationRequest } from '@/shared/contracts/dto'

// Character's translations shape (description only, no skills - same as
// Stage) but only en-GB is mandatory on write (the Power rule, not the
// Stage one - see dto/translation_request.go's validateCharacterTranslations).
// Since there's only one field, es-ES/ca-ES have nothing that could be
// "half-filled" the way Power's description+skills pair can, so no
// superRefine is needed here - just a plain optional-string schema for the
// two optional locales.

export type CharacterTranslationFormValues = CharacterTranslationRequest
export type CharacterTranslationsFormValues = Record<Locale, CharacterTranslationFormValues>

export function createEmptyCharacterTranslationForm(): CharacterTranslationFormValues {
  return { description: '' }
}

export function createEmptyCharacterTranslationsForm(): CharacterTranslationsFormValues {
  return {
    'en-GB': createEmptyCharacterTranslationForm(),
    'es-ES': createEmptyCharacterTranslationForm(),
    'ca-ES': createEmptyCharacterTranslationForm(),
  }
}

const requiredContentSchema = z.object({
  description: z
    .string()
    .min(1, 'validation.descriptionRequired')
    .max(1000, 'validation.descriptionTooLong'),
})

const optionalContentSchema = z.object({
  description: z.string().max(1000, 'validation.descriptionTooLong'),
})

export const characterTranslationsFormSchema = z.object({
  'en-GB': requiredContentSchema,
  'es-ES': optionalContentSchema,
  'ca-ES': optionalContentSchema,
})

// Drops any optional locale that's entirely blank - same convention as
// toTranslationsPayload.
export function toCharacterTranslationsPayload(
  values: CharacterTranslationsFormValues
): Partial<Record<Locale, CharacterTranslationFormValues>> {
  const out: Partial<Record<Locale, CharacterTranslationFormValues>> = {}
  for (const locale of SUPPORTED_LOCALES) {
    const content = values[locale]
    if (content.description.trim().length > 0) {
      out[locale] = content
    }
  }
  return out
}

export function fromCharacterTranslationsResponse(
  translations: Partial<Record<string, CharacterTranslationFormValues>>
): CharacterTranslationsFormValues {
  const out = createEmptyCharacterTranslationsForm()
  for (const locale of SUPPORTED_LOCALES) {
    const content = translations[locale]
    if (content) out[locale] = { description: content.description }
  }
  return out
}
