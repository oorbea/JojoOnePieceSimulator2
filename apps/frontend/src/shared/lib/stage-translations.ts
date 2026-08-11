import { z } from 'zod'

import { SUPPORTED_LOCALES } from '@/shared/i18n'
import type { Locale } from '@/shared/lib/zod'

// Stage's translations shape (description only, no skills - see the vault's
// game-stage-content.md) is close enough to Power's that reuse was tempting,
// but the two hard rules differ: no skills array, and every locale is
// mandatory on write (Power only requires en-GB, see
// dto/translation_request.go's validateStageTranslations vs
// validateTranslations). Rather than parameterize power-translations.ts with
// a "which locales are required" flag, this is a small sibling module - the
// "all-or-nothing per optional locale" superRefine simply doesn't apply here.

export type StageTranslationFormValues = { description: string }
export type StageTranslationsFormValues = Record<Locale, StageTranslationFormValues>

// A function, not a singleton: react-hook-form's defaultValues become the
// form's live internal state for nested paths, so every caller needs its own
// fresh object tree (same reasoning as power-translations.ts).
export function createEmptyStageTranslationForm(): StageTranslationFormValues {
  return { description: '' }
}

export function createEmptyStageTranslationsForm(): StageTranslationsFormValues {
  return {
    'en-GB': createEmptyStageTranslationForm(),
    'es-ES': createEmptyStageTranslationForm(),
    'ca-ES': createEmptyStageTranslationForm(),
  }
}

// Messages are i18n keys, not display strings - resolve with
// t(errors.field?.message) at the render site, same convention as
// power-translations.ts.
const stageTranslationContentSchema = z.object({
  description: z
    .string()
    .min(1, 'validation.descriptionRequired')
    .max(1000, 'validation.descriptionTooLong'),
})

// Every locale is required, unlike Power's "en-GB mandatory, others
// all-or-nothing" rule - so this is a plain object schema, no superRefine.
export const stageTranslationsFormSchema = z.object({
  'en-GB': stageTranslationContentSchema,
  'es-ES': stageTranslationContentSchema,
  'ca-ES': stageTranslationContentSchema,
})

// Unlike toTranslationsPayload, nothing is dropped - the backend requires
// every locale on every write, so the payload always carries all three.
export function toStageTranslationsPayload(
  values: StageTranslationsFormValues
): Record<Locale, StageTranslationFormValues> {
  const out = {} as Record<Locale, StageTranslationFormValues>
  for (const locale of SUPPORTED_LOCALES) {
    out[locale] = { description: values[locale].description }
  }
  return out
}

// Fills in every locale absent from GET .../translations with an empty
// form value, so the modal always has all three keys to render tabs for -
// same purpose as fromTranslationsResponse.
export function fromStageTranslationsResponse(
  translations: Partial<Record<string, StageTranslationFormValues>>
): StageTranslationsFormValues {
  const out = createEmptyStageTranslationsForm()
  for (const locale of SUPPORTED_LOCALES) {
    const content = translations[locale]
    if (content) out[locale] = { description: content.description }
  }
  return out
}
