import { installBackGuard, type BackGuardWindow } from '../back-guard'

// A tiny history stack + popstate dispatcher: enough to model "push guard,
// press back, guard is re-pushed" without jsdom's async history.
function makeWindow() {
  const stack: unknown[] = [null] // the page itself
  let index = 0
  const listeners = new Set<(event: PopStateEvent) => void>()
  const win: BackGuardWindow = {
    history: {
      pushState: jest.fn((state: unknown) => {
        stack.splice(index + 1)
        stack.push(state)
        index += 1
      }),
      back: jest.fn(() => {
        if (index > 0) index -= 1
      }),
      get state() {
        return stack[index]
      },
    },
    addEventListener: (_type, l) => listeners.add(l),
    removeEventListener: (_type, l) => listeners.delete(l),
  }
  const firePopState = () =>
    listeners.forEach((l) => l({ state: stack[index] } as unknown as PopStateEvent))
  return {
    win,
    // The system back action: moves down the stack, then popstate fires.
    pressBack: () => {
      if (index > 0) index -= 1
      firePopState()
    },
    // Forward onto an existing entry.
    pressForward: () => {
      if (index < stack.length - 1) index += 1
      firePopState()
    },
    depth: () => stack.length,
    currentIsGuard: () => typeof stack[index] === 'object' && stack[index] !== null,
    listenerCount: () => listeners.size,
  }
}

describe('installBackGuard', () => {
  it('pushes a guard entry on install', () => {
    const env = makeWindow()
    installBackGuard(env.win, jest.fn())
    expect(env.depth()).toBe(2)
    expect(env.currentIsGuard()).toBe(true)
  })

  it('reports a back press and puts the guard back so the next one is caught too', () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    installBackGuard(env.win, onBackAttempt)

    env.pressBack()
    expect(onBackAttempt).toHaveBeenCalledTimes(1)
    expect(env.currentIsGuard()).toBe(true)

    env.pressBack()
    expect(onBackAttempt).toHaveBeenCalledTimes(2)
    expect(env.currentIsGuard()).toBe(true)
  })

  it('does not treat landing on the guard entry as a back press', () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    installBackGuard(env.win, onBackAttempt)
    // pressBack re-pushes the guard, truncating "forward" history, so build
    // the forward case by hand: step down without going through the handler.
    ;(env.win.history.back as jest.Mock)()
    env.pressForward()
    expect(onBackAttempt).not.toHaveBeenCalled()
  })

  it('uninstall stops listening and pops a guard it is still standing on', () => {
    const env = makeWindow()
    const uninstall = installBackGuard(env.win, jest.fn())

    uninstall()

    expect(env.listenerCount()).toBe(0)
    expect(env.win.history.back).toHaveBeenCalledTimes(1)
    expect(env.currentIsGuard()).toBe(false)
  })

  it('uninstall leaves the history alone when the page already navigated away', () => {
    const env = makeWindow()
    const uninstall = installBackGuard(env.win, jest.fn())
    // router.replace swapped the current (guard) entry for the next page's.
    env.win.history.pushState('router-entry', '')

    uninstall()

    expect(env.win.history.back).not.toHaveBeenCalled()
  })

  it('stops reporting after uninstall', () => {
    const env = makeWindow()
    const onBackAttempt = jest.fn()
    const uninstall = installBackGuard(env.win, onBackAttempt)
    uninstall()
    env.pressBack()
    expect(onBackAttempt).not.toHaveBeenCalled()
  })
})
