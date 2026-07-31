import { useGoogleAuth } from '@/features/auth/hooks/use-google-auth'
import { LoginScreen } from '@/features/auth/components/presentational/login-screen'

export function LoginContainer() {
  const { signIn, isLoading, isReady, error } = useGoogleAuth()

  return (
    <LoginScreen onSignIn={signIn} isLoading={isLoading} isReady={isReady} error={error} />
  )
}
