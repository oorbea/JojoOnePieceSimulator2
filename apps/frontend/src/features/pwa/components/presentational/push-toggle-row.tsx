import { Bell, BellOff } from '@tamagui/lucide-icons-2'
import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'

type Props = {
  enabled: boolean
  busy: boolean
  /** The browser blocks notifications for this site - the button cannot
   * change that, so it is replaced by an explanation. */
  blocked: boolean
  onToggle: () => void
}

// Profile row to turn game notifications on/off for this device.
export function PushToggleRow({ enabled, busy, blocked, onToggle }: Props) {
  const { t } = useTranslation()
  const Icon = enabled ? Bell : BellOff

  return (
    <YStack gap="$2">
      <XStack items="center" justify="space-between" gap="$3" flexWrap="wrap">
        <XStack items="center" gap="$3" flex={1} flexBasis={220}>
          <Icon size={22} color="$panelText" strokeWidth={2.5} />
          <YStack flex={1} gap="$1">
            <GlowText level="label">{t('pwa.push.toggle.label')}</GlowText>
            <GlowText level="label" tone="soft" fontSize="$2">
              {t('pwa.push.toggle.description')}
            </GlowText>
          </YStack>
        </XStack>
        {blocked ? null : (
          <GlossButton
            tone={enabled ? 'glass' : 'blue'}
            btnSize="sm"
            disabled={busy}
            onPress={onToggle}
            accessibilityLabel={
              enabled ? t('pwa.push.toggle.turnOffHint') : t('pwa.push.toggle.turnOnHint')
            }
          >
            {enabled ? t('pwa.push.toggle.turnOff') : t('pwa.push.toggle.turnOn')}
          </GlossButton>
        )}
      </XStack>
      {blocked ? (
        <GlowText level="label" tone="soft" fontSize="$2">
          {t('pwa.push.toggle.blocked')}
        </GlowText>
      ) : null}
    </YStack>
  )
}
