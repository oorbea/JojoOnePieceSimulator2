// Keeps the screen on while it's wanted (waiting in a lobby, playing a game):
// otherwise a phone dims and locks mid-sorteo while the player waits for
// others to vote. The Screen Wake Lock API gives the lock away whenever the
// page is hidden, so it has to be re-requested on every return to visible.
// Browser APIs come in through parameters so the behaviour is unit-testable.

export type WakeLockSentinelLike = {
  release: () => Promise<void>
  addEventListener: (type: 'release', listener: () => void) => void
}

export type WakeLockApi = {
  request: (type: 'screen') => Promise<WakeLockSentinelLike>
}

export type VisibilitySource = {
  readonly visibilityState: string
  addEventListener: (type: 'visibilitychange', listener: () => void) => void
  removeEventListener: (type: 'visibilitychange', listener: () => void) => void
}

export type WakeLock = {
  start: () => void
  stop: () => void
}

export function createWakeLock(api: WakeLockApi, doc: VisibilitySource): WakeLock {
  let sentinel: WakeLockSentinelLike | null = null
  let acquiring = false
  let stopped = true

  const acquire = async () => {
    if (stopped || acquiring || sentinel) return
    acquiring = true
    try {
      const lock = await api.request('screen')
      if (stopped) {
        // stop() ran while the request was in flight.
        void lock.release().catch(() => {})
        return
      }
      sentinel = lock
      lock.addEventListener('release', () => {
        if (sentinel === lock) sentinel = null
      })
    } catch {
      // Denied (battery saver, no user activation, document not visible):
      // the screen just behaves as it always did. The next visibility
      // change tries again.
    } finally {
      acquiring = false
    }
  }

  const onVisibilityChange = () => {
    if (doc.visibilityState === 'visible') void acquire()
  }

  return {
    start: () => {
      if (!stopped) return
      stopped = false
      doc.addEventListener('visibilitychange', onVisibilityChange)
      void acquire()
    },
    stop: () => {
      stopped = true
      doc.removeEventListener('visibilitychange', onVisibilityChange)
      const held = sentinel
      sentinel = null
      void held?.release().catch(() => {})
    },
  }
}
