import { Camera, Map } from '@tamagui/lucide-icons-2'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Image, Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { Controller, type Control, type FieldErrors } from 'react-hook-form'
import { ScrollView, Spinner, XStack, YStack } from 'tamagui'

import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import {
  GlassSelect,
  type GlassSelectOption,
} from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { LocaleTabs } from '@/shared/components/presentational/locale-tabs'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { SUPPORTED_LOCALES } from '@/shared/i18n'
import { mangaSchema, type Locale } from '@/shared/lib/zod'
import type { StageFormValues } from '@/features/stages/types/stages.types'

type Props = {
  visible: boolean
  mode: 'create' | 'edit'
  control: Control<StageFormValues>
  errors: FieldErrors<StageFormValues>
  onSubmit: () => void
  onCancel: () => void
  isSaving: boolean
  pictureUri: string | null
  onPickPicture: () => void
  isPictureBusy: boolean
  activeLocale: Locale
  onLocaleChange: (locale: Locale) => void
  erroredLocales: Locale[]
}

// Create/edit Stage form - same full-window Modal + centered glass card
// recipe as StandFormModal, minus SkillsField (a Stage has none) and with
// every LocaleTabs pill starred instead of just en-GB, since every locale
// is mandatory here (see the vault's game-stage-content.md).
export function StageFormModal({
  visible,
  mode,
  control,
  errors,
  onSubmit,
  onCancel,
  isSaving,
  pictureUri,
  onPickPicture,
  isPictureBusy,
  activeLocale,
  onLocaleChange,
  erroredLocales,
}: Props) {
  const { t } = useTranslation()
  const insets = useSafeAreaInsets()

  // Recomputed on language change, same reasoning as StandFormModal's
  // RARITY_OPTIONS/STAND_STAT_OPTIONS.
  const MANGA_OPTIONS: GlassSelectOption[] = useMemo(
    () => mangaSchema.options.map((v) => ({ value: v, label: t(`enums.manga.${v}`) })),
    [t]
  )

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onCancel}
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
            {mode === 'create' ? t('stages.newStage') : t('stages.editTitle')}
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
                  {...a11yProps(t('stages.changePicture'), 'button', { disabled: isPictureBusy })}
                >
                  <InsetRing rounded="$card" />
                  {pictureUri ? (
                    <Image source={{ uri: pictureUri }} style={{ width: '100%', height: '100%' }} />
                  ) : (
                    <YStack flex={1} items="center" justify="center">
                      <Map size={28} color="$wiiBlue" />
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
                    label={t('stages.name')}
                    value={field.value}
                    onChangeText={field.onChange}
                    error={errors.name?.message && t(errors.name.message)}
                  />
                )}
              />

              <XStack flexWrap="wrap" gap="$3">
                <YStack flexBasis={180} grow={1}>
                  <Controller
                    control={control}
                    name="manga"
                    render={({ field }) => (
                      <GlassSelect
                        label={t('stages.manga')}
                        options={MANGA_OPTIONS}
                        value={field.value}
                        onChange={(v) => field.onChange(v)}
                        error={errors.manga?.message}
                      />
                    )}
                  />
                </YStack>
                <YStack flexBasis={120} grow={1}>
                  <Controller
                    control={control}
                    name="order"
                    render={({ field }) => (
                      <GlassField
                        label={t('stages.order')}
                        value={String(field.value)}
                        onChangeText={(text) =>
                          field.onChange(text === '' ? 0 : Number(text.replace(/[^0-9]/g, '')))
                        }
                        keyboardType="number-pad"
                        error={errors.order?.message && t(errors.order.message)}
                      />
                    )}
                  />
                </YStack>
              </XStack>

              <LocaleTabs
                value={activeLocale}
                onChange={onLocaleChange}
                requiredLocale={SUPPORTED_LOCALES}
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
                    label={t('stages.description')}
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
            </YStack>
          </ScrollView>

          <XStack gap="$2">
            <YStack flex={1}>
              <GlossButton
                tone="glass"
                btnSize="md"
                disabled={isSaving}
                onPress={onCancel}
                accessibilityLabel={t('common.cancel')}
              >
                {t('common.cancel')}
              </GlossButton>
            </YStack>
            <YStack flex={1}>
              <GlossButton
                tone="blue"
                btnSize="md"
                disabled={isSaving}
                onPress={onSubmit}
                accessibilityLabel={t('stages.saveA11y')}
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
