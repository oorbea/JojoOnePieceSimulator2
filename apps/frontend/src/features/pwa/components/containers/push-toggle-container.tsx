import { useTranslation } from 'react-i18next'

import { AppError } from '@/shared/api/errors'
import { showErrorToast, showSuccessToast } from '@/shared/lib/toast'

import { canTogglePush, usePushStore } from '../../stores/push.store'
import { PushToggleRow } from '../presentational/push-toggle-row'

// The profile's notifications switch. Renders nothing where push cannot work
// (unsupported browser, server without VAPID keys).
export function PushToggleContainer() {
  const { t } = useTranslation()
  const available = usePushStore((state) => canTogglePush(state))
  const subscribed = usePushStore((state) => state.subscribed)
  const blocked = usePushStore((state) => state.permission === 'denied')
  const busy = usePushStore((state) => state.busy)
  const enable = usePushStore((state) => state.enable)
  const disable = usePushStore((state) => state.disable)

  if (!available) return null

  const onToggle = async () => {
    if (subscribed) {
      await disable()
      return
    }
    const failure = await enable()
    if (failure === null) showSuccessToast(t('pwa.push.enabled'))
    else if (failure === 'denied') showErrorToast(new AppError(t('pwa.push.denied')))
    else if (failure === 'failed') showErrorToast(new AppError(t('pwa.push.failed')))
  }

  return (
    <PushToggleRow
      enabled={subscribed}
      busy={busy}
      blocked={blocked && !subscribed}
      onToggle={() => void onToggle()}
    />
  )
}
