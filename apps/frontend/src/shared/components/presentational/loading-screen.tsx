import { useTranslation } from 'react-i18next'
import { Spinner } from 'tamagui'

import { GlowText } from './glow-text'
import { PageShell } from './page-shell'
import { WiiCard } from './wii-card'

// Replaces the bare `YStack + Spinner` that used to flash a plain
// background while the session store hydrates.
export function LoadingScreen() {
  const { t } = useTranslation()
  return (
    <PageShell align="center">
      <WiiCard
        tone="glass"
        padded
        items="center"
        gap="$3"
        width={180}
        transition="bouncy"
        enterStyle={{ scale: 0.85, opacity: 0 }}
      >
        <Spinner size="large" color="$channelActive" />
        <GlowText level="label">{t('common.loading')}…</GlowText>
      </WiiCard>
    </PageShell>
  )
}
