import { Plus, Sparkles } from '@tamagui/lucide-icons-2'
import type { Control, FieldErrors } from 'react-hook-form'
import { Spinner, XStack, YStack } from 'tamagui'

import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import type { GlassSelectOption } from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import type { StandFormValues, StandResponse } from '@/features/stands/types/stands.types'

import { StandCard } from './stand-card'
import { StandFormModal } from './stand-form-modal'

type ConfirmState = {
  visible: boolean
  isConfirming: boolean
  onConfirm: () => void
  onCancel: () => void
  standName?: string
}

type FormState = {
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
}

type Props = {
  stands: StandResponse[]
  isLoading: boolean
  onCreateNew: () => void
  onEdit: (stand: StandResponse) => void
  onDelete: (stand: StandResponse) => void
  form: FormState
  deleteConfirm: ConfirmState
}

// Pure UI — a card grid of Stands plus the create/edit modal and the delete
// confirmation. All data fetching, form state, and mutation wiring live in
// StandsContainer.
export function StandsScreen({ stands, isLoading, onCreateNew, onEdit, onDelete, form, deleteConfirm }: Props) {
  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" navPadding scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">Stands</GlowText>
          <GlossButton tone="green" btnSize="md" onPress={onCreateNew} accessibilityLabel="New Stand">
            <Plus size={18} color="white" /> New Stand
          </GlossButton>
        </XStack>

        {isLoading ? (
          <YStack width="100%" items="center" p="$6">
            <Spinner size="large" />
          </YStack>
        ) : stands.length === 0 ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <Sparkles size={28} color="$standPurple" />
            <GlowText level="label" align="center">
              No Stands yet. Create the first one.
            </GlowText>
            <GlossButton tone="green" btnSize="sm" onPress={onCreateNew} accessibilityLabel="New Stand">
              New Stand
            </GlossButton>
          </GlassPanel>
        ) : (
          <XStack flexWrap="wrap" gap="$4" justify="center">
            {stands.map((stand) => (
              <StandCard
                key={stand.id}
                stand={stand}
                onEdit={() => onEdit(stand)}
                onDelete={() => onDelete(stand)}
              />
            ))}
          </XStack>
        )}
      </PageShell>

      <StandFormModal
        visible={form.visible}
        mode={form.mode}
        control={form.control}
        errors={form.errors}
        onSubmit={form.onSubmit}
        onCancel={form.onCancel}
        isSaving={form.isSaving}
        evolvesFromOptions={form.evolvesFromOptions}
        pictureUri={form.pictureUri}
        onPickPicture={form.onPickPicture}
        isPictureBusy={form.isPictureBusy}
      />

      <ConfirmSheet
        visible={deleteConfirm.visible}
        title="Delete Stand?"
        message={`"${deleteConfirm.standName ?? ''}" will be permanently deleted. This can't be undone.`}
        confirmLabel="Delete Stand"
        isConfirming={deleteConfirm.isConfirming}
        onConfirm={deleteConfirm.onConfirm}
        onCancel={deleteConfirm.onCancel}
      />
    </YStack>
  )
}
