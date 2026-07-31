import * as Google from 'expo-auth-session/providers/google'
import * as WebBrowser from 'expo-web-browser'
import { useEffect, useState } from 'react'

import { postGoogleAuth } from '@/features/auth/api/auth.api'
import { toAppError, type AppError } from '@/shared/api/errors'
import { env } from '@/shared/config/env'
import { useSessionStore } from '@/shared/stores/session.store'

// Required so the auth popup/tab closes itself after Google redirects back
// (no-op on native, matters on web).
WebBrowser.maybeCompleteAuthSession()

export function useGoogleAuth() {
  const setSession = useSessionStore((state) => state.setSession)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<AppError | null>(null)

  const [request, response, promptAsync] = Google.useAuthRequest({
    webClientId: env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID,
    iosClientId: env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
    androidClientId: env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  })

  useEffect(() => {
    if (response?.type !== 'success') return

    // Deferred so no setState call runs synchronously in the effect body —
    // every update happens inside this async callback's microtasks instead.
    void (async () => {
      const idToken = response.authentication?.idToken
      if (!idToken) {
        setError(toAppError(new Error('Google did not return an ID token')))
        return
      }

      setIsLoading(true)
      setError(null)

      try {
        const data = await postGoogleAuth(idToken)
        await setSession({
          accessToken: data.accessToken,
          expiresAt: data.expiresAt,
          user: data.user,
        })
      } catch (err) {
        setError(toAppError(err))
      } finally {
        setIsLoading(false)
      }
    })()
    // response is a new object identity each auth attempt — safe dep.
  }, [response, setSession])

  const signIn = () => {
    setError(null)
    void promptAsync()
  }

  return { signIn, isLoading, error, isReady: !!request }
}
