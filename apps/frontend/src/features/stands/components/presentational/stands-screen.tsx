import { Plus, Sparkles, TriangleAlert } from '@tamagui/lucide-icons-2'
import type { Control, FieldErrors } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { Spinner, XStack, YStack } from 'tamagui'

import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import type { GlassSelectOption } from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import type { Locale } from '@/shared/lib/zod'
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
  activeLocale: Locale
  onLocaleChange: (locale: Locale) => void
  erroredLocales: Locale[]
}

type Props = {
  stands: StandResponse[]
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  onCreateNew: () => void
  onEdit: (stand: StandResponse) => void
  onDelete: (stand: StandResponse) => void
  openingEditId: string | null
  form: FormState
  deleteConfirm: ConfirmState
}

// Pure UI — a card grid of Stands plus the create/edit modal and the delete
// confirmation. All data fetching, form state, and mutation wiring live in
// StandsContainer.
export function StandsScreen({
  stands,
  isLoading,
  isError,
  onRetry,
  onCreateNew,
  onEdit,
  onDelete,
  openingEditId,
  form,
  deleteConfirm,
}: Props) {
  const { t } = useTranslation()
  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">{t('stands.title')}</GlowText>
          <GlossButton tone="green" btnSize="md" onPress={onCreateNew} accessibilityLabel={t('stands.newStand')}>
            <Plus size={18} color="white" /> {t('stands.newStand')}
          </GlossButton>
        </XStack>

        {isLoading ? (
          <YStack width="100%" items="center" p="$6">
            <Spinner size="large" />
          </YStack>
        ) : isError ? (
          // Distinct from the "no Stands yet" empty state below - a failed
          // GET (e.g. the backend's 500 on a legacy row with empty skills,
          // or a dropped response) must never look like an empty catalogue.
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <TriangleAlert size={28} color="$strawHatRed" />
            <GlowText level="label" align="center">
              {t('stands.errorTitle')}
            </GlowText>
            <GlossButton tone="blue" btnSize="sm" onPress={onRetry} accessibilityLabel={t('stands.retry')}>
              {t('stands.retry')}
            </GlossButton>
          </GlassPanel>
        ) : stands.length === 0 ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <Sparkles size={28} color="$standPurple" />
            <GlowText level="label" align="center">
              {t('stands.emptyTitle')}
            </GlowText>
            <GlossButton tone="green" btnSize="sm" onPress={onCreateNew} accessibilityLabel={t('stands.newStand')}>
              {t('stands.newStand')}
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
                isEditBusy={openingEditId === stand.id}
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
        activeLocale={form.activeLocale}
        onLocaleChange={form.onLocaleChange}
        erroredLocales={form.erroredLocales}
      />

      <ConfirmSheet
        visible={deleteConfirm.visible}
        title={t('stands.deleteConfirmTitle')}
        message={t('stands.deleteConfirmMessage', { name: deleteConfirm.standName ?? '' })}
        confirmLabel={t('stands.deleteConfirmButton')}
        isConfirming={deleteConfirm.isConfirming}
        onConfirm={deleteConfirm.onConfirm}
        onCancel={deleteConfirm.onCancel}
      />
    </YStack>
  )
}
