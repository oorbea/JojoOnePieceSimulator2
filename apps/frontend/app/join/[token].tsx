import { JoinInviteContainer } from '@/features/game'

// Deliberately outside app/(app)/ - same treatment as /login. A lobby
// invite link must be openable by a visitor with no session at all, so it
// must not go through (app)/_layout.tsx's auth-guard Redirect (which would
// drop the token entirely - see JoinInviteContainer's pending-invite stash
// for how the token survives the detour through /login instead).
export default function JoinInviteRoute() {
  return <JoinInviteContainer />
}
