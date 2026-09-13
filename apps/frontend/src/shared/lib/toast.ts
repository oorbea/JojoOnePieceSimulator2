import { toast } from 'burnt'

import type { AppError } from '@/shared/api/errors'
import type { FieldError } from '@/shared/contracts/dto'
import i18n from '@/shared/i18n'

// Single place feature code calls to surface API/mutation failures — never
// render raw error objects or stack traces directly in a component. Called
// from MutationCache.onError (query-provider.tsx), which is configured once
// outside the React tree - there's no component to call useTranslation()
// from, so this resolves the current language directly off the i18next
// singleton instead. error.code is the stable identifier the backend now
// emits (see dto.ErrorResponse); i18next's defaultValue degrades gracefully
// to the backend's English text for any code this catalog doesn't have yet.
export function showErrorToast(error: AppError) {
  const fallback = error.message || i18n.t('common.somethingWentWrong')
  const title = error.code ? i18n.t(`errors.${error.code}`, { defaultValue: fallback }) : fallback
  toast({ title, preset: 'error' })
}

// Localizes one backend field-validation error: fieldError.code is an i18n
// key (e.g. "validation.nameRequired") the same locale catalogs already
// carry for the frontend's own zod schemas - see i18n-multi-language.md's
// "Not done" follow-up this closes. Falls back to the backend's English
// message when a client doesn't recognize the code, same degrade-to-English
// pattern as showErrorToast's top-level error code.
export function translateFieldError(fieldError: FieldError): string {
  return i18n.t(fieldError.code, { defaultValue: fieldError.message })
}

export function showSuccessToast(message: string) {
  toast({
    title: message,
    preset: 'done',
  })
}
