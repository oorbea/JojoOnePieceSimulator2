import { Camera, Users } from '@tamagui/lucide-icons-2'
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
import { InfoHint } from '@/shared/components/presentational/info-hint'
import { LocaleTabs } from '@/shared/components/presentational/locale-tabs'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { notifyScroll } from '@/shared/lib/scroll-bus'
import { DEFAULT_LOCALE } from '@/shared/i18n'
import {
  fruitMasterySchema,
  hakiLevelSchema,
  hamonLevelSchema,
  physicalFormSchema,
  powerRaritySchema,
  spinLevelSchema,
  type Locale,
} from '@/shared/contracts/enums'
import type {
  CharacterKind,
  JojoCharacterFormValues,
  OnePieceCharacterFormValues,
} from '@/features/characters/types/characters.types'

const KIND_OPTIONS: GlassSelectOption[] = [
  { value: 'JOJO', label: 'enums.manga.JOJO' },
  { value: 'ONE_PIECE', label: 'enums.manga.ONE_PIECE' },
]

type Props = {
  visible: boolean
  mode: 'create' | 'edit'
  kind: CharacterKind | null
  onSelectKind: (kind: CharacterKind) => void
  jojoControl: Control<JojoCharacterFormValues>
  jojoErrors: FieldErrors<JojoCharacterFormValues>
  onePieceControl: Control<OnePieceCharacterFormValues>
  onePieceErrors: FieldErrors<OnePieceCharacterFormValues>
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

// Create/edit Character form. Unlike StageFormModal (where manga is just
// another value), the manga choice here decides which of two completely
// different stat blocks the form renders and submits - JoJo's
// hamon/spin/battleIq vs One Piece's physicalForm/3 hakis/fruitMastery.
// `kind` is asked FIRST, as its own required step with no default in
// create mode; in edit mode it arrives already set and locked (changing a
// saved character's manga would mean destroying its subtype row, so the
// container never lets `onSelectKind` fire once mode === 'edit').
export function CharacterFormModal({
  visible,
  mode,
  kind,
  onSelectKind,
  jojoControl,
  jojoErrors,
  onePieceControl,
  onePieceErrors,
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

  const kindOptions = useMemo(() => KIND_OPTIONS.map((o) => ({ ...o, label: t(o.label) })), [t])
  const RARITY_OPTIONS: GlassSelectOption[] = useMemo(
    () => powerRaritySchema.options.map((v) => ({ value: v, label: t(`enums.rarity.${v}`) })),
    [t]
  )
  const HAMON_OPTIONS: GlassSelectOption[] = useMemo(
    () => hamonLevelSchema.options.map((v) => ({ value: v, label: t(`enums.hamonLevel.${v}`) })),
    [t]
  )
  const SPIN_OPTIONS: GlassSelectOption[] = useMemo(
    () => spinLevelSchema.options.map((v) => ({ value: v, label: t(`enums.spinLevel.${v}`) })),
    [t]
  )
  const PHYSICAL_FORM_OPTIONS: GlassSelectOption[] = useMemo(
    () =>
      physicalFormSchema.options.map((v) => ({ value: v, label: t(`enums.physicalForm.${v}`) })),
    [t]
  )
  const HAKI_OPTIONS: GlassSelectOption[] = useMemo(
    () => hakiLevelSchema.options.map((v) => ({ value: v, label: t(`enums.hakiLevel.${v}`) })),
    [t]
  )
  const FRUIT_MASTERY_OPTIONS: GlassSelectOption[] = useMemo(
    () =>
      fruitMasterySchema.options.map((v) => ({ value: v, label: t(`enums.fruitMastery.${v}`) })),
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
            {mode === 'create' ? t('characters.newCharacter') : t('characters.editTitle')}
          </GlowText>

          <ScrollView
            flex={1}
            minH={0}
            keyboardShouldPersistTaps="handled"
            onScroll={notifyScroll}
            scrollEventThrottle={16}
          >
            <YStack gap="$4" pb="$2">
              {mode === 'edit' ? (
                <YStack gap="$1.5">
                  <GlowText level="label">{t('characters.manga')}</GlowText>
                  <GlassPanel tone="plastic" px="$3" py="$2.5" rounded="$card" elevate={0}>
                    <GlowText level="label">{kind ? t(`enums.manga.${kind}`) : ''}</GlowText>
                  </GlassPanel>
                </YStack>
              ) : (
                <GlassSelect
                  label={t('characters.manga')}
                  options={kindOptions}
                  value={kind}
                  onChange={(v) => onSelectKind(v as CharacterKind)}
                />
              )}

              {kind == null ? (
                <GlowText level="label" tone="soft" align="center">
                  {t('characters.pickMangaFirst')}
                </GlowText>
              ) : (
                <>
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
                      {...a11yProps(t('characters.changePicture'), 'button', {
                        disabled: isPictureBusy,
                      })}
                    >
                      <InsetRing rounded="$card" />
                      {pictureUri ? (
                        <Image
                          source={{ uri: pictureUri }}
                          style={{ width: '100%', height: '100%' }}
                        />
                      ) : (
                        <YStack flex={1} items="center" justify="center">
                          <Users size={28} color="$wiiBlue" />
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

                  {kind === 'JOJO' ? (
                    <Controller
                      control={jojoControl}
                      name="name"
                      render={({ field }) => (
                        <GlassField
                          label={t('characters.name')}
                          value={field.value}
                          onChangeText={field.onChange}
                          error={jojoErrors.name?.message && t(jojoErrors.name.message)}
                        />
                      )}
                    />
                  ) : (
                    <Controller
                      control={onePieceControl}
                      name="name"
                      render={({ field }) => (
                        <GlassField
                          label={t('characters.name')}
                          value={field.value}
                          onChangeText={field.onChange}
                          error={onePieceErrors.name?.message && t(onePieceErrors.name.message)}
                        />
                      )}
                    />
                  )}

                  {kind === 'JOJO' ? (
                    <Controller
                      control={jojoControl}
                      name="rarity"
                      render={({ field }) => (
                        <GlassSelect
                          label={t('characters.rarity')}
                          options={RARITY_OPTIONS}
                          value={field.value}
                          onChange={field.onChange}
                          error={jojoErrors.rarity?.message}
                        />
                      )}
                    />
                  ) : (
                    <Controller
                      control={onePieceControl}
                      name="rarity"
                      render={({ field }) => (
                        <GlassSelect
                          label={t('characters.rarity')}
                          options={RARITY_OPTIONS}
                          value={field.value}
                          onChange={field.onChange}
                          error={onePieceErrors.rarity?.message}
                        />
                      )}
                    />
                  )}

                  {kind === 'JOJO' ? (
                    <XStack flexWrap="wrap" gap="$3">
                      <YStack flexBasis={150} grow={1}>
                        <Controller
                          control={jojoControl}
                          name="hamon"
                          render={({ field }) => (
                            <GlassSelect
                              label={t('characters.stats.hamon')}
                              options={HAMON_OPTIONS}
                              value={field.value}
                              onChange={field.onChange}
                            />
                          )}
                        />
                      </YStack>
                      <YStack flexBasis={150} grow={1}>
                        <Controller
                          control={jojoControl}
                          name="spin"
                          render={({ field }) => (
                            <GlassSelect
                              label={t('characters.stats.spin')}
                              options={SPIN_OPTIONS}
                              value={field.value}
                              onChange={field.onChange}
                            />
                          )}
                        />
                      </YStack>
                      <YStack flexBasis={150} grow={1} gap="$1">
                        <Controller
                          control={jojoControl}
                          name="battleIq"
                          render={({ field }) => (
                            <GlassField
                              label={t('characters.stats.battleIq')}
                              value={String(field.value)}
                              onChangeText={(text) =>
                                field.onChange(
                                  text === '' ? 0 : Number(text.replace(/[^0-9]/g, ''))
                                )
                              }
                              keyboardType="number-pad"
                              error={jojoErrors.battleIq?.message && t(jojoErrors.battleIq.message)}
                            />
                          )}
                        />
                        <XStack items="center" self="flex-start">
                          <InfoHint text={t('characters.battleIqHint')} />
                        </XStack>
                      </YStack>
                    </XStack>
                  ) : (
                    <YStack gap="$3">
                      <Controller
                        control={onePieceControl}
                        name="physicalForm"
                        render={({ field }) => (
                          <GlassSelect
                            label={t('characters.stats.physicalForm')}
                            options={PHYSICAL_FORM_OPTIONS}
                            value={field.value}
                            onChange={field.onChange}
                          />
                        )}
                      />
                      <XStack flexWrap="wrap" gap="$3">
                        <YStack flexBasis={150} grow={1}>
                          <Controller
                            control={onePieceControl}
                            name="armamentHaki"
                            render={({ field }) => (
                              <GlassSelect
                                label={t('characters.stats.armamentHaki')}
                                options={HAKI_OPTIONS}
                                value={field.value}
                                onChange={field.onChange}
                              />
                            )}
                          />
                        </YStack>
                        <YStack flexBasis={150} grow={1}>
                          <Controller
                            control={onePieceControl}
                            name="observationHaki"
                            render={({ field }) => (
                              <GlassSelect
                                label={t('characters.stats.observationHaki')}
                                options={HAKI_OPTIONS}
                                value={field.value}
                                onChange={field.onChange}
                              />
                            )}
                          />
                        </YStack>
                        <YStack flexBasis={150} grow={1}>
                          <Controller
                            control={onePieceControl}
                            name="conquerorHaki"
                            render={({ field }) => (
                              <GlassSelect
                                label={t('characters.stats.conquerorHaki')}
                                options={HAKI_OPTIONS}
                                value={field.value}
                                onChange={field.onChange}
                              />
                            )}
                          />
                        </YStack>
                      </XStack>
                      <Controller
                        control={onePieceControl}
                        name="fruitMastery"
                        render={({ field }) => (
                          <GlassSelect
                            label={t('characters.stats.fruitMastery')}
                            options={FRUIT_MASTERY_OPTIONS}
                            value={field.value}
                            onChange={field.onChange}
                          />
                        )}
                      />
                    </YStack>
                  )}

                  <LocaleTabs
                    value={activeLocale}
                    onChange={onLocaleChange}
                    requiredLocale={DEFAULT_LOCALE}
                    localesWithErrors={erroredLocales}
                    requiredLabel={t('locale.required')}
                    errorLabel={t('locale.hasError')}
                  />

                  {kind === 'JOJO' ? (
                    <Controller
                      key={`translations.${activeLocale}.description`}
                      control={jojoControl}
                      name={`translations.${activeLocale}.description`}
                      render={({ field }) => (
                        <GlassField
                          label={t('characters.description')}
                          value={field.value}
                          onChangeText={field.onChange}
                          error={
                            jojoErrors.translations?.[activeLocale]?.description?.message &&
                            t(jojoErrors.translations[activeLocale].description.message)
                          }
                          multiline
                          numberOfLines={3}
                          height={100}
                        />
                      )}
                    />
                  ) : (
                    <Controller
                      key={`translations.${activeLocale}.description`}
                      control={onePieceControl}
                      name={`translations.${activeLocale}.description`}
                      render={({ field }) => (
                        <GlassField
                          label={t('characters.description')}
                          value={field.value}
                          onChangeText={field.onChange}
                          error={
                            onePieceErrors.translations?.[activeLocale]?.description?.message &&
                            t(onePieceErrors.translations[activeLocale].description.message)
                          }
                          multiline
                          numberOfLines={3}
                          height={100}
                        />
                      )}
                    />
                  )}
                </>
              )}
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
                disabled={isSaving || kind == null}
                onPress={onSubmit}
                accessibilityLabel={t('characters.saveA11y')}
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
