import { create } from 'zustand'

import { AsyncStorage } from '@/shared/lib/async-storage'

// Not in lib.dom yet: the Chromium-only event that hands the page a one-shot
// "show the browser's install dialog" capability.
export type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed'; platform: string }>
}

const BANNER_DISMISSED_KEY = 'jops.pwa-install-banner-dismissed-at'
export const BANNER_SNOOZE_MS = 14 * 24 * 60 * 60 * 1000

// Lives in shared/ (not features/pwa) because the app shell reads it to show
// the install button/banner and it must not import across features: the pwa
// feature's capture hook writes it, the shell only reads it.
type PwaInstallState = {
  // Present only while the browser considers the app installable and has not
  // been asked yet; a prompt can be used exactly once.
  deferred: BeforeInstallPromptEvent | null
  // Already running as an installed app (standalone display mode), or just
  // installed in this session.
  installed: boolean
  bannerDismissedAt: number | null
  isHydrated: boolean

  hydrate: () => Promise<void>
  setDeferred: (event: BeforeInstallPromptEvent | null) => void
  setInstalled: (installed: boolean) => void
  // Opens the browser's install dialog. Resolves true when the user accepted.
  promptInstall: () => Promise<boolean>
  dismissBanner: (now?: number) => Promise<void>
}

export const usePwaInstallStore = create<PwaInstallState>((set, get) => ({
  deferred: null,
  installed: false,
  bannerDismissedAt: null,
  isHydrated: false,

  hydrate: async () => {
    let at: number | null = null
    try {
      const raw = await AsyncStorage.getItem(BANNER_DISMISSED_KEY)
      const parsed = raw ? Number(raw) : NaN
      at = Number.isFinite(parsed) ? parsed : null
    } catch {
      // Storage unavailable (private window, blocked site data): the banner
      // just shows again next visit.
    }
    set({ bannerDismissedAt: at, isHydrated: true })
  },

  setDeferred: (deferred) => set({ deferred }),
  setInstalled: (installed) => set({ installed, ...(installed ? { deferred: null } : {}) }),

  promptInstall: async () => {
    const { deferred } = get()
    if (!deferred) return false
    // The event's prompt() is single-use whatever the answer.
    set({ deferred: null })
    try {
      await deferred.prompt()
      const choice = await deferred.userChoice
      if (choice.outcome === 'accepted') {
        set({ installed: true })
        return true
      }
    } catch {
      // The browser refused to show it (already shown, no user gesture):
      // treated as "not installed"; a later beforeinstallprompt re-arms it.
    }
    return false
  },

  dismissBanner: async (now = Date.now()) => {
    set({ bannerDismissedAt: now })
    try {
      await AsyncStorage.setItem(BANNER_DISMISSED_KEY, String(now))
    } catch {
      // best effort - see hydrate
    }
  },
}))

export function canInstall(state: Pick<PwaInstallState, 'deferred' | 'installed'>): boolean {
  return !state.installed && state.deferred !== null
}

// The first-visit banner: only while installable, never once dismissed or
// installed, and again only after the snooze window.
export function shouldShowInstallBanner(
  state: Pick<PwaInstallState, 'deferred' | 'installed' | 'bannerDismissedAt' | 'isHydrated'>,
  now: number = Date.now()
): boolean {
  if (!state.isHydrated || !canInstall(state)) return false
  if (state.bannerDismissedAt === null) return true
  return now - state.bannerDismissedAt > BANNER_SNOOZE_MS
}
