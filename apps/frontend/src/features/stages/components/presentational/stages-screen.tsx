import { Map, Plus, TriangleAlert } from '@tamagui/lucide-icons-2'
import type { Control, FieldErrors } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { Spinner, XStack, YStack } from 'tamagui'

import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { DetailModal } from '@/shared/components/presentational/detail-modal'
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
import { StageDetail } from './stage-detail'
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

type BaseProps = {
  stages: StageResponse[]
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  search: string
  onSearchChange: (search: string) => void
  mangaFilter: string | null
  mangaFilterOptions: GlassSelectOption[]
  onMangaFilterChange: (manga: string | null) => void
  hasActiveFilters: boolean
  detailStage: StageResponse | null
  onOpenDetail: (stage: StageResponse) => void
  onCloseDetail: () => void
}

type WritableProps = {
  readOnly?: false
  onCreateNew: () => void
  onEdit: (stage: StageResponse) => void
  onDelete: (stage: StageResponse) => void
  openingEditId: string | null
  form: FormState
  deleteConfirm: ConfirmState
}

type ReadOnlyProps = {
  readOnly: true
}

type Props = BaseProps & (WritableProps | ReadOnlyProps)

// Pure UI - a card grid of Stages (filtered by search/manga) plus the
// create/edit modal, the delete confirmation, and the read-only detail
// modal. All data fetching, form state, and mutation wiring live in
// StagesContainer. Same shape as StandsScreen, with a search bar + manga
// filter added on top since the catalogue mixes two mangas' worth of
// content.
export function StagesScreen(props: Props) {
  const {
    stages,
    isLoading,
    isError,
    onRetry,
    search,
    onSearchChange,
    mangaFilter,
    mangaFilterOptions,
    onMangaFilterChange,
    hasActiveFilters,
    detailStage,
    onOpenDetail,
    onCloseDetail,
  } = props
  const { t } = useTranslation()
  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">{t('stages.title')}</GlowText>
          {props.readOnly ? null : (
            <GlossButton
              tone="green"
              btnSize="md"
              onPress={props.onCreateNew}
              accessibilityLabel={t('stages.newStage')}
            >
              <Plus size={18} color="white" /> {t('stages.newStage')}
            </GlossButton>
          )}
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
            {hasActiveFilters || props.readOnly ? null : (
              <GlossButton
                tone="green"
                btnSize="sm"
                onPress={props.onCreateNew}
                accessibilityLabel={t('stages.newStage')}
              >
                {t('stages.newStage')}
              </GlossButton>
            )}
          </GlassPanel>
        ) : (
          <XStack flexWrap="wrap" gap="$4" justify="center">
            {stages.map((stage) =>
              props.readOnly ? (
                <StageCard key={stage.id} stage={stage} onOpenDetail={() => onOpenDetail(stage)} readOnly />
              ) : (
                <StageCard
                  key={stage.id}
                  stage={stage}
                  onOpenDetail={() => onOpenDetail(stage)}
                  onEdit={() => props.onEdit(stage)}
                  onDelete={() => props.onDelete(stage)}
                  isEditBusy={props.openingEditId === stage.id}
                />
              )
            )}
          </XStack>
        )}
      </PageShell>

      <DetailModal
        visible={detailStage !== null}
        title={detailStage?.name ?? ''}
        onClose={onCloseDetail}
        closeA11y={t('common.close')}
        footer={
          detailStage && !props.readOnly ? (
            <GlossButton
              tone="blue"
              btnSize="md"
              onPress={() => {
                onCloseDetail()
                props.onEdit(detailStage)
              }}
              accessibilityLabel={t('stages.editA11y', { name: detailStage.name })}
            >
              {t('common.edit')}
            </GlossButton>
          ) : undefined
        }
      >
        {detailStage ? <StageDetail stage={detailStage} /> : null}
      </DetailModal>

      {props.readOnly ? null : (
        <>
          <StageFormModal
            visible={props.form.visible}
            mode={props.form.mode}
            control={props.form.control}
            errors={props.form.errors}
            onSubmit={props.form.onSubmit}
            onCancel={props.form.onCancel}
            isSaving={props.form.isSaving}
            pictureUri={props.form.pictureUri}
            onPickPicture={props.form.onPickPicture}
            isPictureBusy={props.form.isPictureBusy}
            activeLocale={props.form.activeLocale}
            onLocaleChange={props.form.onLocaleChange}
            erroredLocales={props.form.erroredLocales}
          />

          <ConfirmSheet
            visible={props.deleteConfirm.visible}
            title={t('stages.deleteConfirmTitle')}
            message={t('stages.deleteConfirmMessage', { name: props.deleteConfirm.stageName ?? '' })}
            confirmLabel={t('stages.deleteConfirmButton')}
            isConfirming={props.deleteConfirm.isConfirming}
            onConfirm={props.deleteConfirm.onConfirm}
            onCancel={props.deleteConfirm.onCancel}
          />
        </>
      )}
    </YStack>
  )
}
