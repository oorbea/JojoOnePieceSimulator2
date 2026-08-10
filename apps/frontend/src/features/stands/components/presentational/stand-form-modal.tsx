import { Camera, Sparkles } from '@tamagui/lucide-icons-2'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Image, Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { Controller, type Control, type FieldErrors } from 'react-hook-form'
import { ScrollView, Spinner, XStack, YStack } from 'tamagui'

import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlassSelect, type GlassSelectOption } from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { LocaleTabs } from '@/shared/components/presentational/locale-tabs'
import { SkillsField } from '@/shared/components/presentational/skills-field'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { DEFAULT_LOCALE } from '@/shared/i18n'
import { raritySchema, standStatSchema, type Locale } from '@/shared/lib/zod'
import type { StandFormValues } from '@/features/stands/types/stands.types'

const STAT_FIELD_NAMES: (keyof StandFormValues)[] = [
  'attackPower',
  'speed',
  'attackRange',
  'endurance',
  'precision',
  'potential',
]

type Props = {
  visible: boolean
  mode: 'create' | 'edit'
  control: Control<StandFormValues>
  errors: FieldErrors<StandFormValues>
  onSubmit: () => void
  onCancel: () => void
  isSaving: boolean
  evolvesFromOptions: GlassSelectOption[]
  pictureUri: string | null
  onPickPicture: () => void
  isPictureBusy: boolean
  activeLocale: Locale
  onLocaleChange: (locale: Locale) => void
  erroredLocales: Locale[]
}

// Create/edit Stand form, same full-window Modal + centered glass card
// recipe as ConfirmSheet/GlassSelect. All fields go through react-hook-form
// Controllers so the container only needs to own `control`/`errors`/submit
// instead of threading 13 individual value/onChange pairs down as props.
export function StandFormModal({
  visible,
  mode,
  control,
  errors,
  onSubmit,
  onCancel,
  isSaving,
  evolvesFromOptions,
  pictureUri,
  onPickPicture,
  isPictureBusy,
  activeLocale,
  onLocaleChange,
  erroredLocales,
}: Props) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()

  // Constants at module scope wouldn't react to changeLanguage() - these
  // are recomputed whenever the active language changes instead.
  const RARITY_OPTIONS: GlassSelectOption[] = useMemo(
    () => raritySchema.options.map((v) => ({ value: v, label: t(`enums.rarity.${v}`) })),
    [t]
  )
  const STAND_STAT_OPTIONS: GlassSelectOption[] = useMemo(
    () => standStatSchema.options.map((v) => ({ value: v, label: t(`enums.standStat.${v}`) })),
    [t]
  )
  const STAT_FIELDS = useMemo(
    () =>
      STAT_FIELD_NAMES.map((name) => ({
        name,
        label: t(`stands.stats.${name}`),
      })),
    [t]
  )

  return (
    <Modal visible={visible} transparent animationType="fade" onRequestClose={onCancel} statusBarTranslucent>
      <YStack
        flex={1}
        items="center"
        justify="center"
        p="$4"
        pt={insets.top + 16}
        pb={insets.bottom + 16}
        bg="rgba(10,12,20,0.45)"
      >
        <GlassPanel
          tone="strong"
          radiusSize="panel"
          elevate={3}
          width="100%"
          maxW={520}
          maxH="90%"
          p="$5"
          gap="$4"
        >
          <GlowText level="heading" align="center">
            {mode === 'create' ? t('stands.newStand') : t('stands.editTitle')}
          </GlowText>

          <ScrollView flex={1} minH={0} keyboardShouldPersistTaps="handled">
            <YStack gap="$4" pb="$2">
              <YStack items="center" gap="$2">
                <YStack
                  width={96}
                  height={96}
                  rounded="$card"
                  overflow="hidden"
                  position="relative"
                  bg="$plasticEdge"
                  onPress={onPickPicture}
                  cursor="pointer"
                  {...a11yProps(t('stands.changePicture'), 'button', { disabled: isPictureBusy })}
                >
                  <InsetRing rounded="$card" />
                  {pictureUri ? (
                    <Image source={{ uri: pictureUri }} style={{ width: '100%', height: '100%' }} />
                  ) : (
                    <YStack flex={1} items="center" justify="center">
                      <Sparkles size={28} color="$standPurple" />
                    </YStack>
                  )}
                  <YStack
                    position="absolute"
                    b={0}
                    r={0}
                    width={30}
                    height={30}
                    rounded="$circle"
                    items="center"
                    justify="center"
                    bg="$wiiBlue"
                    borderWidth={1.5}
                    borderColor="$glassEdge"
                  >
                    <Camera size={14} color="white" strokeWidth={2.5} />
                  </YStack>
                  {isPictureBusy ? (
                    <YStack
                      position="absolute"
                      t={0}
                      l={0}
                      r={0}
                      b={0}
                      items="center"
                      justify="center"
                      bg="rgba(10,12,20,0.45)"
                    >
                      <Spinner size="small" color="white" />
                    </YStack>
                  ) : null}
                </YStack>
              </YStack>

              <Controller
                control={control}
                name="name"
                render={({ field }) => (
                  <GlassField
                    label={t('stands.name')}
                    value={field.value}
                    onChangeText={field.onChange}
                    error={errors.name?.message && t(errors.name.message)}
                  />
                )}
              />

              <LocaleTabs
                value={activeLocale}
                onChange={onLocaleChange}
                requiredLocale={DEFAULT_LOCALE}
                localesWithErrors={erroredLocales}
                requiredLabel={t('locale.required')}
                errorLabel={t('locale.hasError')}
              />

              <Controller
                key={`translations.${activeLocale}.description`}
                control={control}
                name={`translations.${activeLocale}.description`}
                render={({ field }) => (
                  <GlassField
                    label={t('stands.description')}
                    value={field.value}
                    onChangeText={field.onChange}
                    error={
                      errors.translations?.[activeLocale]?.description?.message &&
                      t(errors.translations[activeLocale].description.message)
                    }
                    multiline
                    numberOfLines={3}
                    height={100}
                  />
                )}
              />

              <Controller
                key={`translations.${activeLocale}.skills`}
                control={control}
                name={`translations.${activeLocale}.skills`}
                render={({ field }) => (
                  <SkillsField
                    label={t('stands.skills')}
                    skills={field.value}
                    onAdd={(skill) => field.onChange([...field.value, skill])}
                    onRemove={(index) => field.onChange(field.value.filter((_, i) => i !== index))}
                    error={
                      errors.translations?.[activeLocale]?.skills?.message &&
                      t(errors.translations[activeLocale].skills.message)
                    }
                  />
                )}
              />

              <Controller
                control={control}
                name="rarity"
                render={({ field }) => (
                  <GlassSelect
                    label={t('stands.rarity')}
                    options={RARITY_OPTIONS}
                    value={field.value}
                    onChange={(v) => field.onChange(v)}
                    error={errors.rarity?.message}
                  />
                )}
              />

              <XStack flexWrap="wrap" gap="$3">
                {STAT_FIELDS.map(({ name, label }) => (
                  <YStack key={name} flexBasis={150} grow={1}>
                    <Controller
                      control={control}
                      name={name}
                      render={({ field }) => (
                        <GlassSelect
                          label={label}
                          options={STAND_STAT_OPTIONS}
                          value={field.value as string}
                          onChange={(v) => field.onChange(v)}
                        />
                      )}
                    />
                  </YStack>
                ))}
              </XStack>

              <Controller
                control={control}
                name="evolvesFromId"
                render={({ field }) => (
                  <GlassSelect
                    label={t('stands.evolvesFrom')}
                    placeholder={t('stands.evolvesFromPlaceholder')}
                    options={evolvesFromOptions}
                    value={field.value}
                    onChange={(v) => field.onChange(v)}
                    searchable
                    clearable
                  />
                )}
              />
            </YStack>
          </ScrollView>

          <XStack gap="$2">
            <YStack flex={1}>
              <GlossButton tone="glass" btnSize="md" disabled={isSaving} onPress={onCancel} accessibilityLabel={t('common.cancel')}>
                {t('common.cancel')}
              </GlossButton>
            </YStack>
            <YStack flex={1}>
              <GlossButton
                tone="blue"
                btnSize="md"
                disabled={isSaving}
                onPress={onSubmit}
                accessibilityLabel={t('stands.saveA11y')}
              >
                {isSaving ? t('common.saving') : t('common.save')}
              </GlossButton>
            </YStack>
          </XStack>
        </GlassPanel>
      </YStack>
    </Modal>
  )
}
