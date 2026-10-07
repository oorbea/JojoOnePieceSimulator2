import { useTranslation } from 'react-i18next'

import { AppError } from '@/shared/api/errors'
import { vibrate, vibrationSupported } from '@/shared/lib/haptics'
import { showErrorToast, showSuccessToast } from '@/shared/lib/toast'

import { HapticsTestRow } from '../presentational/haptics-test-row'

// "Try vibration" in the profile. Renders nothing off phones (no coarse
// pointer / no Vibration API), where there is nothing to try.
export function HapticsTestContainer() {
  const { t } = useTranslation()

  if (!vibrationSupported()) return null

  const onTest = () => {
    // The browser answers false when it refused to vibrate at all.
    if (vibrate('test')) showSuccessToast(t('pwa.haptics.tried'))
    else showErrorToast(new AppError(t('pwa.haptics.refused')))
  }

  return <HapticsTestRow onTest={onTest} />
}
