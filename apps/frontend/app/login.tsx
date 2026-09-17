import { Redirect } from 'expo-router'

import { LoginContainer } from '@/features/auth'
import { readPendingInvite } from '@/features/game/lib/pending-invite'
import { LoadingScreen } from '@/shared/components/presentational/loading-screen'
import { useSessionStore } from '@/shared/stores/session.store'

export default function LoginRoute() {
  const session = useSessionStore((state) => state.session)
  const isHydrated = useSessionStore((state) => state.isHydrated)

  // Hydration now involves an async silent-refresh round trip (see
  // session.store.ts), so it can no longer be assumed instantaneous - render
  // the loading screen instead of flashing the login UI while it's pending.
  if (!isHydrated) return <LoadingScreen />
  if (session) {
    // A visitor who opened a lobby invite link while signed out gets
    // bounced here by /join/[token] (JoinInviteContainer), which stashes
    // the token first - see pending-invite.ts's doc for why this can't
    // travel as a ?returnTo= query param instead (Google's web redirect
    // flow lands back on this exact /login URL, not on the invite's).
    const pending = readPendingInvite()
    return <Redirect href={pending ? (`/join/${pending}` as never) : '/'} />
  }

  return <LoginContainer />
}
