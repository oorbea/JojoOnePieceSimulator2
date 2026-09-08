import { Apple, Plus, TriangleAlert } from '@tamagui/lucide-icons-2'
import { useEffect, useRef } from 'react'
import type { Control, FieldErrors } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import type { View } from 'react-native'
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
import type {
  DevilFruitFormValues,
  DevilFruitResponse,
} from '@/features/devil-fruits/types/devil-fruits.types'

import { DevilFruitCard } from './devil-fruit-card'
import { DevilFruitDetail } from './devil-fruit-detail'
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
  activeLocale: Locale
  onLocaleChange: (locale: Locale) => void
  erroredLocales: Locale[]
}

type BaseProps = {
  devilFruits: DevilFruitResponse[]
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  search: string
  onSearchChange: (search: string) => void
  rarityFilter: string | null
  rarityFilterOptions: GlassSelectOption[]
  onRarityFilterChange: (rarity: string | null) => void
  fruitTypeFilter: string | null
  fruitTypeFilterOptions: GlassSelectOption[]
  onFruitTypeFilterChange: (fruitType: string | null) => void
  hasActiveFilters: boolean
  detailFruit: DevilFruitResponse | null
  onOpenDetail: (devilFruit: DevilFruitResponse) => void
  onCloseDetail: () => void
  // Pagination is opt-in - see StandsScreen's identical props for the full
  // doc (admin screen omits all of these and renders no "Cargar más").
  hasNextPage?: boolean
  isFetchingNextPage?: boolean
  isLoadMoreError?: boolean
  onLoadMore?: () => void
  total?: number
}

type WritableProps = {
  readOnly?: false
  onCreateNew: () => void
  onEdit: (devilFruit: DevilFruitResponse) => void
  onDelete: (devilFruit: DevilFruitResponse) => void
  openingEditId: string | null
  form: FormState
  deleteConfirm: ConfirmState
}

type ReadOnlyProps = {
  readOnly: true
}

type Props = BaseProps & (WritableProps | ReadOnlyProps)

// Pure UI — same shape as StandsScreen (see that file's doc comment); a
// card grid plus the create/edit modal, delete confirmation and the
// read-only detail modal. All data fetching/mutation wiring lives in the
// container.
export function DevilFruitsScreen(props: Props) {
  const {
    devilFruits,
    isLoading,
    isError,
    onRetry,
    search,
    onSearchChange,
    rarityFilter,
    rarityFilterOptions,
    onRarityFilterChange,
    fruitTypeFilter,
    fruitTypeFilterOptions,
    onFruitTypeFilterChange,
    hasActiveFilters,
    detailFruit,
    onOpenDetail,
    onCloseDetail,
    hasNextPage,
    isFetchingNextPage,
    isLoadMoreError,
    onLoadMore,
    total,
  } = props
  const { t } = useTranslation()

  // Focus management for "Cargar más" - see StandsScreen's identical effect
  // for the full doc.
  const cardRefs = useRef(new Map<string, View | null>())
  const prevLengthRef = useRef(devilFruits.length)
  const wasFetchingRef = useRef(isFetchingNextPage ?? false)
  useEffect(() => {
    const wasFetching = wasFetchingRef.current
    wasFetchingRef.current = isFetchingNextPage ?? false
    if (wasFetching && !isFetchingNextPage && devilFruits.length > prevLengthRef.current) {
      const firstNew = devilFruits[prevLengthRef.current]
      const el = firstNew ? cardRefs.current.get(firstNew.id) : null
      if (el && typeof (el as unknown as { focus?: () => void }).focus === 'function') {
        ;(el as unknown as { focus: () => void }).focus()
      }
    }
    prevLengthRef.current = devilFruits.length
  }, [devilFruits, isFetchingNextPage])

  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">{t('devilFruits.title')}</GlowText>
          {props.readOnly ? null : (
            <GlossButton
              tone="green"
              btnSize="md"
              onPress={props.onCreateNew}
              accessibilityLabel={t('devilFruits.newDevilFruit')}
            >
              <Plus size={18} color="white" /> {t('devilFruits.newDevilFruit')}
            </GlossButton>
          )}
        </XStack>

        <XStack width="100%" flexWrap="wrap" gap="$3">
          <YStack flexBasis={220} grow={1}>
            <GlassField
              label={t('common.search')}
              value={search}
              onChangeText={onSearchChange}
              placeholder={t('devilFruits.searchPlaceholder')}
            />
          </YStack>
          <YStack flexBasis={200} grow={1}>
            <GlassSelect
              label={t('devilFruits.filterRarity')}
              options={rarityFilterOptions}
              value={rarityFilter}
              onChange={onRarityFilterChange}
              clearable
            />
          </YStack>
          <YStack flexBasis={200} grow={1}>
            <GlassSelect
              label={t('devilFruits.filterFruitType')}
              options={fruitTypeFilterOptions}
              value={fruitTypeFilter}
              onChange={onFruitTypeFilterChange}
              clearable
            />
          </YStack>
        </XStack>

        {isLoading ? (
          <YStack width="100%" items="center" p="$6">
            <Spinner size="large" />
          </YStack>
        ) : isError ? (
          // Distinct from the "no Devil Fruits yet" empty state below - see
          // StandsScreen's identical branch for why this can't be collapsed
          // into the empty state.
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <TriangleAlert size={28} color="$strawHatRed" />
            <GlowText level="label" align="center">
              {t('devilFruits.errorTitle')}
            </GlowText>
            <GlossButton
              tone="blue"
              btnSize="sm"
              onPress={onRetry}
              accessibilityLabel={t('devilFruits.retry')}
            >
              {t('devilFruits.retry')}
            </GlossButton>
          </GlassPanel>
        ) : devilFruits.length === 0 ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <Apple size={28} color="$strawHatRed" />
            <GlowText level="label" align="center">
              {t(hasActiveFilters ? 'devilFruits.emptyFilteredTitle' : 'devilFruits.emptyTitle')}
            </GlowText>
            {hasActiveFilters || props.readOnly ? null : (
              <GlossButton
                tone="green"
                btnSize="sm"
                onPress={props.onCreateNew}
                accessibilityLabel={t('devilFruits.newDevilFruit')}
              >
                {t('devilFruits.newDevilFruit')}
              </GlossButton>
            )}
          </GlassPanel>
        ) : (
          <>
            <XStack flexWrap="wrap" gap="$4" justify="center">
              {devilFruits.map((devilFruit) =>
                props.readOnly ? (
                  <DevilFruitCard
                    key={devilFruit.id}
                    ref={(el) => {
                      cardRefs.current.set(devilFruit.id, el)
                    }}
                    devilFruit={devilFruit}
                    onOpenDetail={() => onOpenDetail(devilFruit)}
                    readOnly
                  />
                ) : (
                  <DevilFruitCard
                    key={devilFruit.id}
                    ref={(el) => {
                      cardRefs.current.set(devilFruit.id, el)
                    }}
                    devilFruit={devilFruit}
                    onOpenDetail={() => onOpenDetail(devilFruit)}
                    onEdit={() => props.onEdit(devilFruit)}
                    onDelete={() => props.onDelete(devilFruit)}
                    isEditBusy={props.openingEditId === devilFruit.id}
                  />
                )
              )}
            </XStack>

            {onLoadMore ? (
              <YStack width="100%" items="center" gap="$2" py="$4">
                {hasNextPage ? (
                  <>
                    <GlossButton
                      tone={isLoadMoreError ? 'orange' : 'blue'}
                      btnSize="md"
                      onPress={onLoadMore}
                      disabled={isFetchingNextPage}
                      accessibilityLabel={t(isLoadMoreError ? 'common.loadMoreRetry' : 'common.loadMore')}
                    >
                      {isFetchingNextPage ? (
                        <Spinner size="small" color="white" />
                      ) : (
                        t(isLoadMoreError ? 'common.loadMoreRetry' : 'common.loadMore')
                      )}
                    </GlossButton>
                    {typeof total === 'number' ? (
                      <GlowText level="label" tone="soft">
                        {t('common.itemsLoadedOfTotal', { loaded: devilFruits.length, total })}
                      </GlowText>
                    ) : null}
                  </>
                ) : typeof total === 'number' ? (
                  <GlowText level="label" tone="soft">
                    {t('common.allItemsLoaded', { total })}
                  </GlowText>
                ) : null}
              </YStack>
            ) : null}
          </>
        )}
      </PageShell>

      <DetailModal
        visible={detailFruit !== null}
        title={detailFruit?.name ?? ''}
        onClose={onCloseDetail}
        closeA11y={t('common.close')}
        footer={
          detailFruit && !props.readOnly ? (
            <GlossButton
              tone="blue"
              btnSize="md"
              onPress={() => {
                onCloseDetail()
                props.onEdit(detailFruit)
              }}
              accessibilityLabel={t('devilFruits.editA11y', { name: detailFruit.name })}
            >
              {t('common.edit')}
            </GlossButton>
          ) : undefined
        }
      >
        {detailFruit ? <DevilFruitDetail devilFruit={detailFruit} /> : null}
      </DetailModal>

      {props.readOnly ? null : (
        <>
          <DevilFruitFormModal
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
            title={t('devilFruits.deleteConfirmTitle')}
            message={t('devilFruits.deleteConfirmMessage', { name: props.deleteConfirm.fruitName ?? '' })}
            confirmLabel={t('devilFruits.deleteConfirmButton')}
            isConfirming={props.deleteConfirm.isConfirming}
            onConfirm={props.deleteConfirm.onConfirm}
            onCancel={props.deleteConfirm.onCancel}
          />
        </>
      )}
    </YStack>
  )
}
