import {
  GUARD_DEPTH,
  guardLevelOf,
  installBackGuard,
  type BackGuardWindow,
} from '../back-guard'

// A history stack + event dispatcher, synchronous so tests can reason about
// order: `pressBack` moves down one entry then fires popstate, `go(n)` does the
// same for n entries.
function makeWindow() {
  const stack: unknown[] = [null] // the page's own entry
  let index = 0
  const listeners: Record<string, Set<(event: unknown) => void>> = {}
  const fire = (type: string, event: unknown = {}) =>
    [...(listeners[type] ?? [])].forEach((l) => l(event))

  const win: BackGuardWindow = {
    history: {
      pushState: jest.fn((state: unknown) => {
        stack.splice(index + 1)
        stack.push(state)
        index += 1
      }),
      go: jest.fn((delta: number) => {
        index = Math.max(0, Math.min(stack.length - 1, index + delta))
        fire('popstate', { state: stack[index] })
      }),
      get state() {
        return stack[index]
      },
    },
    addEventListener: ((type: string, l: (e: unknown) => void) => {
      ;(listeners[type] ??= new Set()).add(l)
    }) as BackGuardWindow['addEventListener'],
    removeEventListener: ((type: string, l: (e: unknown) => void) => {
      listeners[type]?.delete(l)
    }) as BackGuardWindow['removeEventListener'],
  }

  return {
    win,
    pressBack: () => {
      if (index > 0) index -= 1
      fire('popstate', { state: stack[index] })
    },
    pressForward: () => {
      if (index < stack.length - 1) index += 1
      fire('popstate', { state: stack[index] })
    },
    tap: () => fire('pointerup'),
    key: () => fire('keydown'),
    depth: () => stack.length,
    currentLevel: () => guardLevelOf(stack[index]),
    listenerCount: (type: string) => listeners[type]?.size ?? 0,
  }
}

describe('installBackGuard', () => {
  it('pushes GUARD_DEPTH guard entries on install', () => {
    const env = makeWindow()
    installBackGuard(env.win, jest.fn())
    expect(env.depth()).toBe(1 + GUARD_DEPTH)
    expect(env.currentLevel()).toBe(GUARD_DEPTH)
  })

  it('catches every back press in a row, even without any tap in between', () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    installBackGuard(env.win, onBackAttempt)

    // More presses than guard entries: the one that lands on the page's own
    // entry is still reported, so the player is asked each time.
    for (let i = 1; i <= GUARD_DEPTH; i++) {
      env.pressBack()
      expect(onBackAttempt).toHaveBeenCalledTimes(i)
    }
    expect(env.currentLevel()).toBe(0)
  })

  it('refills the stack from a tap, so the next press is caught again', () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    installBackGuard(env.win, onBackAttempt)

    env.pressBack()
    env.pressBack()
    expect(env.currentLevel()).toBe(GUARD_DEPTH - 2)

    // e.g. the player taps "Cancelar" on the confirmation sheet
    env.tap()
    expect(env.currentLevel()).toBe(GUARD_DEPTH)

    env.pressBack()
    expect(onBackAttempt).toHaveBeenCalledTimes(3)
  })

  it('refills from a key press too', () => {
    const env = makeWindow()
    installBackGuard(env.win, jest.fn())
    env.pressBack()
    env.key()
    expect(env.currentLevel()).toBe(GUARD_DEPTH)
  })

  it('does not push anything on a tap while the stack is full', () => {
    const env = makeWindow()
    installBackGuard(env.win, jest.fn())
    const pushes = (env.win.history.pushState as jest.Mock).mock.calls.length

    env.tap()
    env.tap()

    expect((env.win.history.pushState as jest.Mock).mock.calls.length).toBe(pushes)
  })

  it('does not treat moving forward onto a guard entry as a back press', () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    installBackGuard(env.win, onBackAttempt)
    env.pressBack()
    onBackAttempt.mockClear()

    env.pressForward()

    expect(onBackAttempt).not.toHaveBeenCalled()
  })

  it('never pushes entries from inside the popstate handler (Chrome would skip them)', () => {
    const env = makeWindow()
    installBackGuard(env.win, jest.fn())
    const pushes = (env.win.history.pushState as jest.Mock).mock.calls.length

    env.pressBack()
    env.pressBack()

    expect((env.win.history.pushState as jest.Mock).mock.calls.length).toBe(pushes)
  })
})

describe('release', () => {
  it('walks back to the page entry and resolves once popstate fired', async () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    const guard = installBackGuard(env.win, onBackAttempt)

    await guard.release()

    expect(env.win.history.go).toHaveBeenCalledWith(-GUARD_DEPTH)
    expect(env.currentLevel()).toBe(0)
    // Our own go() must not be reported as the player pressing back.
    expect(onBackAttempt).not.toHaveBeenCalled()
  })

  it('only walks back as far as the guard entries that are left', async () => {
    const env = makeWindow()
    const guard = installBackGuard(env.win, jest.fn())
    env.pressBack() // level GUARD_DEPTH - 1

    await guard.release()

    expect(env.win.history.go).toHaveBeenCalledWith(-(GUARD_DEPTH - 1))
  })

  it('resolves at once when the page entry is already current', async () => {
    const env = makeWindow()
    const guard = installBackGuard(env.win, jest.fn())
    for (let i = 0; i < GUARD_DEPTH; i++) env.pressBack()

    await guard.release()

    expect(env.win.history.go).not.toHaveBeenCalled()
  })

  it('stops listening and is idempotent', async () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    const guard = installBackGuard(env.win, onBackAttempt)

    await guard.release()
    await guard.release()

    expect(env.listenerCount('popstate')).toBe(0)
    expect(env.listenerCount('pointerup')).toBe(0)
    expect(env.listenerCount('keydown')).toBe(0)
    expect(env.win.history.go).toHaveBeenCalledTimes(1)

    env.tap()
    env.pressBack()
    expect(onBackAttempt).not.toHaveBeenCalled()
  })

  it('resolves even if the browser never fires popstate', async () => {
    jest.useFakeTimers()
    try {
      const env = makeWindow()
      ;(env.win.history.go as jest.Mock).mockImplementation(() => {})
      const guard = installBackGuard(env.win, jest.fn())

      const released = guard.release()
      jest.advanceTimersByTime(400)

      await expect(released).resolves.toBeUndefined()
    } finally {
      jest.useRealTimers()
    }
  })
})

describe('guardLevelOf', () => {
  it('reads the level from a guard state and treats anything else as level 0', () => {
    expect(guardLevelOf({ jopsBackGuard: 2 })).toBe(2)
    expect(guardLevelOf({ jopsBackGuard: true })).toBe(1)
    expect(guardLevelOf({ id: 'router-entry' })).toBe(0)
    expect(guardLevelOf(null)).toBe(0)
    expect(guardLevelOf('x')).toBe(0)
  })
})
