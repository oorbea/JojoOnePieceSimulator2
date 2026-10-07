import { Smartphone } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from './glass-panel'
import { GlossButton } from './gloss-button'
import { GlowText } from './glow-text'

type Props = {
  onInstall: () => void
  onDismiss: () => void
}

// One-time nudge to install the app, floated above the mobile dock by
// AppShell. Deliberately small and dismissible: the permanent entry point is
// the install button in the top bar, this only makes the first visit aware
// that the game can live on the home screen.
export function InstallBanner({ onInstall, onDismiss }: Props) {
  const { t } = useTranslation()

  return (
    <GlassPanel tone="strong" rounded="$card" px="$4" py="$3" width="100%">
      <YStack gap="$3">
        <XStack items="center" gap="$3">
          <Smartphone size={28} color="$wiiBlue" strokeWidth={2.5} />
          <YStack flex={1} gap="$1">
            <GlowText level="heading" fontSize="$5">
              {t('pwa.banner.title')}
            </GlowText>
            <GlowText level="label" tone="soft">
              {t('pwa.banner.body')}
            </GlowText>
          </YStack>
        </XStack>
        <XStack gap="$2" justify="flex-end">
          <GlossButton
            tone="glass"
            btnSize="sm"
            onPress={onDismiss}
            accessibilityLabel={t('pwa.banner.dismissHint')}
          >
            {t('pwa.banner.dismiss')}
          </GlossButton>
          <GlossButton
            tone="blue"
            btnSize="sm"
            onPress={onInstall}
            accessibilityLabel={t('pwa.banner.installHint')}
          >
            {t('pwa.banner.install')}
          </GlossButton>
        </XStack>
      </YStack>
    </GlassPanel>
  )
}
