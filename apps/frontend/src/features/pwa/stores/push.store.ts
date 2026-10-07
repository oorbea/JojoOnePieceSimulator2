import { create } from 'zustand'

import { AsyncStorage } from '@/shared/lib/async-storage'

import {
  getPushConfig,
  registerPushSubscription,
  removePushSubscription,
} from '../api/push.api'
import { webPushBrowser } from '../lib/push-browser'
import {
  PushEnableError,
  UNSUPPORTED_STATUS,
  disablePush,
  enablePush,
  loadPushStatus,
  type PushApi,
  type PushBrowser,
  type PushEnableFailure,
  type PushStatus,
} from '../lib/push-controller'

const PROMPT_DISMISSED_KEY = 'jops.push-prompt-dismissed'

export type PushStoreDeps = {
  browser: PushBrowser
  api: PushApi
}

export type PushState = PushStatus & {
  /** init() finished at least once - nothing about push is known before. */
  isLoaded: boolean
  /** An enable/disable is in flight. */
  busy: boolean
  /** The user answered "not now" to the lobby card; never ask again. */
  promptDismissed: boolean

  init: () => Promise<void>
  /** Resolves to null on success, else why it did not work. */
  enable: () => Promise<PushEnableFailure | null>
  disable: () => Promise<void>
  dismissPrompt: () => Promise<void>
  /** Forget everything account-specific (logout). */
  reset: () => void
}

export function createPushStore(deps: PushStoreDeps) {
  return create<PushState>((set, get) => ({
    ...UNSUPPORTED_STATUS,
    isLoaded: false,
    busy: false,
    promptDismissed: false,

    init: async () => {
      let promptDismissed = get().promptDismissed
      try {
        promptDismissed = (await AsyncStorage.getItem(PROMPT_DISMISSED_KEY)) === 'true'
      } catch {
        // Storage unavailable: the card just shows again next visit.
      }
      const snapshot = await loadPushStatus(deps.browser, deps.api)
      set({ ...snapshot, promptDismissed, isLoaded: true })
    },

    enable: async () => {
      if (get().busy) return null
      set({ busy: true })
      try {
        const result = await enablePush(deps.browser, deps.api, get().publicKey)
        set({ permission: result.permission, subscribed: true })
        return null
      } catch (error) {
        if (error instanceof PushEnableError) {
          if (error.permission !== 'unsupported') set({ permission: error.permission })
          return error.reason
        }
        return 'failed'
      } finally {
        set({ busy: false })
      }
    },

    disable: async () => {
      if (get().busy) return
      set({ busy: true })
      try {
        await disablePush(deps.browser, deps.api)
        set({ subscribed: false })
      } finally {
        set({ busy: false })
      }
    },

    dismissPrompt: async () => {
      set({ promptDismissed: true })
      try {
        await AsyncStorage.setItem(PROMPT_DISMISSED_KEY, 'true')
      } catch {
        // best effort - see init
      }
    },

    reset: () => set({ ...UNSUPPORTED_STATUS, isLoaded: false, busy: false }),
  }))
}

// The real browser + API, shared with the logout hook (see use-push-bootstrap.ts).
export const pushDeps: PushStoreDeps = {
  browser: webPushBrowser,
  api: { getConfig: getPushConfig, register: registerPushSubscription, remove: removePushSubscription },
}

export const usePushStore = createPushStore(pushDeps)

// The lobby card: only while push can work on this device and the user has
// neither enabled it, blocked it, nor said "not now".
export function shouldShowPushPrompt(
  state: Pick<
    PushState,
    'isLoaded' | 'supported' | 'serverEnabled' | 'subscribed' | 'permission' | 'promptDismissed'
  >
): boolean {
  return (
    state.isLoaded &&
    state.supported &&
    state.serverEnabled &&
    !state.subscribed &&
    state.permission !== 'denied' &&
    !state.promptDismissed
  )
}

// The profile toggle exists whenever push can work here at all, including
// after the user blocked it (it then explains how to unblock).
export function canTogglePush(
  state: Pick<PushState, 'isLoaded' | 'supported' | 'serverEnabled'>
): boolean {
  return state.isLoaded && state.supported && state.serverEnabled
}
