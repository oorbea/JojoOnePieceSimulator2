import { WifiOff } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from './glass-panel'
import { GlowText } from './glow-text'

// Floating notice under the top bar while the device has no network. Purely
// informational (no action): the app reconnects on its own when the network
// returns, see use-game-socket.ts's online handler.
export function OfflineBanner() {
  const { t } = useTranslation()

  return (
    <GlassPanel tone="strong" rounded="$pill" px="$4" py="$2.5" maxW={520}>
      <XStack items="center" gap="$2">
        <WifiOff size={16} color="$panelTextSoft" />
        <YStack shrink={1}>
          <GlowText level="label">{t('pwa.offline.banner')}</GlowText>
        </YStack>
      </XStack>
    </GlassPanel>
  )
}
