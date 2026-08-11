import { Apple, Plus, TriangleAlert } from '@tamagui/lucide-icons-2'
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
import type { Locale } from '@/shared/lib/zod'
import type {
  DevilFruitFormValues,
  DevilFruitResponse,
} from '@/features/devil-fruits/types/devil-fruits.types'

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
  activeLocale: Locale
  onLocaleChange: (locale: Locale) => void
  erroredLocales: Locale[]
}

type Props = {
  devilFruits: DevilFruitResponse[]
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  onCreateNew: () => void
  onEdit: (devilFruit: DevilFruitResponse) => void
  onDelete: (devilFruit: DevilFruitResponse) => void
  openingEditId: string | null
  search: string
  onSearchChange: (search: string) => void
  rarityFilter: string | null
  rarityFilterOptions: GlassSelectOption[]
  onRarityFilterChange: (rarity: string | null) => void
  fruitTypeFilter: string | null
  fruitTypeFilterOptions: GlassSelectOption[]
  onFruitTypeFilterChange: (fruitType: string | null) => void
  hasActiveFilters: boolean
  form: FormState
  deleteConfirm: ConfirmState
}

export function DevilFruitsScreen({
  devilFruits,
  isLoading,
  isError,
  onRetry,
  onCreateNew,
  onEdit,
  onDelete,
  openingEditId,
  search,
  onSearchChange,
  rarityFilter,
  rarityFilterOptions,
  onRarityFilterChange,
  fruitTypeFilter,
  fruitTypeFilterOptions,
  onFruitTypeFilterChange,
  hasActiveFilters,
  form,
  deleteConfirm,
}: Props) {
  const { t } = useTranslation()
  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">{t('devilFruits.title')}</GlowText>
          <GlossButton
            tone="green"
            btnSize="md"
            onPress={onCreateNew}
            accessibilityLabel={t('devilFruits.newDevilFruit')}
          >
            <Plus size={18} color="white" /> {t('devilFruits.newDevilFruit')}
          </GlossButton>
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
            {hasActiveFilters ? null : (
              <GlossButton
                tone="green"
                btnSize="sm"
                onPress={onCreateNew}
                accessibilityLabel={t('devilFruits.newDevilFruit')}
              >
                {t('devilFruits.newDevilFruit')}
              </GlossButton>
            )}
          </GlassPanel>
        ) : (
          <XStack flexWrap="wrap" gap="$4" justify="center">
            {devilFruits.map((devilFruit) => (
              <DevilFruitCard
                key={devilFruit.id}
                devilFruit={devilFruit}
                onEdit={() => onEdit(devilFruit)}
                onDelete={() => onDelete(devilFruit)}
                isEditBusy={openingEditId === devilFruit.id}
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
        activeLocale={form.activeLocale}
        onLocaleChange={form.onLocaleChange}
        erroredLocales={form.erroredLocales}
      />

      <ConfirmSheet
        visible={deleteConfirm.visible}
        title={t('devilFruits.deleteConfirmTitle')}
        message={t('devilFruits.deleteConfirmMessage', { name: deleteConfirm.fruitName ?? '' })}
        confirmLabel={t('devilFruits.deleteConfirmButton')}
        isConfirming={deleteConfirm.isConfirming}
        onConfirm={deleteConfirm.onConfirm}
        onCancel={deleteConfirm.onCancel}
      />
    </YStack>
  )
}
