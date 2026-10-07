import { AsyncStorage } from '@/shared/lib/async-storage'

import {
  BANNER_SNOOZE_MS,
  canInstall,
  shouldShowInstallBanner,
  usePwaInstallStore,
  type BeforeInstallPromptEvent,
} from '../pwa-install.store'

function fakePrompt(outcome: 'accepted' | 'dismissed') {
  const event = new Event('beforeinstallprompt') as BeforeInstallPromptEvent
  event.prompt = jest.fn().mockResolvedValue(undefined)
  Object.defineProperty(event, 'userChoice', {
    value: Promise.resolve({ outcome, platform: 'web' }),
  })
  return event
}

describe('usePwaInstallStore', () => {
  beforeEach(async () => {
    await AsyncStorage.clear()
    usePwaInstallStore.setState({
      deferred: null,
      installed: false,
      bannerDismissedAt: null,
      isHydrated: false,
    })
  })

  it('promptInstall does nothing without a captured prompt', async () => {
    expect(await usePwaInstallStore.getState().promptInstall()).toBe(false)
  })

  it('an accepted prompt marks the app installed and consumes the event', async () => {
    const event = fakePrompt('accepted')
    usePwaInstallStore.getState().setDeferred(event)

    expect(await usePwaInstallStore.getState().promptInstall()).toBe(true)

    expect(event.prompt).toHaveBeenCalledTimes(1)
    const state = usePwaInstallStore.getState()
    expect(state.installed).toBe(true)
    expect(state.deferred).toBeNull()
  })

  it('a dismissed prompt consumes the event without marking installed', async () => {
    usePwaInstallStore.getState().setDeferred(fakePrompt('dismissed'))

    expect(await usePwaInstallStore.getState().promptInstall()).toBe(false)

    const state = usePwaInstallStore.getState()
    expect(state.installed).toBe(false)
    expect(state.deferred).toBeNull()
  })

  it('a prompt() that throws is treated as not installed', async () => {
    const event = fakePrompt('accepted')
    ;(event.prompt as jest.Mock).mockRejectedValue(new Error('no user gesture'))
    usePwaInstallStore.getState().setDeferred(event)

    expect(await usePwaInstallStore.getState().promptInstall()).toBe(false)
  })

  it('setInstalled(true) drops any captured prompt', () => {
    usePwaInstallStore.getState().setDeferred(fakePrompt('accepted'))
    usePwaInstallStore.getState().setInstalled(true)
    expect(usePwaInstallStore.getState().deferred).toBeNull()
  })

  it('dismissBanner persists and hydrate restores it', async () => {
    await usePwaInstallStore.getState().dismissBanner(1234)
    expect(await AsyncStorage.getItem('jops.pwa-install-banner-dismissed-at')).toBe('1234')

    usePwaInstallStore.setState({ bannerDismissedAt: null, isHydrated: false })
    await usePwaInstallStore.getState().hydrate()
    expect(usePwaInstallStore.getState().bannerDismissedAt).toBe(1234)
    expect(usePwaInstallStore.getState().isHydrated).toBe(true)
  })

  it('hydrate ignores a corrupt stored value', async () => {
    await AsyncStorage.setItem('jops.pwa-install-banner-dismissed-at', 'banana')
    await usePwaInstallStore.getState().hydrate()
    expect(usePwaInstallStore.getState().bannerDismissedAt).toBeNull()
  })
})

describe('canInstall / shouldShowInstallBanner', () => {
  const deferred = fakePrompt('accepted')
  const base = { deferred, installed: false, bannerDismissedAt: null, isHydrated: true }

  it('is installable only with a captured prompt and not yet installed', () => {
    expect(canInstall(base)).toBe(true)
    expect(canInstall({ ...base, deferred: null })).toBe(false)
    expect(canInstall({ ...base, installed: true })).toBe(false)
  })

  it('shows the banner on first visit', () => {
    expect(shouldShowInstallBanner(base)).toBe(true)
  })

  it('waits for hydration so a dismissed banner never flashes', () => {
    expect(shouldShowInstallBanner({ ...base, isHydrated: false })).toBe(false)
  })

  it('hides inside the snooze window and returns after it', () => {
    const dismissedAt = 1_000_000
    const state = { ...base, bannerDismissedAt: dismissedAt }
    expect(shouldShowInstallBanner(state, dismissedAt + BANNER_SNOOZE_MS - 1)).toBe(false)
    expect(shouldShowInstallBanner(state, dismissedAt + BANNER_SNOOZE_MS + 1)).toBe(true)
  })

  it('never shows when not installable', () => {
    expect(shouldShowInstallBanner({ ...base, deferred: null })).toBe(false)
  })
})
