import { BellRing } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  busy: boolean
  onEnable: () => void
  onDismiss: () => void
}

// The one-time ask, shown in a lobby: that is the moment a notification is
// obviously useful (you are about to wait for others), which is also what
// browsers reward - a permission prompt out of nowhere gets denied and then
// can never be asked again.
export function PushPromptCard({ busy, onEnable, onDismiss }: Props) {
  const { t } = useTranslation()

  return (
    <GlassPanel tone="strong" rounded="$card" px="$4" py="$3" width="100%">
      <YStack gap="$3">
        <XStack items="center" gap="$3">
          <BellRing size={28} color="$wiiBlue" strokeWidth={2.5} />
          <YStack flex={1} gap="$1">
            <GlowText level="heading" fontSize="$5">
              {t('pwa.push.prompt.title')}
            </GlowText>
            <GlowText level="label" tone="soft">
              {t('pwa.push.prompt.body')}
            </GlowText>
          </YStack>
        </XStack>
        <XStack gap="$2" justify="flex-end">
          <GlossButton
            tone="glass"
            btnSize="sm"
            disabled={busy}
            onPress={onDismiss}
            accessibilityLabel={t('pwa.push.prompt.dismissHint')}
          >
            {t('pwa.push.prompt.dismiss')}
          </GlossButton>
          <GlossButton
            tone="blue"
            btnSize="sm"
            disabled={busy}
            onPress={onEnable}
            accessibilityLabel={t('pwa.push.prompt.enableHint')}
          >
            {t('pwa.push.prompt.enable')}
          </GlossButton>
        </XStack>
      </YStack>
    </GlassPanel>
  )
}
