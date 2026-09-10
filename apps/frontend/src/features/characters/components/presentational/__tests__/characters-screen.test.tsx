import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyCharacterTranslationsForm } from '@/shared/lib/character-translations'
import {
  jojoCharacterFormSchema,
  type JojoCharacterFormValues,
  type JojoCharacterResponse,
} from '@/features/characters/types/characters.types'
import { JOJO_STAT_ROWS } from '@/features/characters/lib/character-stats'

import { CharactersScreen } from '../characters-screen'

function baseCharacter(overrides: Partial<JojoCharacterResponse> = {}): JojoCharacterResponse {
  return {
    id: 'char-1',
    name: 'Jotaro Kujo',
    description: 'A delinquent with a powerful Stand.',
    rarity: 'LEGENDARY',
    picture: '',
    pictureThumb: '',
    pictureCard: '',
    pictureStatus: 'NONE',
    pictureLqip: '',
    hamon: 'NONE',
    spin: 'NONE',
    battleIq: 130,
    ...overrides,
  }
}

type HarnessProps = {
  readOnly?: boolean
  characters?: JojoCharacterResponse[]
  search?: string
  onSearchChange?: (search: string) => void
  mangaFilter?: 'JOJO' | 'ONE_PIECE'
  onMangaFilterChange?: (manga: 'JOJO' | 'ONE_PIECE') => void
  hasActiveFilters?: boolean
  hasNextPage?: boolean
  isFetchingNextPage?: boolean
  isLoadMoreError?: boolean
  onLoadMore?: () => void
  total?: number
}

// Real (not mocked) detail-modal state - see devil-fruits-screen.test.tsx's
// Harness for why.
function Harness({
  readOnly,
  characters,
  search,
  onSearchChange,
  mangaFilter,
  onMangaFilterChange,
  hasActiveFilters,
  hasNextPage,
  isFetchingNextPage,
  isLoadMoreError,
  onLoadMore,
  total,
}: HarnessProps) {
  const { control, formState } = useForm<JojoCharacterFormValues>({
    resolver: zodResolver(jojoCharacterFormSchema),
    defaultValues: {
      name: '',
      rarity: 'COMMON',
      hamon: 'NONE',
      spin: 'NONE',
      battleIq: 100,
      translations: createEmptyCharacterTranslationsForm(),
    },
  })
  const [detailCharacter, setDetailCharacter] = useState<JojoCharacterResponse | null>(null)

  const base = {
    characters: characters ?? [baseCharacter()],
    rows: JOJO_STAT_ROWS,
    isLoading: false,
    isError: false,
    onRetry: jest.fn(),
    search: search ?? '',
    onSearchChange: onSearchChange ?? jest.fn(),
    mangaFilter: mangaFilter ?? 'JOJO',
    onMangaFilterChange: onMangaFilterChange ?? jest.fn(),
    hasActiveFilters: hasActiveFilters ?? false,
    detailCharacter,
    onOpenDetail: setDetailCharacter,
    onCloseDetail: () => setDetailCharacter(null),
    hasNextPage,
    isFetchingNextPage,
    isLoadMoreError,
    onLoadMore,
    total,
  }

  if (readOnly) return <CharactersScreen {...base} readOnly />

  return (
    <CharactersScreen
      {...base}
      onCreateNew={jest.fn()}
      onEdit={jest.fn()}
      onDelete={jest.fn()}
      openingEditId={null}
      form={{
        visible: false,
        mode: 'create',
        kind: null,
        onSelectKind: jest.fn(),
        jojoControl: control,
        jojoErrors: formState.errors,
        onePieceControl: control as never,
        onePieceErrors: {},
        onSubmit: jest.fn(),
        onCancel: jest.fn(),
        isSaving: false,
        pictureUri: null,
        onPickPicture: jest.fn(),
        isPictureBusy: false,
        activeLocale: 'en-GB',
        onLocaleChange: jest.fn(),
        erroredLocales: [],
      }}
      deleteConfirm={{
        visible: false,
        isConfirming: false,
        onConfirm: jest.fn(),
        onCancel: jest.fn(),
      }}
    />
  )
}

describe('CharactersScreen', () => {
  it('renders the search and manga filter', async () => {
    await renderWithProviders(<Harness />)

    expect(screen.getByLabelText('Search')).toBeTruthy()
    expect(screen.getByText('Filter by manga')).toBeTruthy()
  })

  it('calls onMangaFilterChange when the manga filter changes', async () => {
    const onMangaFilterChange = jest.fn()
    await renderWithProviders(<Harness onMangaFilterChange={onMangaFilterChange} />)

    // GlassSelect's trigger a11y label is "<field label>: <selected
    // option's label>" - see glass-select.tsx.
    await act(async () => {
      fireEvent.press(screen.getByLabelText("Filter by manga: JoJo's Bizarre Adventure"))
    })
    await act(async () => {
      fireEvent.press(screen.getByLabelText('One Piece'))
    })

    expect(onMangaFilterChange).toHaveBeenCalledWith('ONE_PIECE')
  })

  it('calls onSearchChange as the user types', async () => {
    const onSearchChange = jest.fn()
    await renderWithProviders(<Harness onSearchChange={onSearchChange} />)

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search'), 'jotaro')
    })

    expect(onSearchChange).toHaveBeenCalledWith('jotaro')
  })

  it('shows the plain empty state with no filters active', async () => {
    await renderWithProviders(<Harness characters={[]} hasActiveFilters={false} />)

    expect(screen.getByText('No Characters yet. Create the first one.')).toBeTruthy()
  })

  it('shows the filtered empty state once a filter is active', async () => {
    await renderWithProviders(<Harness characters={[]} hasActiveFilters />)

    expect(screen.getByText('No Characters match. Try a different search or filter.')).toBeTruthy()
  })

  it("opens the detail modal with the character's description on card press", async () => {
    await renderWithProviders(<Harness />)

    expect(screen.queryByText('A delinquent with a powerful Stand.')).toBeNull()

    await act(async () => {
      fireEvent.press(screen.getByLabelText('View Jotaro Kujo details'))
    })

    expect(screen.getByText('A delinquent with a powerful Stand.')).toBeTruthy()
  })

  describe('readOnly', () => {
    it('hides "New Character" and every card\'s edit/delete buttons', async () => {
      await renderWithProviders(<Harness readOnly />)

      expect(screen.queryByLabelText('New Character')).toBeNull()
      expect(screen.queryByLabelText('Edit Jotaro Kujo')).toBeNull()
      expect(screen.queryByLabelText('Delete Jotaro Kujo')).toBeNull()
    })

    it('still opens the detail modal on card press', async () => {
      await renderWithProviders(<Harness readOnly />)

      await act(async () => {
        fireEvent.press(screen.getByLabelText('View Jotaro Kujo details'))
      })

      expect(screen.getByText('A delinquent with a powerful Stand.')).toBeTruthy()
      expect(screen.queryByLabelText('Edit Jotaro Kujo')).toBeNull()
    })
  })

  describe('"Cargar más" (load more)', () => {
    it('renders no load-more section when onLoadMore is omitted', async () => {
      await renderWithProviders(<Harness hasNextPage total={40} />)

      expect(screen.queryByLabelText('Load more')).toBeNull()
    })

    it('shows the Load more button with the loaded/total count when there is a next page', async () => {
      const onLoadMore = jest.fn()
      await renderWithProviders(<Harness hasNextPage onLoadMore={onLoadMore} total={40} />)

      expect(screen.getByLabelText('Load more')).toBeTruthy()
      expect(screen.getByText('1 / 40')).toBeTruthy()

      await act(async () => {
        fireEvent.press(screen.getByLabelText('Load more'))
      })
      expect(onLoadMore).toHaveBeenCalled()
    })

    it('shows "That\'s all" once the list is exhausted', async () => {
      await renderWithProviders(<Harness hasNextPage={false} onLoadMore={jest.fn()} total={1} />)

      expect(screen.getByText("That's all (1)")).toBeTruthy()
      expect(screen.queryByLabelText('Load more')).toBeNull()
    })
  })
})
