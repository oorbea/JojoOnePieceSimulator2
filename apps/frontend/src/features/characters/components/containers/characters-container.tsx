import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'burnt'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import {
  getJojoCharacterTranslations,
  getOnePieceCharacterTranslations,
} from '@/features/characters/api/characters.api'
import { characterKeys } from '@/features/characters/api/characters.keys'
import { CharactersScreen } from '@/features/characters/components/presentational/characters-screen'
import {
  useCreateJojoCharacter,
  useCreateOnePieceCharacter,
  useDeleteJojoCharacter,
  useDeleteOnePieceCharacter,
  useUpdateJojoCharacter,
  useUpdateOnePieceCharacter,
  useUploadJojoCharacterPicture,
  useUploadOnePieceCharacterPicture,
} from '@/features/characters/hooks/use-character-mutations'
import {
  useJojoCharacters,
  useOnePieceCharacters,
} from '@/features/characters/hooks/use-characters'
import { rowsForCharacter } from '@/features/characters/lib/character-stats'
import {
  jojoCharacterFormSchema,
  onePieceCharacterFormSchema,
  type CharacterKind,
  type CharacterMangaFilter,
  type JojoCharacterFormValues,
  type JojoCharacterInput,
  type OnePieceCharacterFormValues,
  type OnePieceCharacterInput,
  type TaggedCharacter,
} from '@/features/characters/types/characters.types'
import { useDebouncedValue } from '@/shared/hooks/use-debounced-value'
import type { PickedPicture } from '@/shared/hooks/use-picture-picker'
import { usePicturePicker } from '@/shared/hooks/use-picture-picker'
import { DEFAULT_LOCALE, SUPPORTED_LOCALES } from '@/shared/i18n'
import {
  createEmptyCharacterTranslationsForm,
  fromCharacterTranslationsResponse,
  toCharacterTranslationsPayload,
} from '@/shared/lib/character-translations'
import type { Locale } from '@/shared/contracts/enums'

function createDefaultJojoValues(): JojoCharacterFormValues {
  return {
    name: '',
    rarity: 'COMMON',
    hamon: 'NONE',
    spin: 'NONE',
    battleIq: 100,
    translations: createEmptyCharacterTranslationsForm(),
  }
}

function createDefaultOnePieceValues(): OnePieceCharacterFormValues {
  return {
    name: '',
    rarity: 'COMMON',
    physicalForm: 'PRIVATE',
    armamentHaki: 'NONE',
    observationHaki: 'NONE',
    conquerorHaki: 'NONE',
    fruitMastery: 'NONE',
    translations: createEmptyCharacterTranslationsForm(),
  }
}

function toJojoInput(values: JojoCharacterFormValues): JojoCharacterInput {
  const { translations, ...rest } = values
  return { ...rest, translations: toCharacterTranslationsPayload(translations) }
}

function toOnePieceInput(values: OnePieceCharacterFormValues): OnePieceCharacterInput {
  const { translations, ...rest } = values
  return { ...rest, translations: toCharacterTranslationsPayload(translations) }
}

// Admin CRUD for both character kinds behind one manga toggle - JoJo, One
// Piece, or 'ALL' to mix both kinds' cards in one grid (see CharactersScreen's
// doc for why 'ALL' means merging two client-side lists rather than a mixed
// backend page). Both list queries and both react-hook-form instances are
// always mounted (hooks can't be called conditionally); `enabled` on each
// query tracks whether its kind is currently wanted, and only the kind
// picked in the create/edit modal's own selector (`modalState.kind`, not
// the grid's `mangaFilter`) has its form rendered/submitted - see
// useJojoCharacters' doc on the `enabled` param.
export function CharactersContainer() {
  const { t } = useTranslation()
  const [mangaFilter, setMangaFilter] = useState<CharacterMangaFilter>('JOJO')
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search)
  const queryClient = useQueryClient()
  const { pickPicture } = usePicturePicker()

  const filters = useMemo(() => {
    const f: { q?: string } = {}
    if (debouncedSearch.trim()) f.q = debouncedSearch.trim()
    return f
  }, [debouncedSearch])
  const hasActiveFilters = Object.keys(filters).length > 0
  const appliedFilters = hasActiveFilters ? filters : undefined

  const wantJojo = mangaFilter !== 'ONE_PIECE'
  const wantOnePiece = mangaFilter !== 'JOJO'
  const jojoQuery = useJojoCharacters(appliedFilters, wantJojo)
  const onePieceQuery = useOnePieceCharacters(appliedFilters, wantOnePiece)

  const createJojo = useCreateJojoCharacter()
  const updateJojo = useUpdateJojoCharacter()
  const deleteJojo = useDeleteJojoCharacter()
  const uploadJojoPicture = useUploadJojoCharacterPicture()

  const createOnePiece = useCreateOnePieceCharacter()
  const updateOnePiece = useUpdateOnePieceCharacter()
  const deleteOnePiece = useDeleteOnePieceCharacter()
  const uploadOnePiecePicture = useUploadOnePieceCharacterPicture()

  const jojoForm = useForm<JojoCharacterFormValues>({
    resolver: zodResolver(jojoCharacterFormSchema),
    defaultValues: createDefaultJojoValues(),
  })
  const onePieceForm = useForm<OnePieceCharacterFormValues>({
    resolver: zodResolver(onePieceCharacterFormSchema),
    defaultValues: createDefaultOnePieceValues(),
  })

  const [modalState, setModalState] = useState<{
    visible: boolean
    mode: 'create' | 'edit'
    kind: CharacterKind | null
    editingId: string | null
  }>({ visible: false, mode: 'create', kind: null, editingId: null })

  const [activeLocale, setActiveLocale] = useState<Locale>(DEFAULT_LOCALE)
  const [pendingPicture, setPendingPicture] = useState<PickedPicture | null>(null)
  const [openingEditId, setOpeningEditId] = useState<string | null>(null)
  const [characterToDelete, setCharacterToDelete] = useState<{
    kind: CharacterKind
    id: string
    name: string
  } | null>(null)

  const activeCharacters = useMemo<TaggedCharacter[]>(() => {
    const jojoTagged = wantJojo
      ? (jojoQuery.data ?? []).map((c) => ({ ...c, kind: 'JOJO' as const }))
      : []
    const onePieceTagged = wantOnePiece
      ? (onePieceQuery.data ?? []).map((c) => ({ ...c, kind: 'ONE_PIECE' as const }))
      : []
    return [...jojoTagged, ...onePieceTagged]
  }, [wantJojo, wantOnePiece, jojoQuery.data, onePieceQuery.data])

  // isLoading/isError: OR'd across whichever kind(s) are currently wanted -
  // a disabled query's isLoading/isError are always false, so this collapses
  // to "the one active query's state" outside 'ALL' mode.
  const isLoading = (wantJojo && jojoQuery.isLoading) || (wantOnePiece && onePieceQuery.isLoading)
  const isError = (wantJojo && jojoQuery.isError) || (wantOnePiece && onePieceQuery.isError)
  const onRetry = () => {
    if (wantJojo) void jojoQuery.refetch()
    if (wantOnePiece) void onePieceQuery.refetch()
  }

  const [detailCharacter, setDetailCharacter] = useState<TaggedCharacter | null>(null)

  const openCreate = () => {
    jojoForm.reset(createDefaultJojoValues())
    onePieceForm.reset(createDefaultOnePieceValues())
    setPendingPicture(null)
    setActiveLocale(DEFAULT_LOCALE)
    setModalState({ visible: true, mode: 'create', kind: null, editingId: null })
  }

  const openEdit = async (character: TaggedCharacter) => {
    setOpeningEditId(character.id)
    try {
      if (character.kind === 'JOJO') {
        const jc = character
        const translations = await queryClient.fetchQuery({
          queryKey: characterKeys.jojo.translations(jc.id),
          queryFn: () => getJojoCharacterTranslations(jc.id),
        })
        jojoForm.reset({
          name: jc.name,
          rarity: jc.rarity,
          hamon: jc.hamon,
          spin: jc.spin,
          battleIq: jc.battleIq,
          translations: fromCharacterTranslationsResponse(translations),
        })
      } else {
        const oc = character
        const translations = await queryClient.fetchQuery({
          queryKey: characterKeys.onePiece.translations(oc.id),
          queryFn: () => getOnePieceCharacterTranslations(oc.id),
        })
        onePieceForm.reset({
          name: oc.name,
          rarity: oc.rarity,
          physicalForm: oc.physicalForm,
          armamentHaki: oc.armamentHaki,
          observationHaki: oc.observationHaki,
          conquerorHaki: oc.conquerorHaki,
          fruitMastery: oc.fruitMastery,
          translations: fromCharacterTranslationsResponse(translations),
        })
      }
      setPendingPicture(null)
      setActiveLocale(DEFAULT_LOCALE)
      setModalState({ visible: true, mode: 'edit', kind: character.kind, editingId: character.id })
    } finally {
      setOpeningEditId(null)
    }
  }

  const closeModal = () => setModalState((prev) => ({ ...prev, visible: false }))

  const onPickPicture = async () => {
    const asset = await pickPicture()
    if (asset) setPendingPicture(asset)
  }

  const jumpToFirstErroredLocale = (
    formErrors: typeof jojoForm.formState.errors | typeof onePieceForm.formState.errors
  ) => {
    const erroredLocale = SUPPORTED_LOCALES.find((locale) => formErrors.translations?.[locale])
    if (erroredLocale) setActiveLocale(erroredLocale)
  }

  const onSubmitJojo = jojoForm.handleSubmit((values) => {
    const input = toJojoInput(values)
    if (modalState.mode === 'create') {
      createJojo.mutate(input, {
        onSuccess: (created) => {
          if (pendingPicture) uploadJojoPicture.mutate({ id: created.id, asset: pendingPicture })
          closeModal()
        },
      })
      return
    }
    const id = modalState.editingId
    if (!id) return
    updateJojo.mutate(
      { id, input },
      {
        onSuccess: () => {
          if (pendingPicture) uploadJojoPicture.mutate({ id, asset: pendingPicture })
          closeModal()
        },
      }
    )
  }, jumpToFirstErroredLocale)

  const onSubmitOnePiece = onePieceForm.handleSubmit((values) => {
    const input = toOnePieceInput(values)
    if (modalState.mode === 'create') {
      createOnePiece.mutate(input, {
        onSuccess: (created) => {
          if (pendingPicture)
            uploadOnePiecePicture.mutate({ id: created.id, asset: pendingPicture })
          closeModal()
        },
      })
      return
    }
    const id = modalState.editingId
    if (!id) return
    updateOnePiece.mutate(
      { id, input },
      {
        onSuccess: () => {
          if (pendingPicture) uploadOnePiecePicture.mutate({ id, asset: pendingPicture })
          closeModal()
        },
      }
    )
  }, jumpToFirstErroredLocale)

  const onSubmit = () => {
    if (modalState.kind === 'JOJO') void onSubmitJojo()
    else if (modalState.kind === 'ONE_PIECE') void onSubmitOnePiece()
  }

  const onSelectKind = (kind: CharacterKind) => {
    setModalState((prev) => ({ ...prev, kind }))
  }

  const onConfirmDelete = () => {
    if (!characterToDelete) return
    if (characterToDelete.kind === 'JOJO') {
      deleteJojo.mutate(characterToDelete.id, { onSuccess: () => setCharacterToDelete(null) })
    } else {
      deleteOnePiece.mutate(characterToDelete.id, { onSuccess: () => setCharacterToDelete(null) })
    }
  }

  const editingCharacter =
    modalState.editingId != null
      ? activeCharacters.find((c) => c.id === modalState.editingId)
      : undefined
  const pictureUri = pendingPicture?.uri ?? (editingCharacter?.picture || null)

  // Same "the picture worker only ever surfaces here" pattern as
  // StagesContainer - see its identical effect for the full doc.
  const previouslyPendingIds = useRef<Set<string>>(new Set())
  useEffect(() => {
    const failed = activeCharacters.filter(
      (c) => c.pictureStatus === 'FAILED' && previouslyPendingIds.current.has(c.id)
    )
    failed.forEach(() => toast({ title: t('toasts.characterPictureFailed'), preset: 'error' }))
    previouslyPendingIds.current = new Set(
      activeCharacters.filter((c) => c.pictureStatus === 'PENDING').map((c) => c.id)
    )
  }, [activeCharacters, t])

  const activeFormErrors = (modalState.kind === 'JOJO' ? jojoForm : onePieceForm).formState
    .errors as {
    translations?: Partial<Record<Locale, unknown>>
  }
  const erroredLocales = SUPPORTED_LOCALES.filter((locale) =>
    Boolean(activeFormErrors.translations?.[locale])
  )

  return (
    <CharactersScreen
      characters={activeCharacters}
      rows={rowsForCharacter}
      isLoading={isLoading}
      isError={isError}
      onRetry={onRetry}
      onCreateNew={openCreate}
      onEdit={(character) => void openEdit(character)}
      onDelete={(character) =>
        setCharacterToDelete({ kind: character.kind, id: character.id, name: character.name })
      }
      openingEditId={openingEditId}
      search={search}
      onSearchChange={setSearch}
      mangaFilter={mangaFilter}
      onMangaFilterChange={setMangaFilter}
      hasActiveFilters={hasActiveFilters}
      detailCharacter={detailCharacter}
      onOpenDetail={setDetailCharacter}
      onCloseDetail={() => setDetailCharacter(null)}
      form={{
        visible: modalState.visible,
        mode: modalState.mode,
        kind: modalState.kind,
        onSelectKind,
        jojoControl: jojoForm.control,
        jojoErrors: jojoForm.formState.errors,
        onePieceControl: onePieceForm.control,
        onePieceErrors: onePieceForm.formState.errors,
        onSubmit,
        onCancel: closeModal,
        isSaving:
          createJojo.isPending ||
          updateJojo.isPending ||
          createOnePiece.isPending ||
          updateOnePiece.isPending,
        pictureUri,
        onPickPicture: () => void onPickPicture(),
        isPictureBusy: uploadJojoPicture.isPending || uploadOnePiecePicture.isPending,
        activeLocale,
        onLocaleChange: setActiveLocale,
        erroredLocales,
      }}
      deleteConfirm={{
        visible: characterToDelete !== null,
        isConfirming: deleteJojo.isPending || deleteOnePiece.isPending,
        onConfirm: onConfirmDelete,
        onCancel: () => setCharacterToDelete(null),
        characterName: characterToDelete?.name,
      }}
    />
  )
}
