import { toast } from 'burnt'

import type { AppError } from '@/shared/api/errors'

// Single place feature code calls to surface API/mutation failures — never
// render raw error objects or stack traces directly in a component.
export function showErrorToast(error: AppError) {
  toast({
    title: error.message || 'Something went wrong',
    preset: 'error',
  })
}

export function showSuccessToast(message: string) {
  toast({
    title: message,
    preset: 'done',
  })
}
