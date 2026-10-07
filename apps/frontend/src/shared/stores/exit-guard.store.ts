import { create } from 'zustand'

// One place that decides whether leaving the current screen needs the player's
// OK. While a game is in progress the game room registers itself here; every
// way out of it - the bottom/top navigation, logout, the system back button or
// swipe - then goes through `request`, which asks first instead of navigating.
// Lives in shared/ so the app shell (which owns the navigation and the sheet)
// needn't import the game feature.
export type ExitGuardRegistration = {
  // Leaves the game for good (LEAVE command + clearing its client state). No
  // navigation: the caller's own destination follows.
  leaveGame: () => void
  // Removes the browser-history guard entries (see lib/back-guard.ts) so no
  // dead entries are left behind the destination. May be a no-op.
  releaseHistory: () => Promise<void>
}

type ExitGuardState = {
  registration: ExitGuardRegistration | null
  // The exit waiting for the player's answer, shown by the app shell.
  pending: { proceed: () => void } | null

  register: (registration: ExitGuardRegistration | null) => void
  // Runs `proceed` straight away when nothing needs protecting, otherwise
  // parks it until confirm()/cancel().
  request: (proceed: () => void) => void
  confirm: () => Promise<void>
  cancel: () => void
}

export const useExitGuardStore = create<ExitGuardState>((set, get) => ({
  registration: null,
  pending: null,

  register: (registration) => {
    // A game that stops being protected must not leave a stale question open.
    set(registration ? { registration } : { registration: null, pending: null })
  },

  request: (proceed) => {
    if (!get().registration) {
      proceed()
      return
    }
    set({ pending: { proceed } })
  },

  confirm: async () => {
    const { pending, registration } = get()
    if (!pending) return
    set({ pending: null })
    // History first: the destination must land on the page's own entry, not on
    // top of leftover guard entries.
    await registration?.releaseHistory()
    registration?.leaveGame()
    pending.proceed()
  },

  cancel: () => set({ pending: null }),
}))
