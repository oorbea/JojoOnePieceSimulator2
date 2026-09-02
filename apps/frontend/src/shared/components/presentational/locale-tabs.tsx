import { XStack, YStack } from 'tamagui'

import { LOCALE_ENDONYMS, SUPPORTED_LOCALES } from '@/shared/i18n'
import { a11yProps } from '@/shared/lib/a11y'
import type { Locale } from '@/shared/contracts/enums'

import { ChannelBarItem } from './channel-bar'
import { GlowText } from './glow-text'

type LocaleTabsProps = {
  value: Locale
  onChange: (locale: Locale) => void
  // A single Locale for Stand/Devil Fruit's "only en-GB is mandatory" rule;
  // an array for Stage's "every locale is mandatory" rule (see the vault's
  // game-stage-content.md) - pass SUPPORTED_LOCALES to star every tab.
  requiredLocale: Locale | Locale[]
  localesWithErrors?: Locale[]
  requiredLabel: string
  errorLabel: string
}

// Locale switcher for the Stand/Devil Fruit/Stage admin forms - one pill per
// supported locale, reusing ChannelBarItem's existing active/press styling
// instead of a new glass recipe. A mandatory locale never shows an error dot
// on its own - a real validation error there gets its own dot like any
// other locale. The error state is never color-only: it's also spelled out
// in accessibilityLabel (errorLabel), per the project's touch/a11y
// guidelines (visual-only error indication is an anti-pattern).
export function LocaleTabs({
  value,
  onChange,
  requiredLocale,
  localesWithErrors = [],
  requiredLabel,
  errorLabel,
}: LocaleTabsProps) {
  return (
    <XStack gap="$2" flexWrap="wrap">
      {SUPPORTED_LOCALES.map((locale) => {
        const isActive = locale === value
        const isRequired = Array.isArray(requiredLocale)
          ? requiredLocale.includes(locale)
          : locale === requiredLocale
        const hasError = localesWithErrors.includes(locale)
        const label = LOCALE_ENDONYMS[locale]
        const a11yLabel = [label, isRequired && requiredLabel, hasError && errorLabel]
          .filter(Boolean)
          .join(', ')

        return (
          <ChannelBarItem
            key={locale}
            active={isActive}
            onPress={() => onChange(locale)}
            position="relative"
            {...a11yProps(a11yLabel, 'tab', { disabled: false })}
          >
            <GlowText level="label" tone={isActive ? 'onColor' : 'ink'}>
              {label}
              {isRequired ? ' *' : ''}
            </GlowText>
            {hasError ? (
              <YStack
                position="absolute"
                t={4}
                r={4}
                width={8}
                height={8}
                rounded="$circle"
                bg="$strawHatRedDeep"
              />
            ) : null}
          </ChannelBarItem>
        )
      })}
    </XStack>
  )
}
