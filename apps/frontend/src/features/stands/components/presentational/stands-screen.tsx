import { Plus, Sparkles, TriangleAlert } from '@tamagui/lucide-icons-2'
import { useEffect, useRef } from 'react'
import type { Control, FieldErrors } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import type { View } from 'react-native'
import { Spinner, XStack, YStack } from 'tamagui'

import { ConfirmSheet } from '@/shared/components/presentational/confirm-sheet'
import { DetailModal } from '@/shared/components/presentational/detail-modal'
import { FilterDisclosure } from '@/shared/components/presentational/filter-disclosure'
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
import type { StandFormValues, StandResponse } from '@/features/stands/types/stands.types'

import { StandCard } from './stand-card'
import { StandDetail } from './stand-detail'
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

export type StandStatFilterKey =
  'attackPower' | 'speed' | 'attackRange' | 'endurance' | 'precision' | 'potential'

const STAT_FILTER_ORDER: StandStatFilterKey[] = [
  'attackPower',
  'speed',
  'attackRange',
  'endurance',
  'precision',
  'potential',
]

type BaseProps = {
  stands: StandResponse[]
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  search: string
  onSearchChange: (search: string) => void
  rarityFilter: string | null
  rarityFilterOptions: GlassSelectOption[]
  onRarityFilterChange: (rarity: string | null) => void
  statFilters: Record<StandStatFilterKey, string | null>
  statFilterOptions: GlassSelectOption[]
  onStatFilterChange: (key: StandStatFilterKey, value: string | null) => void
  evolvesFromFilter: string | null
  evolvesFromFilterOptions: GlassSelectOption[]
  onEvolvesFromFilterChange: (id: string | null) => void
  filtersExpanded: boolean
  onToggleFilters: () => void
  moreFiltersCount: number
  onClearFilters: () => void
  hasActiveFilters: boolean
  detailStand: StandResponse | null
  onOpenDetail: (stand: StandResponse) => void
  onCloseDetail: () => void
  // Pagination is opt-in: omitting all of these (the admin container's
  // case, which deliberately keeps the full unpaginated fetch - see
  // ObsidianVault/entrega-imagenes-red-lenta-2026-09-07.md's T1.4 note)
  // renders no "Cargar más" section at all.
  hasNextPage?: boolean
  isFetchingNextPage?: boolean
  isLoadMoreError?: boolean
  onLoadMore?: () => void
  total?: number
}

type WritableProps = {
  readOnly?: false
  onCreateNew: () => void
  onEdit: (stand: StandResponse) => void
  onDelete: (stand: StandResponse) => void
  openingEditId: string | null
  form: FormState
  deleteConfirm: ConfirmState
}

type ReadOnlyProps = {
  readOnly: true
}

type Props = BaseProps & (WritableProps | ReadOnlyProps)

// Pure UI — a card grid of Stands plus the create/edit modal and the delete
// confirmation. All data fetching, form state, and mutation wiring live in
// StandsContainer. Search + rarity stay always visible; the six stat
// filters and evolvesFrom live behind a FilterDisclosure so the always-on
// row doesn't grow to nine controls.
export function StandsScreen(props: Props) {
  const {
    stands,
    isLoading,
    isError,
    onRetry,
    search,
    onSearchChange,
    rarityFilter,
    rarityFilterOptions,
    onRarityFilterChange,
    statFilters,
    statFilterOptions,
    onStatFilterChange,
    evolvesFromFilter,
    evolvesFromFilterOptions,
    onEvolvesFromFilterChange,
    filtersExpanded,
    onToggleFilters,
    moreFiltersCount,
    onClearFilters,
    hasActiveFilters,
    detailStand,
    onOpenDetail,
    onCloseDetail,
    hasNextPage,
    isFetchingNextPage,
    isLoadMoreError,
    onLoadMore,
    total,
  } = props
  const { t } = useTranslation()

  // Focus management for "Cargar más": after a successful append, move
  // focus to the first newly-added card so a keyboard user isn't left on a
  // button that may have moved or unmounted (norma-teclado.md) - see
  // ObsidianVault/entrega-imagenes-red-lenta-2026-09-07.md's T3.10 note.
  const cardRefs = useRef(new Map<string, View | null>())
  const prevLengthRef = useRef(stands.length)
  const wasFetchingRef = useRef(isFetchingNextPage ?? false)
  useEffect(() => {
    const wasFetching = wasFetchingRef.current
    wasFetchingRef.current = isFetchingNextPage ?? false
    if (wasFetching && !isFetchingNextPage && stands.length > prevLengthRef.current) {
      const firstNew = stands[prevLengthRef.current]
      const el = firstNew ? cardRefs.current.get(firstNew.id) : null
      // Native RN Views have no DOM-style .focus() (unlike react-native-web's
      // host node) - moving accessibility focus there needs
      // AccessibilityInfo.setAccessibilityFocus, out of scope for this pass.
      // Guarding on the method's existence keeps native a no-op instead of a
      // crash while still fixing the actual reported bug's platform (web).
      if (el && typeof (el as unknown as { focus?: () => void }).focus === 'function') {
        ;(el as unknown as { focus: () => void }).focus()
      }
    }
    prevLengthRef.current = stands.length
  }, [stands, isFetchingNextPage])

  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">{t('stands.title')}</GlowText>
          {props.readOnly ? null : (
            <GlossButton
              tone="green"
              btnSize="md"
              onPress={props.onCreateNew}
              accessibilityLabel={t('stands.newStand')}
            >
              <Plus size={18} color="white" /> {t('stands.newStand')}
            </GlossButton>
          )}
        </XStack>

        <XStack width="100%" flexWrap="wrap" gap="$3">
          <YStack flexBasis={220} grow={1}>
            <GlassField
              label={t('common.search')}
              value={search}
              onChangeText={onSearchChange}
              placeholder={t('stands.searchPlaceholder')}
            />
          </YStack>
          <YStack flexBasis={200} grow={1}>
            <GlassSelect
              label={t('stands.filterRarity')}
              options={rarityFilterOptions}
              value={rarityFilter}
              onChange={onRarityFilterChange}
              clearable
            />
          </YStack>
        </XStack>

        <FilterDisclosure
          label={t('stands.moreFilters')}
          activeCount={moreFiltersCount}
          expanded={filtersExpanded}
          onToggle={onToggleFilters}
          onClearAll={hasActiveFilters ? onClearFilters : undefined}
          clearLabel={t('stands.clearFilters')}
        >
          {STAT_FILTER_ORDER.map((key) => (
            <YStack key={key} flexBasis={200} grow={1}>
              <GlassSelect
                label={t(`stands.stats.${key}`)}
                options={statFilterOptions}
                value={statFilters[key]}
                onChange={(value) => onStatFilterChange(key, value)}
                clearable
              />
            </YStack>
          ))}
          <YStack flexBasis={200} grow={1}>
            <GlassSelect
              label={t('stands.filterEvolvesFrom')}
              options={evolvesFromFilterOptions}
              value={evolvesFromFilter}
              onChange={onEvolvesFromFilterChange}
              searchable
              clearable
            />
          </YStack>
        </FilterDisclosure>

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
            <GlossButton
              tone="blue"
              btnSize="sm"
              onPress={onRetry}
              accessibilityLabel={t('stands.retry')}
            >
              {t('stands.retry')}
            </GlossButton>
          </GlassPanel>
        ) : stands.length === 0 ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <Sparkles size={28} color="$standPurple" />
            <GlowText level="label" align="center">
              {t(hasActiveFilters ? 'stands.emptyFilteredTitle' : 'stands.emptyTitle')}
            </GlowText>
            {hasActiveFilters || props.readOnly ? null : (
              <GlossButton
                tone="green"
                btnSize="sm"
                onPress={props.onCreateNew}
                accessibilityLabel={t('stands.newStand')}
              >
                {t('stands.newStand')}
              </GlossButton>
            )}
          </GlassPanel>
        ) : (
          <>
            <XStack flexWrap="wrap" gap="$4" justify="center">
              {stands.map((stand) =>
                props.readOnly ? (
                  <StandCard
                    key={stand.id}
                    ref={(el) => {
                      cardRefs.current.set(stand.id, el)
                    }}
                    stand={stand}
                    onOpenDetail={() => onOpenDetail(stand)}
                    readOnly
                  />
                ) : (
                  <StandCard
                    key={stand.id}
                    ref={(el) => {
                      cardRefs.current.set(stand.id, el)
                    }}
                    stand={stand}
                    onOpenDetail={() => onOpenDetail(stand)}
                    onEdit={() => props.onEdit(stand)}
                    onDelete={() => props.onDelete(stand)}
                    isEditBusy={props.openingEditId === stand.id}
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
                        {t('common.itemsLoadedOfTotal', { loaded: stands.length, total })}
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
        visible={detailStand !== null}
        title={detailStand?.name ?? ''}
        onClose={onCloseDetail}
        closeA11y={t('common.close')}
        footer={
          detailStand && !props.readOnly ? (
            <GlossButton
              tone="blue"
              btnSize="md"
              onPress={() => {
                onCloseDetail()
                props.onEdit(detailStand)
              }}
              accessibilityLabel={t('stands.editA11y', { name: detailStand.name })}
            >
              {t('common.edit')}
            </GlossButton>
          ) : undefined
        }
      >
        {detailStand ? <StandDetail stand={detailStand} /> : null}
      </DetailModal>

      {props.readOnly ? null : (
        <>
          <StandFormModal
            visible={props.form.visible}
            mode={props.form.mode}
            control={props.form.control}
            errors={props.form.errors}
            onSubmit={props.form.onSubmit}
            onCancel={props.form.onCancel}
            isSaving={props.form.isSaving}
            evolvesFromOptions={props.form.evolvesFromOptions}
            pictureUri={props.form.pictureUri}
            onPickPicture={props.form.onPickPicture}
            isPictureBusy={props.form.isPictureBusy}
            activeLocale={props.form.activeLocale}
            onLocaleChange={props.form.onLocaleChange}
            erroredLocales={props.form.erroredLocales}
          />

          <ConfirmSheet
            visible={props.deleteConfirm.visible}
            title={t('stands.deleteConfirmTitle')}
            message={t('stands.deleteConfirmMessage', { name: props.deleteConfirm.standName ?? '' })}
            confirmLabel={t('stands.deleteConfirmButton')}
            isConfirming={props.deleteConfirm.isConfirming}
            onConfirm={props.deleteConfirm.onConfirm}
            onCancel={props.deleteConfirm.onCancel}
          />
        </>
      )}
    </YStack>
  )
}
