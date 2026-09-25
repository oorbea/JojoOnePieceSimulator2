import { useState } from 'react'

import { postDevLogin } from '@/features/auth/api/auth.api'
import { rememberDevAccount } from '@/features/auth/lib/dev-login-recent-accounts'
import { toAppError, type AppError } from '@/shared/api/errors'
import { setDevRefreshToken } from '@/shared/api/dev-refresh-token'
import { fromUserResponse, useSessionStore } from '@/shared/stores/session.store'

// Mirrors use-google-auth.ts's completeSignIn shape (isLoading/error state,
// setSession on success) without any of Google's redirect/popup machinery -
// dev-login is a single synchronous POST. Web-only in practice (dev-login-
// container.tsx never renders on native), but nothing here assumes that.
export function useDevLogin() {
  const setSession = useSessionStore((state) => state.setSession)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<AppError | null>(null)

  const signIn = async (name: string, admin: boolean) => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await postDevLogin(name, admin)
      if (data.refreshToken) setDevRefreshToken(data.refreshToken)
      rememberDevAccount({ name, admin })
      setSession({ accessToken: data.accessToken, user: fromUserResponse(data.user) })
    } catch (err) {
      setError(toAppError(err))
    } finally {
      setIsLoading(false)
    }
  }

  return { signIn, isLoading, error }
}
