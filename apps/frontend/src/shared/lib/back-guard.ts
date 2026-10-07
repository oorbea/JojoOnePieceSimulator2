// Intercepts the system "back" action (Android back button / edge-swipe
// gesture in an installed PWA) so a stray swipe mid-game asks before it
// navigates away.
//
// A browser can't cancel back, only react to it: guard entries are pushed on
// top of the current page (same URL), so a back press merely pops one and fires
// `popstate`, and the caller is told to ask the player what to do.
//
// Two Chrome behaviours shape the design (the first version - one entry,
// re-pushed from inside the popstate handler - confirmed once and then let the
// next back through):
//
//  1. History-manipulation intervention: entries added by pushState WITHOUT a
//     user activation are flagged "skippable", and the back button jumps over
//     them. An entry pushed from a popstate handler or a timer has no
//     activation, so it was skipped. Entries are therefore (re)filled from
//     real user input (`pointerup` / `keydown`), which is activation.
//  2. A single press must not use up the protection. GUARD_DEPTH entries are
//     kept, so several presses in a row (before the player taps anything) are
//     each caught; any tap tops the stack back up.
//
// Each entry carries its level (1..GUARD_DEPTH); the page's own entry is level
// 0. Comparing the level a popstate lands on with the current one tells a back
// press (level dropped) from a forward navigation (level rose).
const GUARD_STATE_KEY = 'jopsBackGuard'
export const GUARD_DEPTH = 3

// How long release() waits for the popstate its history.go() causes.
const RELEASE_TIMEOUT_MS = 300

// Loosely typed on purpose: the real Window overloads addEventListener per
// event name, and one structural type has to accept both it and the test fake.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Listener = (event: any) => void

export type BackGuardWindow = {
  history: Pick<History, 'pushState' | 'go' | 'state'>
  addEventListener: (type: 'popstate' | 'pointerup' | 'keydown', listener: Listener) => void
  removeEventListener: (type: 'popstate' | 'pointerup' | 'keydown', listener: Listener) => void
}

export type BackGuard = {
  // Removes the guard entries again (history.go back to the page's own entry)
  // and resolves once that has happened. Call it BEFORE navigating away on
  // purpose, so no dead guard entries are left behind the destination.
  // Idempotent.
  release: () => Promise<void>
}

export function guardLevelOf(state: unknown): number {
  if (typeof state !== 'object' || state === null || !(GUARD_STATE_KEY in state)) return 0
  const level = (state as Record<string, unknown>)[GUARD_STATE_KEY]
  return typeof level === 'number' && level > 0 ? level : 1
}

export function installBackGuard(win: BackGuardWindow, onBackAttempt: () => void): BackGuard {
  let level = guardLevelOf(win.history.state)
  let released = false

  const topUp = () => {
    while (level < GUARD_DEPTH) {
      level += 1
      win.history.pushState({ [GUARD_STATE_KEY]: level }, '')
    }
  }

  const onPopState = (event: { state: unknown }) => {
    const next = guardLevelOf(event.state)
    const wentBack = next < level
    level = next
    if (wentBack) onBackAttempt()
  }

  // Only real input counts as user activation; this is what keeps the entries
  // from being flagged skippable. Cheap no-op once the stack is full.
  const onActivation = () => topUp()

  topUp()
  win.addEventListener('popstate', onPopState as Listener)
  win.addEventListener('pointerup', onActivation as Listener)
  win.addEventListener('keydown', onActivation as Listener)

  const release = () =>
    new Promise<void>((resolve) => {
      if (released) return resolve()
      released = true
      win.removeEventListener('popstate', onPopState as Listener)
      win.removeEventListener('pointerup', onActivation as Listener)
      win.removeEventListener('keydown', onActivation as Listener)
      if (level === 0) return resolve()

      const steps = level
      level = 0
      const done = () => {
        clearTimeout(timer)
        win.removeEventListener('popstate', done as Listener)
        resolve()
      }
      const timer = setTimeout(done, RELEASE_TIMEOUT_MS)
      win.addEventListener('popstate', done as Listener)
      win.history.go(-steps)
    })

  return { release }
}
