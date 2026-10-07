import { useEffect } from 'react'

import { registerLogoutHook } from '@/shared/lib/logout-hooks'

import { disablePush } from '../lib/push-controller'
import { pushDeps, usePushStore } from '../stores/push.store'

// Loads push availability once the user is signed in, and arranges for this
// device's subscription to be dropped when they sign out - so notifications
// for one account never keep landing on a phone someone else may pick up.
export function usePushBootstrap(signedIn: boolean) {
  useEffect(() => {
    if (!signedIn) return

    void usePushStore.getState().init()
    return registerLogoutHook(async () => {
      await disablePush(pushDeps.browser, pushDeps.api)
      usePushStore.getState().reset()
    })
  }, [signedIn])
}
