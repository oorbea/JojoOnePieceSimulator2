import { Plus, TriangleAlert, Users } from '@tamagui/lucide-icons-2'
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
import type { PictureStatus, PowerRarity, Locale } from '@/shared/contracts/enums'
import { focusElement } from '@/shared/lib/a11y'
import type {
  CharacterKind,
  CharacterMangaFilter,
  JojoCharacterFormValues,
  OnePieceCharacterFormValues,
} from '@/features/characters/types/characters.types'
import type { CharacterStatRow } from '@/features/characters/lib/character-stats'

import { CharacterCard } from './character-card'
import { CharacterDetail } from './character-detail'
import { CharacterFormModal } from './character-form-modal'

type CharacterLike = {
  id: string
  name: string
  description: string
  rarity: PowerRarity
  picture: string
  pictureThumb: string
  pictureCard?: string
  pictureLqip?: string
  pictureStatus: PictureStatus
}

type ConfirmState = {
  visible: boolean
  isConfirming: boolean
  onConfirm: () => void
  onCancel: () => void
  characterName?: string
}

type FormState = {
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

type BaseProps<T extends CharacterLike> = {
  characters: T[]
  // Either a fixed descriptor (single-kind screens) or one resolved per
  // item - the latter is what lets the 'ALL' filter mix both kinds' cards
  // in one grid without CharacterCard/CharacterDetail knowing about kinds.
  rows: CharacterStatRow<T>[] | ((character: T) => CharacterStatRow<T>[])
  isLoading: boolean
  isError: boolean
  onRetry: () => void
  search: string
  onSearchChange: (search: string) => void
  mangaFilter: CharacterMangaFilter
  onMangaFilterChange: (manga: CharacterMangaFilter) => void
  hasActiveFilters: boolean
  detailCharacter: T | null
  onOpenDetail: (character: T) => void
  onCloseDetail: () => void
  // Pagination is opt-in - the admin screen omits it, the catalog screen
  // wires it through usePaginatedCatalogue - same split as StagesScreen.
  hasNextPage?: boolean
  isFetchingNextPage?: boolean
  isLoadMoreError?: boolean
  onLoadMore?: () => void
  total?: number
}

type WritableProps<T extends CharacterLike> = {
  readOnly?: false
  onCreateNew: () => void
  onEdit: (character: T) => void
  onDelete: (character: T) => void
  openingEditId: string | null
  form: FormState
  deleteConfirm: ConfirmState
}

type ReadOnlyProps = {
  readOnly: true
}

type Props<T extends CharacterLike> = BaseProps<T> & (WritableProps<T> | ReadOnlyProps)

const MANGA_FILTER_OPTIONS: GlassSelectOption[] = [
  { value: 'ALL', label: 'characters.filterMangaAll' },
  { value: 'JOJO', label: 'enums.manga.JOJO' },
  { value: 'ONE_PIECE', label: 'enums.manga.ONE_PIECE' },
]

// Pure UI - a card grid of Characters plus the create/edit modal, delete
// confirmation, and read-only detail modal. The manga filter has three
// values: JoJo, One Piece, or 'ALL' to mix both kinds in one grid.
// JojoCharacter and OnePieceCharacter are two separate backend endpoints
// with independent keyset pagination, so 'ALL' means the container fetches
// both and merges the results client-side rather than the backend ever
// returning a mixed page - see CatalogCharactersContainer/
// CharactersContainer for that merge. Kind-agnostic itself:
// CharactersContainer/CatalogCharactersContainer decide which kind(s)'
// data and `rows` descriptor (character-stats.ts) to pass in. Same shape
// as StagesScreen otherwise.
export function CharactersScreen<T extends CharacterLike>(props: Props<T>) {
  const {
    characters,
    rows,
    isLoading,
    isError,
    onRetry,
    search,
    onSearchChange,
    mangaFilter,
    onMangaFilterChange,
    hasActiveFilters,
    detailCharacter,
    onOpenDetail,
    onCloseDetail,
    hasNextPage,
    isFetchingNextPage,
    isLoadMoreError,
    onLoadMore,
    total,
  } = props
  const { t } = useTranslation()

  const mangaFilterOptions = MANGA_FILTER_OPTIONS.map((o) => ({ ...o, label: t(o.label) }))
  const rowsFor = (character: T): CharacterStatRow<T>[] =>
    typeof rows === 'function' ? rows(character) : rows

  // Focus management for "Cargar más" - see StagesScreen's identical effect.
  const cardRefs = useRef(new Map<string, View | null>())
  const prevLengthRef = useRef(characters.length)
  const wasFetchingRef = useRef(isFetchingNextPage ?? false)
  useEffect(() => {
    const wasFetching = wasFetchingRef.current
    wasFetchingRef.current = isFetchingNextPage ?? false
    if (wasFetching && !isFetchingNextPage && characters.length > prevLengthRef.current) {
      const firstNew = characters[prevLengthRef.current]
      focusElement(firstNew ? (cardRefs.current.get(firstNew.id) ?? null) : null)
    }
    prevLengthRef.current = characters.length
  }, [characters, isFetchingNextPage])

  return (
    <YStack flex={1} position="relative">
      <PageShell align="top" scroll maxWidth={960}>
        <XStack width="100%" items="center" justify="space-between" flexWrap="wrap" gap="$3">
          <GlowText level="title">{t('characters.title')}</GlowText>
          {props.readOnly ? null : (
            <GlossButton
              tone="green"
              btnSize="md"
              onPress={props.onCreateNew}
              accessibilityLabel={t('characters.newCharacter')}
            >
              <Plus size={18} color="white" /> {t('characters.newCharacter')}
            </GlossButton>
          )}
        </XStack>

        <XStack width="100%" flexWrap="wrap" gap="$3">
          <YStack flexBasis={220} grow={1}>
            <GlassField
              label={t('common.search')}
              value={search}
              onChangeText={onSearchChange}
              placeholder={t('characters.searchPlaceholder')}
            />
          </YStack>
          <YStack flexBasis={200} grow={1}>
            <GlassSelect
              label={t('characters.filterManga')}
              options={mangaFilterOptions}
              value={mangaFilter}
              onChange={(v) => onMangaFilterChange((v ?? 'JOJO') as CharacterMangaFilter)}
            />
          </YStack>
        </XStack>

        {isLoading ? (
          <YStack width="100%" items="center" p="$6">
            <Spinner size="large" />
          </YStack>
        ) : isError ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <TriangleAlert size={28} color="$strawHatRed" />
            <GlowText level="label" align="center">
              {t('characters.errorTitle')}
            </GlowText>
            <GlossButton
              tone="blue"
              btnSize="sm"
              onPress={onRetry}
              accessibilityLabel={t('characters.retry')}
            >
              {t('characters.retry')}
            </GlossButton>
          </GlassPanel>
        ) : characters.length === 0 ? (
          <GlassPanel tone="plastic" elevate={0} width="100%" p="$6" gap="$3" items="center">
            <Users size={28} color="$wiiBlue" />
            <GlowText level="label" align="center">
              {t(hasActiveFilters ? 'characters.emptyFilteredTitle' : 'characters.emptyTitle')}
            </GlowText>
            {hasActiveFilters || props.readOnly ? null : (
              <GlossButton
                tone="green"
                btnSize="sm"
                onPress={props.onCreateNew}
                accessibilityLabel={t('characters.newCharacter')}
              >
                {t('characters.newCharacter')}
              </GlossButton>
            )}
          </GlassPanel>
        ) : (
          <>
            <XStack flexWrap="wrap" gap="$4" justify="center">
              {characters.map((character) =>
                props.readOnly ? (
                  <CharacterCard
                    key={character.id}
                    ref={(el) => {
                      cardRefs.current.set(character.id, el)
                    }}
                    character={character}
                    rows={rowsFor(character)}
                    onOpenDetail={() => onOpenDetail(character)}
                    readOnly
                  />
                ) : (
                  <CharacterCard
                    key={character.id}
                    ref={(el) => {
                      cardRefs.current.set(character.id, el)
                    }}
                    character={character}
                    rows={rowsFor(character)}
                    onOpenDetail={() => onOpenDetail(character)}
                    onEdit={() => props.onEdit(character)}
                    onDelete={() => props.onDelete(character)}
                    isEditBusy={props.openingEditId === character.id}
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
                      accessibilityLabel={t(
                        isLoadMoreError ? 'common.loadMoreRetry' : 'common.loadMore'
                      )}
                    >
                      {isFetchingNextPage ? (
                        <Spinner size="small" color="white" />
                      ) : (
                        t(isLoadMoreError ? 'common.loadMoreRetry' : 'common.loadMore')
                      )}
                    </GlossButton>
                    {typeof total === 'number' ? (
                      <GlowText level="label" tone="soft">
                        {t('common.itemsLoadedOfTotal', { loaded: characters.length, total })}
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
        visible={detailCharacter !== null}
        title={detailCharacter?.name ?? ''}
        onClose={onCloseDetail}
        closeA11y={t('common.close')}
        footer={
          detailCharacter && !props.readOnly ? (
            <GlossButton
              tone="blue"
              btnSize="md"
              onPress={() => {
                onCloseDetail()
                props.onEdit(detailCharacter)
              }}
              accessibilityLabel={t('characters.editA11y', { name: detailCharacter.name })}
            >
              {t('common.edit')}
            </GlossButton>
          ) : undefined
        }
      >
        {detailCharacter ? (
          <CharacterDetail character={detailCharacter} rows={rowsFor(detailCharacter)} />
        ) : null}
      </DetailModal>

      {props.readOnly ? null : (
        <>
          <CharacterFormModal
            visible={props.form.visible}
            mode={props.form.mode}
            kind={props.form.kind}
            onSelectKind={props.form.onSelectKind}
            jojoControl={props.form.jojoControl}
            jojoErrors={props.form.jojoErrors}
            onePieceControl={props.form.onePieceControl}
            onePieceErrors={props.form.onePieceErrors}
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
            title={t('characters.deleteConfirmTitle')}
            message={t('characters.deleteConfirmMessage', {
              name: props.deleteConfirm.characterName ?? '',
            })}
            confirmLabel={t('characters.deleteConfirmButton')}
            isConfirming={props.deleteConfirm.isConfirming}
            onConfirm={props.deleteConfirm.onConfirm}
            onCancel={props.deleteConfirm.onCancel}
          />
        </>
      )}
    </YStack>
  )
}
