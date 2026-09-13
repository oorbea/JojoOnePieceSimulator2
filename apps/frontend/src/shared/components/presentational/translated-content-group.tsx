import { useTranslation } from 'react-i18next'
import { XStack, YStack } from 'tamagui'

import { LOCALE_ENDONYMS } from '@/shared/i18n'
import { a11yProps } from '@/shared/lib/a11y'
import type { Locale } from '@/shared/contracts/enums'

import { GlassPanel } from './glass-panel'
import { GlowText } from './glow-text'

type Props = {
  locale: Locale
  children: React.ReactNode
}

// Wraps the LocaleTabs + translated-field block in the Stand/Devil
// Fruit/Stage/Character admin forms with a header pill naming the active
// locale, so scrolling past LocaleTabs itself never loses sight of which
// language Description/Skills belong to (see ObsidianVault/
// i18n-multi-language.md's admin multi-locale form section). Purely
// presentational - `locale` is the container's existing activeLocale
// state, nothing new to wire up.
export function TranslatedContentGroup({ locale, children }: Props) {
  const { t } = useTranslation()
  const endonym = LOCALE_ENDONYMS[locale]

  return (
    <GlassPanel tone="plastic" rounded="$card" elevate={0} p="$3" gap="$3">
      <XStack items="center" justify="space-between" gap="$2">
        <GlowText level="label" tone="soft">
          {t('locale.translatedContent')}
        </GlowText>
        <GlassPanel
          tone="plastic"
          px="$2.5"
          py="$1"
          rounded="$pill"
          elevate={0}
          {...a11yProps(endonym, 'text')}
        >
          <GlowText level="label">{endonym}</GlowText>
        </GlassPanel>
      </XStack>
      <YStack gap="$4">{children}</YStack>
    </GlassPanel>
  )
}
