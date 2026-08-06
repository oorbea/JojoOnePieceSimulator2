import * as Google from 'expo-auth-session/providers/google'
import * as WebBrowser from 'expo-web-browser'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Platform } from 'react-native'

import { postGoogleAuth } from '@/features/auth/api/auth.api'
import { toAppError, type AppError } from '@/shared/api/errors'
import { env } from '@/shared/config/env'
import { useSessionStore } from '@/shared/stores/session.store'

// Required so the auth popup/tab closes itself after Google redirects back
// (no-op on native, matters on web for platforms other than Google itself —
// see the web redirect flow below for why Google needs different handling).
WebBrowser.maybeCompleteAuthSession()

const WEB_STATE_KEY = 'jojo_google_auth_state'
const WEB_NONCE_KEY = 'jojo_google_auth_nonce'

// accounts.google.com sends its own strict Cross-Origin-Opener-Policy header.
// That severs `window.opener` the moment a popup navigates there, permanently
// for that popup's lifetime — even after it redirects back to our own origin.
// So the popup can never signal completion back to the tab that opened it
// (see ExpoWebBrowser.web.js's `maybeCompleteAuthSession`, which needs
// `window.opener`/`window.parent` to still work). There's no fix on our side
// for that; the standard workaround is to not use a popup at all on web —
// full-page redirect instead, with the response read back from the URL hash
// on next load of this same page.
function buildWebAuthUrl(state: string, nonce: string): string {
  const params = new URLSearchParams({
    client_id: env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID,
    redirect_uri: `${window.location.origin}${window.location.pathname}`,
    response_type: 'id_token',
    scope: 'openid email profile',
    prompt: 'select_account',
    state,
    nonce,
  })
  return `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`
}

function decodeJwtPayload(idToken: string): Record<string, unknown> {
  const payload = idToken.split('.')[1] ?? ''
  const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
  const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=')
  return JSON.parse(atob(padded))
}

export function useGoogleAuth() {
  const { t } = useTranslation()
  const setSession = useSessionStore((state) => state.setSession)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<AppError | null>(null)
  const hasHandledWebRedirect = useRef(false)

  const [request, response, promptAsync] = Google.useAuthRequest({
    webClientId: env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID,
    iosClientId: env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID,
    androidClientId: env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID,
  })

  const completeSignIn = async (idToken: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await postGoogleAuth(idToken)
      await setSession({
        accessToken: data.accessToken,
        expiresAt: data.expiresAt,
        user: {
          id: data.user.id,
          email: data.user.email,
          username: data.user.username,
          completeName: data.user.completeName,
          picture: data.user.avatar || null,
          role: data.user.role,
          language: data.user.language,
        },
      })
    } catch (err) {
      setError(toAppError(err))
    } finally {
      setIsLoading(false)
    }
  }

  // Web-only: read the id_token back out of the URL hash after Google's
  // full-page redirect lands us back on this route.
  useEffect(() => {
    if (Platform.OS !== 'web' || hasHandledWebRedirect.current) return

    const hash = window.location.hash.startsWith('#') ? window.location.hash.slice(1) : ''
    if (!hash) return

    const params = new URLSearchParams(hash)
    const idToken = params.get('id_token')
    const returnedState = params.get('state')
    if (!idToken) return

    hasHandledWebRedirect.current = true
    window.history.replaceState(null, '', window.location.pathname)

    // Deferred so no setState call runs synchronously in the effect body —
    // every update happens inside this async callback's microtasks instead.
    void (async () => {
      const expectedState = window.sessionStorage.getItem(WEB_STATE_KEY)
      const expectedNonce = window.sessionStorage.getItem(WEB_NONCE_KEY)
      window.sessionStorage.removeItem(WEB_STATE_KEY)
      window.sessionStorage.removeItem(WEB_NONCE_KEY)

      // "state mismatch" and "nonce mismatch" are the same failure from the
      // user's point of view, so they get the same message; the technical
      // detail stays in the thrown Error, not on screen.
      if (!expectedState || returnedState !== expectedState) {
        setError(toAppError(new Error(t('auth.googleSignInError'))))
        return
      }

      const nonce = decodeJwtPayload(idToken).nonce
      if (!expectedNonce || nonce !== expectedNonce) {
        setError(toAppError(new Error(t('auth.googleSignInError'))))
        return
      }

      await completeSignIn(idToken)
    })()
    // Runs once on mount to inspect the URL left by Google's redirect.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Native (iOS/Android): expo-auth-session's own in-app browser session,
  // unaffected by the web-only COOP issue above.
  useEffect(() => {
    if (Platform.OS === 'web') return
    if (response?.type !== 'success') return

    void (async () => {
      const idToken = response.authentication?.idToken
      if (!idToken) {
        setError(toAppError(new Error(t('auth.googleNoIdToken'))))
        return
      }
      await completeSignIn(idToken)
    })()
    // response is a new object identity each auth attempt — safe dep.
    // completeSignIn is intentionally excluded, it's redefined every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [response])

  const signIn = () => {
    setError(null)

    if (Platform.OS === 'web') {
      const state = crypto.randomUUID()
      const nonce = crypto.randomUUID()
      window.sessionStorage.setItem(WEB_STATE_KEY, state)
      window.sessionStorage.setItem(WEB_NONCE_KEY, nonce)
      window.location.assign(buildWebAuthUrl(state, nonce))
      return
    }

    void promptAsync()
  }

  return { signIn, isLoading, error, isReady: Platform.OS === 'web' ? true : !!request }
}
