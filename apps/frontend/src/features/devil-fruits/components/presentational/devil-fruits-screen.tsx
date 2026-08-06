import { Apple, Plus } from '@tamagui/lucide-icons-2'
import type { Control, FieldErrors } from 'react-hook-form'
import { Spinner, XStack, YStack } from 'tamagui'

import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import type { DevilFruitFormValues, DevilFruitResponse } from '@/features/devil-fruits/types/devil-fruits.types'

import { DevilFruitCard } from './devil-fruit-card'
import { DevilFruitFormModal } from './devil-fruit-form-modal'

type ConfirmState = {
  visible: boolean
  isConfirming: boolean
  onConfirm: () => void
  onCancel: () => void
  fruitName?: string
}

type FormState = {
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

type Props = {
  devilFruits: DevilFruitResponse[]
  isLoading: boolean
  onCreateNew: () => void
  onEdit: (devilFruit: DevilFruitResponse) => void
  onDelete: (devilFruit: DevilFruitResponse) => void
  form: FormState
  deleteConfirm: ConfirmState
}

export function DevilFruitsScreen({
  devilFruits,
  isLoading,
  onCreateNew,
  onEdit,
  onDelete,
  form,
  deleteConfirm,
}: Props) {
  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">Devil Fruits</GlowText>
          <GlossButton tone="green" btnSize="md" onPress={onCreateNew} accessibilityLabel="New Devil Fruit">
            <Plus size={18} color="white" /> New Devil Fruit
          </GlossButton>
        </XStack>

        {isLoading ? (
          <YStack width="100%" items="center" p="$6">
            <Spinner size="large" />
          </YStack>
        ) : devilFruits.length === 0 ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <Apple size={28} color="$strawHatRed" />
            <GlowText level="label" align="center">
              No Devil Fruits yet. Create the first one.
            </GlowText>
            <GlossButton tone="green" btnSize="sm" onPress={onCreateNew} accessibilityLabel="New Devil Fruit">
              New Devil Fruit
            </GlossButton>
          </GlassPanel>
        ) : (
          <XStack flexWrap="wrap" gap="$4" justify="center">
            {devilFruits.map((devilFruit) => (
              <DevilFruitCard
                key={devilFruit.id}
                devilFruit={devilFruit}
                onEdit={() => onEdit(devilFruit)}
                onDelete={() => onDelete(devilFruit)}
              />
            ))}
          </XStack>
        )}
      </PageShell>

      <DevilFruitFormModal
        visible={form.visible}
        mode={form.mode}
        control={form.control}
        errors={form.errors}
        onSubmit={form.onSubmit}
        onCancel={form.onCancel}
        isSaving={form.isSaving}
        pictureUri={form.pictureUri}
        onPickPicture={form.onPickPicture}
        isPictureBusy={form.isPictureBusy}
      />

      <ConfirmSheet
        visible={deleteConfirm.visible}
        title="Delete Devil Fruit?"
        message={`"${deleteConfirm.fruitName ?? ''}" will be permanently deleted. This can't be undone.`}
        confirmLabel="Delete Devil Fruit"
        isConfirming={deleteConfirm.isConfirming}
        onConfirm={deleteConfirm.onConfirm}
        onCancel={deleteConfirm.onCancel}
      />
    </YStack>
  )
}
