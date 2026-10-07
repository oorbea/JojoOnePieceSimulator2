// Intercepts the system "back" action (Android back button/gesture in an
// installed PWA) so a stray swipe mid-game asks before it navigates away.
//
// A browser can't cancel back, only react to it: an extra history entry is
// pushed on top of the current page ("the guard"), so the first back press
// merely pops that entry - same URL, nothing visible changes - and fires
// `popstate`. The guard is pushed again (so the next press is caught too) and
// the caller is told to ask the player what to do.
const GUARD_STATE_KEY = 'jopsBackGuard'

export type BackGuardWindow = {
  history: Pick<History, 'pushState' | 'back' | 'state'>
  addEventListener: (type: 'popstate', listener: (event: PopStateEvent) => void) => void
  removeEventListener: (type: 'popstate', listener: (event: PopStateEvent) => void) => void
}

function isGuardState(state: unknown): boolean {
  return typeof state === 'object' && state !== null && GUARD_STATE_KEY in state
}

// Returns the uninstall function.
export function installBackGuard(win: BackGuardWindow, onBackAttempt: () => void): () => void {
  const pushGuard = () => win.history.pushState({ [GUARD_STATE_KEY]: true }, '')

  const onPopState = (event: PopStateEvent) => {
    // Landing ON the guard entry (a forward navigation) is not a back press.
    if (isGuardState(event.state)) return
    pushGuard()
    onBackAttempt()
  }

  pushGuard()
  win.addEventListener('popstate', onPopState)

  return () => {
    win.removeEventListener('popstate', onPopState)
    // Still standing on the guard (e.g. the game ended in place): drop it so
    // the player's next back press isn't silently swallowed by a dead entry.
    // If the screen navigated away instead, the current entry is the new
    // page's and must be left alone.
    if (isGuardState(win.history.state)) win.history.back()
  }
}
