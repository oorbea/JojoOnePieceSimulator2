import { z } from 'zod'

import { SUPPORTED_LOCALES } from '@/shared/i18n'
import type { Locale } from '@/shared/contracts/enums'
import type { TranslationRequest } from '@/shared/contracts/dto'

// Shared between the Stand and Devil Fruit admin forms - both Powers carry
// the exact same translations shape (description + skills per locale, see
// the vault's i18n-multi-language.md). One module instead of duplicating
// this in each feature's types file.

// Aliased (not re-declared) to the generated wire shape, so a field added
// to dto.TranslationRequest surfaces as a type error at every call site
// here instead of silently drifting - see
// ObsidianVault/contratos-tipos-generados.md's note on toTranslationsPayload.
export type TranslationFormValues = TranslationRequest
export type PowerTranslationsFormValues = Record<Locale, TranslationFormValues>

// Factories, not singletons: react-hook-form's defaultValues become the
// form's live internal state for nested paths, so handing out the same
// object reference to two form instances (e.g. opening "create" twice, or
// two tests in the same file) lets edits in one leak into the other. Every
// caller gets its own fresh object tree.
export function createEmptyTranslationForm(): TranslationFormValues {
  return { description: '', skills: [] }
}

export function createEmptyTranslationsForm(): PowerTranslationsFormValues {
  return {
    'en-GB': createEmptyTranslationForm(),
    'es-ES': createEmptyTranslationForm(),
    'ca-ES': createEmptyTranslationForm(),
  }
}

// Messages are i18n keys, not display strings - see the vault's zod
// decision. Resolve with t(errors.field?.message) at the render site.
const translationContentSchema = z.object({
  description: z
    .string()
    .min(1, 'validation.descriptionRequired')
    .max(1000, 'validation.descriptionTooLong'),
  skills: z.array(z.string().min(1)).min(1, 'validation.skillsRequired'),
})

// A locale is either fully filled or fully empty - no partial overrides,
// same rule the backend enforces in dto/translation_request.go. en-GB alone
// is mandatory; es-ES/ca-ES may be entirely blank (then they're dropped
// from the payload) but can't be half-filled.
const optionalTranslationContentSchema = z.object({
  description: z.string().max(1000, 'validation.descriptionTooLong'),
  skills: z.array(z.string().min(1)),
})

export const powerTranslationsFormSchema = z
  .object({
    'en-GB': translationContentSchema,
    'es-ES': optionalTranslationContentSchema,
    'ca-ES': optionalTranslationContentSchema,
  })
  .superRefine((value, ctx) => {
    for (const locale of ['es-ES', 'ca-ES'] as const) {
      const content = value[locale]
      const hasDescription = content.description.trim().length > 0
      const hasSkills = content.skills.length > 0
      if (hasDescription === hasSkills) continue
      ctx.addIssue({
        code: 'custom',
        message: 'validation.translationIncomplete',
        path: [locale, hasDescription ? 'skills' : 'description'],
      })
    }
  })

// Drops any locale that's entirely blank - the backend expects a map with
// only the locales actually being submitted (an absent locale is untouched,
// not cleared).
export function toTranslationsPayload(
  values: PowerTranslationsFormValues
): Partial<Record<Locale, TranslationFormValues>> {
  const out: Partial<Record<Locale, TranslationFormValues>> = {}
  for (const locale of SUPPORTED_LOCALES) {
    const content = values[locale]
    if (content.description.trim().length > 0 || content.skills.length > 0) {
      out[locale] = content
    }
  }
  return out
}

// Fills in every locale absent from GET .../translations with an empty
// form value, so the modal always has all three keys to render tabs for.
export function fromTranslationsResponse(
  translations: Partial<Record<string, TranslationFormValues>>
): PowerTranslationsFormValues {
  const out = createEmptyTranslationsForm()
  for (const locale of SUPPORTED_LOCALES) {
    const content = translations[locale]
    if (content) out[locale] = { description: content.description, skills: content.skills }
  }
  return out
}
