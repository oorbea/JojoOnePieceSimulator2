import { useTranslation } from 'react-i18next'

import { AppError } from '@/shared/api/errors'
import { showErrorToast, showSuccessToast } from '@/shared/lib/toast'

import { shouldShowPushPrompt, usePushStore } from '../../stores/push.store'
import { PushPromptCard } from '../presentational/push-prompt-card'

// Offers push notifications in a lobby, once. Renders nothing unless push can
// work on this device and the player has not decided yet.
export function PushPromptContainer() {
  const { t } = useTranslation()
  const show = usePushStore((state) => shouldShowPushPrompt(state))
  const busy = usePushStore((state) => state.busy)
  const enable = usePushStore((state) => state.enable)
  const dismissPrompt = usePushStore((state) => state.dismissPrompt)

  if (!show) return null

  const onEnable = async () => {
    const failure = await enable()
    if (failure === null) {
      showSuccessToast(t('pwa.push.enabled'))
    } else if (failure === 'denied') {
      showErrorToast(new AppError(t('pwa.push.denied')))
    } else if (failure === 'failed') {
      showErrorToast(new AppError(t('pwa.push.failed')))
    }
    // 'dismissed': the player closed the browser prompt - say nothing, the
    // card stays so they can still choose.
  }

  return (
    <PushPromptCard
      busy={busy}
      onEnable={() => void onEnable()}
      onDismiss={() => void dismissPrompt()}
    />
  )
}
