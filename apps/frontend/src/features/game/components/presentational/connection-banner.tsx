import { WifiOff } from '@tamagui/lucide-icons-2'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { XStack } from 'tamagui'

import type { SocketStatus } from '@/features/game/stores/game-socket.store'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  status: SocketStatus
  nextRetryAt: number | null
  onRetryNow: () => void
}

export function ConnectionBanner({ status, nextRetryAt, onRetryNow }: Props) {
  const { t } = useTranslation()
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (status !== 'reconnecting') return
    const id = setInterval(() => setNow(Date.now()), 500)
    return () => clearInterval(id)
  }, [status])

  if (status === 'open' || status === 'idle') return null

  const seconds = nextRetryAt ? Math.max(0, Math.ceil((nextRetryAt - now) / 1000)) : 0

  return (
    <GlassPanel tone="strong" rounded="$pill" px="$4" py="$2.5" width="100%">
      <XStack items="center" gap="$2" justify="space-between" flexWrap="wrap">
        <XStack items="center" gap="$2">
          <WifiOff size={16} color="$panelTextSoft" />
          <GlowText level="label">
            {status === 'unavailable'
              ? t('game.connection.unavailable')
              : t('game.connection.reconnecting', { seconds })}
          </GlowText>
        </XStack>
        <GlossButton tone="glass" btnSize="sm" onPress={onRetryNow} accessibilityLabel={t('game.connection.retryNow')}>
          {t('game.connection.retryNow')}
        </GlossButton>
      </XStack>
    </GlassPanel>
  )
}
