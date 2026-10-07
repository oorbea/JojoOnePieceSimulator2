import { useExitGuardStore, type ExitGuardRegistration } from '../exit-guard.store'

function makeRegistration() {
  const order: string[] = []
  const registration: ExitGuardRegistration = {
    leaveGame: jest.fn(() => {
      order.push('leave')
    }),
    releaseHistory: jest.fn(async () => {
      order.push('release')
    }),
  }
  return { registration, order }
}

describe('useExitGuardStore', () => {
  beforeEach(() => {
    useExitGuardStore.setState({ registration: null, pending: null })
  })

  it('runs the exit straight away when no game is being protected', () => {
    const proceed = jest.fn()

    useExitGuardStore.getState().request(proceed)

    expect(proceed).toHaveBeenCalledTimes(1)
    expect(useExitGuardStore.getState().pending).toBeNull()
  })

  it('parks the exit and asks while a game is protected', () => {
    useExitGuardStore.getState().register(makeRegistration().registration)
    const proceed = jest.fn()

    useExitGuardStore.getState().request(proceed)

    expect(proceed).not.toHaveBeenCalled()
    expect(useExitGuardStore.getState().pending).not.toBeNull()
  })

  it('confirm releases history, then leaves the game, then navigates - in that order', async () => {
    const { registration, order } = makeRegistration()
    useExitGuardStore.getState().register(registration)
    const proceed = jest.fn(() => {
      order.push('proceed')
    })
    useExitGuardStore.getState().request(proceed)

    await useExitGuardStore.getState().confirm()

    expect(order).toEqual(['release', 'leave', 'proceed'])
    expect(useExitGuardStore.getState().pending).toBeNull()
  })

  it('cancel drops the exit and touches nothing', async () => {
    const { registration } = makeRegistration()
    useExitGuardStore.getState().register(registration)
    const proceed = jest.fn()
    useExitGuardStore.getState().request(proceed)

    useExitGuardStore.getState().cancel()
    await useExitGuardStore.getState().confirm()

    expect(proceed).not.toHaveBeenCalled()
    expect(registration.leaveGame).not.toHaveBeenCalled()
    expect(registration.releaseHistory).not.toHaveBeenCalled()
  })

  it('a second request replaces the first (a repeated back press)', async () => {
    useExitGuardStore.getState().register(makeRegistration().registration)
    const first = jest.fn()
    const second = jest.fn()
    useExitGuardStore.getState().request(first)
    useExitGuardStore.getState().request(second)

    await useExitGuardStore.getState().confirm()

    expect(first).not.toHaveBeenCalled()
    expect(second).toHaveBeenCalledTimes(1)
  })

  it('unregistering clears a question that was still open', () => {
    useExitGuardStore.getState().register(makeRegistration().registration)
    useExitGuardStore.getState().request(jest.fn())

    useExitGuardStore.getState().register(null)

    expect(useExitGuardStore.getState().pending).toBeNull()
    expect(useExitGuardStore.getState().registration).toBeNull()
  })
})
