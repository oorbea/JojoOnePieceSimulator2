import type { GameActivity } from '@/shared/stores/game-activity.store'

import {
  RELOAD_GUARD_WINDOW_MS,
  createReloadGuard,
  createUpdateController,
  type RegistrationLike,
  type ServiceWorkerContainerLike,
  type WorkerLike,
} from '../sw-update'

function makeWorker(state = 'installing') {
  const listeners: (() => void)[] = []
  const worker = {
    state,
    postMessage: jest.fn(),
    addEventListener: (_type: 'statechange', listener: () => void) => {
      listeners.push(listener)
    },
  } satisfies WorkerLike
  return {
    worker,
    setState: (next: string) => {
      worker.state = next
      listeners.forEach((l) => l())
    },
  }
}

function makeEnv(options: { controlled?: boolean; waiting?: WorkerLike | null } = {}) {
  const updateFoundListeners: (() => void)[] = []
  const registration: RegistrationLike & { installing: WorkerLike | null } = {
    waiting: options.waiting ?? null,
    installing: null,
    update: jest.fn().mockResolvedValue(undefined),
    addEventListener: (_type, listener) => {
      updateFoundListeners.push(listener)
    },
  }
  const controllerListeners = new Set<() => void>()
  const container: ServiceWorkerContainerLike = {
    controller: options.controlled === false ? null : {},
    register: jest.fn().mockResolvedValue(registration),
    addEventListener: (_type, listener) => controllerListeners.add(listener),
    removeEventListener: (_type, listener) => controllerListeners.delete(listener),
  }
  let activity: GameActivity = 'none'
  const reload = jest.fn()
  const guard = { canReload: jest.fn(() => true), markReload: jest.fn() }
  const controller = createUpdateController({
    container,
    getActivity: () => activity,
    reload,
    guard,
  })
  return {
    controller,
    container,
    registration,
    reload,
    guard,
    setActivity: (next: GameActivity) => {
      activity = next
    },
    // A new worker appears and finishes installing while the page is open.
    installUpdate: () => {
      const installing = makeWorker()
      registration.installing = installing.worker
      registration.waiting = installing.worker
      updateFoundListeners.forEach((l) => l())
      installing.setState('installed')
      return installing.worker
    },
    fireControllerChange: () => controllerListeners.forEach((l) => l()),
    controllerListenerCount: () => controllerListeners.size,
  }
}

describe('createUpdateController', () => {
  it('registers the worker bypassing the HTTP cache for the script', async () => {
    const env = makeEnv()
    await env.controller.start()
    expect(env.container.register).toHaveBeenCalledWith('/sw.js', { updateViaCache: 'none' })
  })

  it('activates an installed update right away outside a game, then reloads on takeover', async () => {
    const env = makeEnv()
    await env.controller.start()

    const worker = env.installUpdate()
    expect(worker.postMessage).toHaveBeenCalledWith({ type: 'SKIP_WAITING' })
    expect(env.reload).not.toHaveBeenCalled()

    env.fireControllerChange()
    expect(env.reload).toHaveBeenCalledTimes(1)
    expect(env.guard.markReload).toHaveBeenCalledTimes(1)
  })

  it('applies an update that was already waiting when the page loaded', async () => {
    const waiting = makeWorker('installed').worker
    const env = makeEnv({ waiting })
    await env.controller.start()
    expect(waiting.postMessage).toHaveBeenCalledWith({ type: 'SKIP_WAITING' })
  })

  it('holds the update while a game is in progress and applies it once it ends', async () => {
    const env = makeEnv()
    env.setActivity('playing')
    await env.controller.start()

    const worker = env.installUpdate()
    expect(worker.postMessage).not.toHaveBeenCalled()

    env.setActivity('none')
    env.controller.flush()
    expect(worker.postMessage).toHaveBeenCalledWith({ type: 'SKIP_WAITING' })
  })

  it('applies updates in the lobby (only a started game blocks them)', async () => {
    const env = makeEnv()
    env.setActivity('lobby')
    await env.controller.start()
    const worker = env.installUpdate()
    expect(worker.postMessage).toHaveBeenCalledTimes(1)
  })

  it('does not reload a tab that is playing when another tab activated the update', async () => {
    const env = makeEnv()
    env.setActivity('playing')
    await env.controller.start()

    env.fireControllerChange()
    expect(env.reload).not.toHaveBeenCalled()

    env.setActivity('none')
    env.controller.flush()
    expect(env.reload).toHaveBeenCalledTimes(1)
  })

  it('ignores the first-ever install (no previous controller, nothing stale)', async () => {
    const env = makeEnv({ controlled: false })
    await env.controller.start()

    // First install: the worker installs, activates on its own and claims.
    const installing = makeWorker()
    env.registration.installing = installing.worker
    installing.setState('installed')
    env.fireControllerChange()

    expect(installing.worker.postMessage).not.toHaveBeenCalled()
    expect(env.reload).not.toHaveBeenCalled()
  })

  it('does not reload again within the guard window', async () => {
    const env = makeEnv()
    env.guard.canReload.mockReturnValue(false)
    await env.controller.start()
    env.installUpdate()
    env.fireControllerChange()
    expect(env.reload).not.toHaveBeenCalled()
    expect(env.guard.markReload).not.toHaveBeenCalled()
  })

  it('checkForUpdate asks the registration and swallows offline failures', async () => {
    const env = makeEnv()
    ;(env.registration.update as jest.Mock).mockRejectedValue(new Error('offline'))
    await env.controller.start()
    expect(() => env.controller.checkForUpdate()).not.toThrow()
    await Promise.resolve()
    expect(env.registration.update).toHaveBeenCalledTimes(1)
  })

  it('dispose stops reacting to takeovers', async () => {
    const env = makeEnv()
    await env.controller.start()
    env.controller.dispose()
    expect(env.controllerListenerCount()).toBe(0)
    env.fireControllerChange()
    expect(env.reload).not.toHaveBeenCalled()
  })
})

describe('createReloadGuard', () => {
  function memoryStorage(initial: Record<string, string> = {}) {
    const data = { ...initial }
    return {
      getItem: (key: string) => data[key] ?? null,
      setItem: (key: string, value: string) => {
        data[key] = value
      },
    }
  }

  it('allows the first reload, then blocks inside the window, then allows again', () => {
    let now = 1_000_000
    const guard = createReloadGuard(memoryStorage(), () => now)

    expect(guard.canReload()).toBe(true)
    guard.markReload()
    expect(guard.canReload()).toBe(false)

    now += RELOAD_GUARD_WINDOW_MS - 1
    expect(guard.canReload()).toBe(false)
    now += 2
    expect(guard.canReload()).toBe(true)
  })

  it('allows reloads when storage is unavailable or throws', () => {
    expect(createReloadGuard(undefined).canReload()).toBe(true)
    const throwing = {
      getItem: () => {
        throw new Error('blocked')
      },
      setItem: () => {
        throw new Error('blocked')
      },
    }
    const guard = createReloadGuard(throwing)
    expect(guard.canReload()).toBe(true)
    expect(() => guard.markReload()).not.toThrow()
  })
})
