import { registerLogoutHook, runLogoutHooks } from '../logout-hooks'

describe('logout hooks', () => {
  it('runs every registered hook', async () => {
    const a = jest.fn().mockResolvedValue(undefined)
    const b = jest.fn().mockResolvedValue(undefined)
    const offA = registerLogoutHook(a)
    const offB = registerLogoutHook(b)

    await runLogoutHooks()

    expect(a).toHaveBeenCalledTimes(1)
    expect(b).toHaveBeenCalledTimes(1)
    offA()
    offB()
  })

  it('stops running a hook once it is unregistered', async () => {
    const hook = jest.fn().mockResolvedValue(undefined)
    registerLogoutHook(hook)()

    await runLogoutHooks()

    expect(hook).not.toHaveBeenCalled()
  })

  it('a rejecting hook never breaks the others or the sign-out', async () => {
    const failing = jest.fn().mockRejectedValue(new Error('boom'))
    const fine = jest.fn().mockResolvedValue(undefined)
    const offFailing = registerLogoutHook(failing)
    const offFine = registerLogoutHook(fine)

    await expect(runLogoutHooks()).resolves.toBeUndefined()

    expect(fine).toHaveBeenCalledTimes(1)
    offFailing()
    offFine()
  })

  it('does not wait longer than the timeout for a hanging hook', async () => {
    const off = registerLogoutHook(() => new Promise<void>(() => {}))
    const start = Date.now()

    await runLogoutHooks(30)

    expect(Date.now() - start).toBeLessThan(1000)
    off()
  })

  it('ignores a re-entrant call (a hook whose request 401s back into logout)', async () => {
    const inner = jest.fn().mockResolvedValue(undefined)
    let reentered: Promise<void> | undefined
    const off = registerLogoutHook(async () => {
      registerLogoutHook(inner)
      reentered = runLogoutHooks()
      await reentered
    })

    await runLogoutHooks()

    expect(inner).not.toHaveBeenCalled()
    off()
  })
})
