import { errorResponseSchema } from '@/shared/lib/zod'

// Normalizes the backend's {error, details?[]} shape (and network/unknown
// failures) into one type every feature can catch and render consistently.
export class AppError extends Error {
  readonly status?: number
  readonly details?: string[]

  constructor(message: string, options?: { status?: number; details?: string[] }) {
    super(message)
    this.name = 'AppError'
    this.status = options?.status
    this.details = options?.details
  }
}

export function isAppError(error: unknown): error is AppError {
  return error instanceof AppError
}

export function toAppError(error: unknown): AppError {
  if (isAppError(error)) return error

  if (error && typeof error === 'object' && 'response' in error) {
    const axiosError = error as {
      response?: { status?: number; data?: unknown }
      message?: string
    }
    const parsed = errorResponseSchema.safeParse(axiosError.response?.data)
    if (parsed.success) {
      return new AppError(parsed.data.error, {
        status: axiosError.response?.status,
        details: parsed.data.details,
      })
    }
    return new AppError(axiosError.message ?? 'Request failed', {
      status: axiosError.response?.status,
    })
  }

  return new AppError(error instanceof Error ? error.message : 'Unknown error')
}
