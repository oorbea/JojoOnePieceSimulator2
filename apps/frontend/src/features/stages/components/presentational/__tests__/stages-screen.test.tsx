import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyStageTranslationsForm } from '@/shared/lib/stage-translations'
import {
  stageFormSchema,
  type StageFormValues,
  type StageResponse,
} from '@/features/stages/types/stages.types'

import { StagesScreen } from '../stages-screen'

function baseStage(overrides: Partial<StageResponse> = {}): StageResponse {
  return {
    id: 'stage-1',
    manga: 'JOJO',
    order: 3,
    name: 'Stardust Crusaders',
    description: 'A globe-trotting journey to save Holly Kujo.',
    picture: '',
    pictureThumb: '',
    pictureStatus: 'NONE',
    ...overrides,
  }
}

type HarnessProps = {
  readOnly?: boolean
  stages?: StageResponse[]
  search?: string
  onSearchChange?: (search: string) => void
  hasActiveFilters?: boolean
}

// Real (not mocked) detail-modal state - see stands-screen.test.tsx's
// Harness for why.
function Harness({ readOnly, stages, search, onSearchChange, hasActiveFilters }: HarnessProps) {
  const {
    control,
    formState: { errors },
  } = useForm<StageFormValues>({
    resolver: zodResolver(stageFormSchema),
    defaultValues: {
      manga: 'JOJO',
      order: 0,
      name: '',
      translations: createEmptyStageTranslationsForm(),
    },
  })
  const [detailStage, setDetailStage] = useState<StageResponse | null>(null)

  const base = {
    stages: stages ?? [baseStage()],
    isLoading: false,
    isError: false,
    onRetry: jest.fn(),
    search: search ?? '',
    onSearchChange: onSearchChange ?? jest.fn(),
    mangaFilter: null,
    mangaFilterOptions: [{ value: 'JOJO', label: "JoJo's Bizarre Adventure" }],
    onMangaFilterChange: jest.fn(),
    hasActiveFilters: hasActiveFilters ?? false,
    detailStage,
    onOpenDetail: setDetailStage,
    onCloseDetail: () => setDetailStage(null),
  }

  if (readOnly) return <StagesScreen {...base} readOnly />

  return (
    <StagesScreen
      {...base}
      onCreateNew={jest.fn()}
      onEdit={jest.fn()}
      onDelete={jest.fn()}
      openingEditId={null}
      form={{
        visible: false,
        mode: 'create',
        control,
        errors,
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

describe('StagesScreen', () => {
  it('renders the search field and manga filter', async () => {
    await renderWithProviders(<Harness />)

    expect(screen.getByLabelText('Search')).toBeTruthy()
    expect(screen.getByText('Filter by manga')).toBeTruthy()
  })

  it('calls onSearchChange as the user types', async () => {
    const onSearchChange = jest.fn()
    await renderWithProviders(<Harness onSearchChange={onSearchChange} />)

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search'), 'stardust')
    })

    expect(onSearchChange).toHaveBeenCalledWith('stardust')
  })

  it('shows the plain empty state with no filters active', async () => {
    await renderWithProviders(<Harness stages={[]} hasActiveFilters={false} />)

    expect(screen.getByText('No Stages yet. Create the first one.')).toBeTruthy()
  })

  it('shows the filtered empty state once a filter is active', async () => {
    await renderWithProviders(<Harness stages={[]} hasActiveFilters />)

    expect(screen.getByText('No Stages match. Try a different search or filter.')).toBeTruthy()
  })

  it('opens the detail modal with the stage\'s description on card press', async () => {
    await renderWithProviders(<Harness />)

    // StageCard already shows the (visually truncated) description, so a
    // second occurrence appearing after the press is what proves the
    // detail modal rendered - a plain getByText would ambiguously match
    // both the card's and the modal's copy.
    expect(screen.getAllByText('A globe-trotting journey to save Holly Kujo.')).toHaveLength(1)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('View Stardust Crusaders details'))
    })

    expect(screen.getAllByText('A globe-trotting journey to save Holly Kujo.')).toHaveLength(2)
  })

  describe('readOnly', () => {
    it('hides "New Stage" and every card\'s edit/delete buttons', async () => {
      await renderWithProviders(<Harness readOnly />)

      expect(screen.queryByLabelText('New Stage')).toBeNull()
      expect(screen.queryByLabelText('Edit Stardust Crusaders')).toBeNull()
      expect(screen.queryByLabelText('Delete Stardust Crusaders')).toBeNull()
    })

    it('still opens the detail modal on card press', async () => {
      await renderWithProviders(<Harness readOnly />)

      await act(async () => {
        fireEvent.press(screen.getByLabelText('View Stardust Crusaders details'))
      })

      expect(screen.getAllByText('A globe-trotting journey to save Holly Kujo.')).toHaveLength(2)
      expect(screen.queryByLabelText('Edit Stardust Crusaders')).toBeNull()
    })
  })
})
