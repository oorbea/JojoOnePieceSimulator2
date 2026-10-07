import type { GameActivity } from '@/shared/stores/game-activity.store'

// Coordinates "a new build was deployed" -> "the installed PWA is running it".
//
// public/sw.js is stamped with the build id at image build time, so every
// deploy yields a byte-different service worker. The browser installs it in
// the background and leaves it `waiting` (sw.js deliberately does NOT
// skipWaiting on its own - activating mid-session would swap the cache under
// a page still running the old bundle). This controller decides *when* to
// activate it and reload:
//
//   - never while a game is in progress (a reload would drop the player out
//     of a live round),
//   - immediately otherwise (lobby / menus: the socket reconnects by itself),
//   - deferred updates apply the moment the game activity leaves "playing".
//
// All browser APIs arrive through `deps`, so the whole flow is unit-testable
// without a real service worker.

export const SW_URL = '/sw.js'
export const UPDATE_CHECK_INTERVAL_MS = 15 * 60 * 1000
// A reload this soon after the previous one means something is wrong (a
// controllerchange storm); stop instead of looping the user forever.
export const RELOAD_GUARD_WINDOW_MS = 30_000
const RELOAD_GUARD_KEY = 'jops.sw-reload-at'

export type WorkerLike = {
  state: string
  postMessage: (message: unknown) => void
  addEventListener: (type: 'statechange', listener: () => void) => void
}

export type RegistrationLike = {
  waiting: WorkerLike | null
  installing: WorkerLike | null
  update: () => Promise<unknown>
  addEventListener: (type: 'updatefound', listener: () => void) => void
}

export type ServiceWorkerContainerLike = {
  controller: unknown
  register: (url: string, options?: { updateViaCache?: 'none' }) => Promise<RegistrationLike>
  addEventListener: (type: 'controllerchange', listener: () => void) => void
  removeEventListener: (type: 'controllerchange', listener: () => void) => void
}

export type ReloadGuard = {
  canReload: () => boolean
  markReload: () => void
}

export type UpdateControllerDeps = {
  container: ServiceWorkerContainerLike
  getActivity: () => GameActivity
  reload: () => void
  guard: ReloadGuard
}

export type UpdateController = {
  start: () => Promise<void>
  // Call whenever the game activity may have changed (store subscription,
  // tab visibility) so a deferred update gets applied as soon as it is safe.
  flush: () => void
  checkForUpdate: () => void
  dispose: () => void
}

type StorageLike = Pick<Storage, 'getItem' | 'setItem'>

// sessionStorage-backed guard. Every access is wrapped: storage can be
// missing or throw (private windows, blocked site data), and the guard must
// never be the thing that breaks updates - without storage it simply allows
// the reload.
export function createReloadGuard(
  storage: StorageLike | undefined,
  now: () => number = Date.now
): ReloadGuard {
  return {
    canReload: () => {
      try {
        const raw = storage?.getItem(RELOAD_GUARD_KEY)
        if (!raw) return true
        const last = Number(raw)
        return !Number.isFinite(last) || now() - last > RELOAD_GUARD_WINDOW_MS
      } catch {
        return true
      }
    },
    markReload: () => {
      try {
        storage?.setItem(RELOAD_GUARD_KEY, String(now()))
      } catch {
        // see above - best effort only
      }
    },
  }
}

export function createUpdateController(deps: UpdateControllerDeps): UpdateController {
  const { container, getActivity, reload, guard } = deps

  let registration: RegistrationLike | null = null
  // True when a worker is waiting to be activated.
  let updateReady = false
  // True once the new worker took control; the page still runs the old
  // bundle and needs a reload.
  let reloadPending = false
  let hadController = !!container.controller
  let disposed = false

  const flush = () => {
    if (disposed || getActivity() === 'playing') return

    if (reloadPending) {
      reloadPending = false
      if (guard.canReload()) {
        guard.markReload()
        reload()
      }
      return
    }

    if (updateReady && registration?.waiting) {
      updateReady = false
      registration.waiting.postMessage({ type: 'SKIP_WAITING' })
    }
  }

  const onControllerChange = () => {
    // The very first install claims the page (clients.claim) without any
    // previous version having run - nothing stale to reload away from.
    if (!hadController) {
      hadController = true
      return
    }
    reloadPending = true
    flush()
  }

  const onUpdateFound = () => {
    const worker = registration?.installing
    if (!worker) return
    worker.addEventListener('statechange', () => {
      // `controller` set => this is an update to a running app, not the
      // first-ever install (which activates on its own).
      if (worker.state === 'installed' && container.controller) {
        updateReady = true
        flush()
      }
    })
  }

  return {
    start: async () => {
      container.addEventListener('controllerchange', onControllerChange)
      registration = await container.register(SW_URL, { updateViaCache: 'none' })
      if (disposed) return
      registration.addEventListener('updatefound', onUpdateFound)
      if (registration.waiting && container.controller) {
        updateReady = true
        flush()
      }
    },
    flush,
    checkForUpdate: () => {
      registration?.update().catch(() => {
        // Offline / transient network error: the next check retries.
      })
    },
    dispose: () => {
      disposed = true
      container.removeEventListener('controllerchange', onControllerChange)
    },
  }
}
