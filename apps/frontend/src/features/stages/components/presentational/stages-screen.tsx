import { Map, Plus, TriangleAlert } from '@tamagui/lucide-icons-2'
import type { Control, FieldErrors } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { Spinner, XStack, YStack } from 'tamagui'

import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { GlassField } from '@/shared/components/presentational/glass-field'
import { GlassPanel } from '@/shared/components/presentational/glass-panel'
import {
  GlassSelect,
  type GlassSelectOption,
} from '@/shared/components/presentational/glass-select'
import { GlossButton } from '@/shared/components/presentational/gloss-button'
import { GlowText } from '@/shared/components/presentational/glow-text'
import { PageShell } from '@/shared/components/presentational/page-shell'
import type { Locale } from '@/shared/contracts/enums'
import type { StageFormValues, StageResponse } from '@/features/stages/types/stages.types'

import { StageCard } from './stage-card'
import { StageFormModal } from './stage-form-modal'

type ConfirmState = {
  visible: boolean
  isConfirming: boolean
  onConfirm: () => void
  onCancel: () => void
  stageName?: string
}

type FormState = {
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

type Props = {
  stages: StageResponse[]
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  onCreateNew: () => void
  onEdit: (stage: StageResponse) => void
  onDelete: (stage: StageResponse) => void
  openingEditId: string | null
  search: string
  onSearchChange: (search: string) => void
  mangaFilter: string | null
  mangaFilterOptions: GlassSelectOption[]
  onMangaFilterChange: (manga: string | null) => void
  hasActiveFilters: boolean
  form: FormState
  deleteConfirm: ConfirmState
}

// Pure UI - a card grid of Stages (filtered by search/manga) plus the
// create/edit modal and the delete confirmation. All data fetching, form
// state, and mutation wiring live in StagesContainer. Same shape as
// StandsScreen, with a search bar + manga filter added on top since the
// catalogue mixes two mangas' worth of content.
export function StagesScreen({
  stages,
  isLoading,
  isError,
  onRetry,
  onCreateNew,
  onEdit,
  onDelete,
  openingEditId,
  search,
  onSearchChange,
  mangaFilter,
  mangaFilterOptions,
  onMangaFilterChange,
  hasActiveFilters,
  form,
  deleteConfirm,
}: Props) {
  const { t } = useTranslation()
  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">{t('stages.title')}</GlowText>
          <GlossButton
            tone="green"
            btnSize="md"
            onPress={onCreateNew}
            accessibilityLabel={t('stages.newStage')}
          >
            <Plus size={18} color="white" /> {t('stages.newStage')}
          </GlossButton>
        </XStack>

        <XStack width="100%" flexWrap="wrap" gap="$3">
          <YStack flexBasis={220} grow={1}>
            <GlassField
              label={t('common.search')}
              value={search}
              onChangeText={onSearchChange}
              placeholder={t('stages.searchPlaceholder')}
            />
          </YStack>
          <YStack flexBasis={200} grow={1}>
            <GlassSelect
              label={t('stages.filterManga')}
              options={mangaFilterOptions}
              value={mangaFilter}
              onChange={onMangaFilterChange}
              clearable
            />
          </YStack>
        </XStack>

        {isLoading ? (
          <YStack width="100%" items="center" p="$6">
            <Spinner size="large" />
          </YStack>
        ) : isError ? (
          // Distinct from the "no Stages match" empty state below - a
          // failed GET must never look like an empty catalogue.
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <TriangleAlert size={28} color="$strawHatRed" />
            <GlowText level="label" align="center">
              {t('stages.errorTitle')}
            </GlowText>
            <GlossButton
              tone="blue"
              btnSize="sm"
              onPress={onRetry}
              accessibilityLabel={t('stages.retry')}
            >
              {t('stages.retry')}
            </GlossButton>
          </GlassPanel>
        ) : stages.length === 0 ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <Map size={28} color="$wiiBlue" />
            <GlowText level="label" align="center">
              {t(hasActiveFilters ? 'stages.emptyFilteredTitle' : 'stages.emptyTitle')}
            </GlowText>
            {hasActiveFilters ? null : (
              <GlossButton
                tone="green"
                btnSize="sm"
                onPress={onCreateNew}
                accessibilityLabel={t('stages.newStage')}
              >
                {t('stages.newStage')}
              </GlossButton>
            )}
          </GlassPanel>
        ) : (
          <XStack flexWrap="wrap" gap="$4" justify="center">
            {stages.map((stage) => (
              <StageCard
                key={stage.id}
                stage={stage}
                onEdit={() => onEdit(stage)}
                onDelete={() => onDelete(stage)}
                isEditBusy={openingEditId === stage.id}
              />
            ))}
          </XStack>
        )}
      </PageShell>

      <StageFormModal
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
        activeLocale={form.activeLocale}
        onLocaleChange={form.onLocaleChange}
        erroredLocales={form.erroredLocales}
      />

      <ConfirmSheet
        visible={deleteConfirm.visible}
        title={t('stages.deleteConfirmTitle')}
        message={t('stages.deleteConfirmMessage', { name: deleteConfirm.stageName ?? '' })}
        confirmLabel={t('stages.deleteConfirmButton')}
        isConfirming={deleteConfirm.isConfirming}
        onConfirm={deleteConfirm.onConfirm}
        onCancel={deleteConfirm.onCancel}
      />
    </YStack>
  )
}
