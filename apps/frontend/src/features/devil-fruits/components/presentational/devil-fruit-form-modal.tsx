import { Apple, Camera } from '@tamagui/lucide-icons-2'
import { Image, Modal } from 'react-native'
import { useSafeAreaInsets } from 'react-native-safe-area-context'
import { Controller, type Control, type FieldErrors } from 'react-hook-form'
import { ScrollView, Spinner, XStack, YStack } from 'tamagui'

import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlassSelect, type GlassSelectOption } from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { SkillsField } from '@/shared/components/presentational/skills-field'
import { InsetRing } from '@/shared/components/presentational/wii-card'
import { a11yProps } from '@/shared/lib/a11y'
import { fruitTypeSchema, raritySchema } from '@/shared/lib/zod'
import type { DevilFruitFormValues } from '@/features/devil-fruits/types/devil-fruits.types'

const RARITY_OPTIONS: GlassSelectOption[] = raritySchema.options.map((v) => ({ value: v, label: v }))
const FRUIT_TYPE_OPTIONS: GlassSelectOption[] = fruitTypeSchema.options.map((v) => ({ value: v, label: v }))

type Props = {
  visible: boolean
  mode: 'create' | 'edit'
  control: Control<DevilFruitFormValues>
  errors: FieldErrors<DevilFruitFormValues>
  onSubmit: () => void
  onCancel: () => void
  isSaving: boolean
  pictureUri: string | null
  onPickPicture: () => void
  isPictureBusy: boolean
}

export function DevilFruitFormModal({
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
}: Props) {
  const insets = useSafeAreaInsets()

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
            {mode === 'create' ? 'New Devil Fruit' : 'Edit Devil Fruit'}
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
                  {...a11yProps('Change Devil Fruit picture', 'button', { disabled: isPictureBusy })}
                >
                  <InsetRing rounded="$card" />
                  {pictureUri ? (
                    <Image source={{ uri: pictureUri }} style={{ width: '100%', height: '100%' }} />
                  ) : (
                    <YStack flex={1} items="center" justify="center">
                      <Apple size={28} color="$strawHatRed" />
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
                    label="Name"
                    value={field.value}
                    onChangeText={field.onChange}
                    error={errors.name?.message}
                  />
                )}
              />

              <Controller
                control={control}
                name="description"
                render={({ field }) => (
                  <GlassField
                    label="Description"
                    value={field.value}
                    onChangeText={field.onChange}
                    error={errors.description?.message}
                    multiline
                    numberOfLines={3}
                    height={90}
                  />
                )}
              />

              <Controller
                control={control}
                name="rarity"
                render={({ field }) => (
                  <GlassSelect
                    label="Rarity"
                    options={RARITY_OPTIONS}
                    value={field.value}
                    onChange={(v) => field.onChange(v)}
                    error={errors.rarity?.message}
                  />
                )}
              />

              <Controller
                control={control}
                name="fruitType"
                render={({ field }) => (
                  <GlassSelect
                    label="Fruit Type"
                    options={FRUIT_TYPE_OPTIONS}
                    value={field.value}
                    onChange={(v) => field.onChange(v)}
                    error={errors.fruitType?.message}
                  />
                )}
              />

              <Controller
                control={control}
                name="skills"
                render={({ field }) => (
                  <SkillsField
                    label="Skills"
                    skills={field.value}
                    onAdd={(skill) => field.onChange([...field.value, skill])}
                    onRemove={(index) => field.onChange(field.value.filter((_, i) => i !== index))}
                    error={errors.skills?.message}
                  />
                )}
              />
            </YStack>
          </ScrollView>

          <XStack gap="$2">
            <YStack flex={1}>
              <GlossButton tone="glass" btnSize="md" disabled={isSaving} onPress={onCancel} accessibilityLabel="Cancel">
                Cancel
              </GlossButton>
            </YStack>
            <YStack flex={1}>
              <GlossButton
                tone="blue"
                btnSize="md"
                disabled={isSaving}
                onPress={onSubmit}
                accessibilityLabel="Save Devil Fruit"
              >
                {isSaving ? 'Saving…' : 'Save'}
              </GlossButton>
            </YStack>
          </XStack>
        </GlassPanel>
      </YStack>
    </Modal>
  )
}
