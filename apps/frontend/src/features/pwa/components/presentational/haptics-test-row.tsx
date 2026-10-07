import { Vibrate } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  onTest: () => void
}

// Profile row to try the phone's vibration. The game buzzes at key moments;
// when it seems not to, this separates "the app never asked" from "the phone
// ignores the browser" (system vibration/haptics off, battery saver...).
export function HapticsTestRow({ onTest }: Props) {
  const { t } = useTranslation()

  return (
    <XStack items="center" justify="space-between" gap="$3" flexWrap="wrap">
      <XStack items="center" gap="$3" flex={1} flexBasis={220}>
        <Vibrate size={22} color="$panelText" strokeWidth={2.5} />
        <YStack flex={1} gap="$1">
          <GlowText level="label">{t('pwa.haptics.label')}</GlowText>
          <GlowText level="label" tone="soft" fontSize="$2">
            {t('pwa.haptics.description')}
          </GlowText>
        </YStack>
      </XStack>
      <GlossButton
        tone="glass"
        btnSize="sm"
        onPress={onTest}
        accessibilityLabel={t('pwa.haptics.testHint')}
      >
        {t('pwa.haptics.test')}
      </GlossButton>
    </XStack>
  )
}
