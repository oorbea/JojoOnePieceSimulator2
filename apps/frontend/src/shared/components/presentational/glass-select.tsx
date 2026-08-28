import { ChevronDown, X } from '@tamagui/lucide-icons-2'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { ScrollView, XStack, YStack } from 'tamagui'

import { a11yProps } from '@/shared/lib/a11y'
import { notifyScroll } from '@/shared/lib/scroll-bus'

import { GlassField } from './glass-field'
import { GlowText } from './glow-text'
import { InsetRing } from './wii-card'

export type GlassSelectOption = { value: string; label: string }

type GlassSelectProps = {
  label: string
  error?: string
  options: GlassSelectOption[]
  value: string | null
  onChange: (value: string | null) => void
  placeholder?: string
  /** Adds a filter field at the top of the option list — needed once the
   * list is long enough that scrolling to find one entry isn't practical
   * (e.g. picking a Stand out of the whole roster). */
  searchable?: boolean
  /** Shows a clear (X) button on the field once a value is selected, and
   * resets `value` to null instead of forcing a re-pick. */
  clearable?: boolean
}

// The one select primitive the admin forms need for enum fields (rarity,
// stand stats, fruit type) and the searchable/clearable evolvesFrom picker.
// Renders through the same full-window Modal + centered GlassPanel recipe
// as ConfirmSheet — a native `<select>` styled to match the rest of the
// glass system doesn't exist without diverging web/native markup, and this
// keeps one identical UX on every platform.
export function GlassSelect({
  label,
  error,
  options,
  value,
  onChange,
  placeholder,
  searchable = false,
  clearable = false,
}: GlassSelectProps) {
  const { t } = useTranslation()
  const effectivePlaceholder = placeholder ?? t('common.select')
  const insets = useSafeAreaInsets()
  const [isOpen, setIsOpen] = useState(false)
  const [search, setSearch] = useState('')

  const selected = useMemo(() => options.find((o) => o.value === value) ?? null, [options, value])

  const filtered = useMemo(() => {
    if (!searchable || !search.trim()) return options
    const needle = search.trim().toLowerCase()
    return options.filter((o) => o.label.toLowerCase().includes(needle))
  }, [options, search, searchable])

  const openPicker = () => {
    setSearch('')
    setIsOpen(true)
  }

  const choose = (option: GlassSelectOption | null) => {
    onChange(option?.value ?? null)
    setIsOpen(false)
  }

  return (
    <YStack width="100%" gap="$1.5">
      <GlowText level="label">{label}</GlowText>
      <XStack
        position="relative"
        rounded="$card"
        overflow="hidden"
        items="center"
        onPress={openPicker}
        cursor="pointer"
        {...a11yProps(selected ? `${label}: ${selected.label}` : `${label}: ${effectivePlaceholder}`, 'button')}
      >
        <InsetRing rounded="$card" />
        <XStack
          flex={1}
          bg="$plasticFill"
          borderWidth={1.5}
          borderColor={error ? '$strawHatRedDeep' : '$glassEdge'}
          rounded="$card"
          height={52}
          px="$4"
          items="center"
          justify="space-between"
        >
          <GlowText level="label" color={selected ? '$panelText' : '$panelTextSoft'} numberOfLines={1}>
            {selected ? selected.label : effectivePlaceholder}
          </GlowText>
          <XStack items="center" gap="$2">
            {clearable && selected ? (
              <XStack
                p="$1"
                onPress={(e: { stopPropagation?: () => void }) => {
                  e.stopPropagation?.()
                  choose(null)
                }}
                {...a11yProps(t('common.clearLabel', { label }), 'button')}
              >
                <X size={16} color="$panelTextSoft" />
              </XStack>
            ) : null}
            <ChevronDown size={18} color="$panelTextSoft" />
          </XStack>
        </XStack>
      </XStack>
      {error ? (
        <GlowText level="label" color="$strawHatRedDeep" {...a11yProps(error, 'alert')}>
          {error}
        </GlowText>
      ) : null}

      <Modal
        visible={isOpen}
        transparent
        animationType="fade"
        onRequestClose={() => setIsOpen(false)}
        statusBarTranslucent
      >
        <YStack
          flex={1}
          items="center"
          justify="center"
          p="$4"
          pt={insets.top + 16}
          pb={insets.bottom + 16}
          bg="rgba(10,12,20,0.45)"
          onPress={() => setIsOpen(false)}
          {...a11yProps(label, 'none')}
        >
          <YStack
            width="100%"
            maxW={380}
            maxH="70%"
            rounded="$panel"
            borderWidth={1.5}
            borderColor="$glassEdge"
            bg="$plasticFill"
            p="$4"
            gap="$3"
            onPress={(e: { stopPropagation?: () => void }) => e.stopPropagation?.()}
          >
            <GlowText level="heading" align="center">
              {label}
            </GlowText>

            {searchable ? (
              <GlassField
                label={t('common.search')}
                value={search}
                onChangeText={setSearch}
                placeholder={t('common.typeToFilter')}
                autoFocus
              />
            ) : null}

            <ScrollView onScroll={notifyScroll} scrollEventThrottle={16}>
              <YStack gap="$1.5">
                {clearable ? (
                  <YStack
                    p="$3"
                    rounded="$card"
                    bg={value === null ? '$glassFillStrong' : undefined}
                    onPress={() => choose(null)}
                    cursor="pointer"
                    {...a11yProps(t('common.none'), 'button')}
                  >
                    <GlowText level="label" tone="soft">
                      {t('common.none')}
                    </GlowText>
                  </YStack>
                ) : null}
                {filtered.map((option) => (
                  <YStack
                    key={option.value}
                    p="$3"
                    rounded="$card"
                    bg={option.value === value ? '$glassFillStrong' : undefined}
                    onPress={() => choose(option)}
                    cursor="pointer"
                    {...a11yProps(option.label, 'button')}
                  >
                    <GlowText level="label">{option.label}</GlowText>
                  </YStack>
                ))}
                {searchable && filtered.length === 0 ? (
                  <GlowText level="label" tone="soft" align="center">
                    {t('common.noMatches')}
                  </GlowText>
                ) : null}
              </YStack>
            </ScrollView>
          </YStack>
        </YStack>
      </Modal>
    </YStack>
  )
}
