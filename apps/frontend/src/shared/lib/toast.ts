import { toast } from 'burnt'

import type { AppError } from '@/shared/api/errors'
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

export function showSuccessToast(message: string) {
  toast({
    title: message,
    preset: 'done',
  })
}
