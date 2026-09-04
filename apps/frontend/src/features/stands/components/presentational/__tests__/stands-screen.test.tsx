import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyTranslationsForm } from '@/shared/lib/power-translations'
import {
  standFormSchema,
  type StandFormValues,
  type StandResponse,
} from '@/features/stands/types/stands.types'

import { StandsScreen } from '../stands-screen'

function baseStand(overrides: Partial<StandResponse> = {}): StandResponse {
  return {
    id: 'stand-1',
    name: 'Star Platinum',
    description: 'A close-range powerhouse',
    rarity: 'LEGENDARY',
    skills: ['Time Stop'],
    picture: '',
    pictureThumb: '',
    pictureStatus: 'NONE',
    attackPower: 'A',
    speed: 'A',
    attackRange: 'E',
    endurance: 'C',
    precision: 'A',
    potential: 'A',
    evolvesFrom: null,
    ...overrides,
  }
}

type HarnessProps = {
  readOnly?: boolean
  stands?: StandResponse[]
  search?: string
  onSearchChange?: (search: string) => void
  filtersExpanded?: boolean
  hasActiveFilters?: boolean
}

// Real (not mocked) detail-modal state, so pressing a card's "View details"
// affordance exercises the same open/close flow a real container would -
// see stands-screen's own detailStand/onOpenDetail/onCloseDetail props,
// which are identical in shape for both the readOnly and writable variants.
function Harness({ readOnly, stands, search, onSearchChange, filtersExpanded, hasActiveFilters }: HarnessProps) {
  const {
    control,
    formState: { errors },
  } = useForm<StandFormValues>({
    resolver: zodResolver(standFormSchema),
    defaultValues: {
      name: '',
      translations: createEmptyTranslationsForm(),
      rarity: 'COMMON',
      attackPower: 'NULL',
      speed: 'NULL',
      attackRange: 'NULL',
      endurance: 'NULL',
      precision: 'NULL',
      potential: 'NULL',
      evolvesFromId: null,
    },
  })
  const [detailStand, setDetailStand] = useState<StandResponse | null>(null)

  const base = {
    stands: stands ?? [baseStand()],
    isLoading: false,
    isError: false,
    onRetry: jest.fn(),
    search: search ?? '',
    onSearchChange: onSearchChange ?? jest.fn(),
    rarityFilter: null,
    rarityFilterOptions: [{ value: 'LEGENDARY', label: 'Legendary' }],
    onRarityFilterChange: jest.fn(),
    statFilters: {
      attackPower: null,
      speed: null,
      attackRange: null,
      endurance: null,
      precision: null,
      potential: null,
    },
    statFilterOptions: [{ value: 'A', label: 'A' }],
    onStatFilterChange: jest.fn(),
    evolvesFromFilter: null,
    evolvesFromFilterOptions: [],
    onEvolvesFromFilterChange: jest.fn(),
    filtersExpanded: filtersExpanded ?? false,
    onToggleFilters: jest.fn(),
    moreFiltersCount: 0,
    onClearFilters: jest.fn(),
    hasActiveFilters: hasActiveFilters ?? false,
    detailStand,
    onOpenDetail: setDetailStand,
    onCloseDetail: () => setDetailStand(null),
  }

  if (readOnly) return <StandsScreen {...base} readOnly />

  return (
    <StandsScreen
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
        evolvesFromOptions: [],
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

describe('StandsScreen', () => {
  it('renders the search field and rarity filter always visible', async () => {
    await renderWithProviders(<Harness />)

    expect(screen.getByLabelText('Search')).toBeTruthy()
    expect(screen.getByText('Filter by rarity')).toBeTruthy()
  })

  it('calls onSearchChange as the user types', async () => {
    const onSearchChange = jest.fn()
    await renderWithProviders(<Harness onSearchChange={onSearchChange} />)

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search'), 'chariot')
    })

    expect(onSearchChange).toHaveBeenCalledWith('chariot')
  })

  it('does not render the six stat filters until "More filters" is expanded', async () => {
    await renderWithProviders(<Harness filtersExpanded={false} />)

    expect(screen.queryByText('Attack Power')).toBeNull()

    await renderWithProviders(<Harness filtersExpanded />)
    expect(screen.getByText('Attack Power')).toBeTruthy()
    expect(screen.getByText('Filter by evolves from')).toBeTruthy()
  })

  it('shows the plain empty state with no filters active', async () => {
    await renderWithProviders(<Harness stands={[]} hasActiveFilters={false} />)

    expect(screen.getByText('No Stands yet. Create the first one.')).toBeTruthy()
  })

  it('shows the filtered empty state once a filter is active', async () => {
    await renderWithProviders(<Harness stands={[]} hasActiveFilters />)

    expect(screen.getByText('No Stands match. Try a different search or filter.')).toBeTruthy()
  })

  it('opens the detail modal with the stand\'s description on card press', async () => {
    await renderWithProviders(<Harness />)

    expect(screen.queryByText('A close-range powerhouse')).toBeNull()

    await act(async () => {
      fireEvent.press(screen.getByLabelText('View Star Platinum details'))
    })

    expect(screen.getByText('A close-range powerhouse')).toBeTruthy()
  })

  describe('readOnly', () => {
    it('hides "New Stand" and every card\'s edit/delete buttons', async () => {
      await renderWithProviders(<Harness readOnly />)

      expect(screen.queryByLabelText('New Stand')).toBeNull()
      expect(screen.queryByLabelText('Edit Star Platinum')).toBeNull()
      expect(screen.queryByLabelText('Delete Star Platinum')).toBeNull()
    })

    it('still opens the detail modal on card press', async () => {
      await renderWithProviders(<Harness readOnly />)

      await act(async () => {
        fireEvent.press(screen.getByLabelText('View Star Platinum details'))
      })

      expect(screen.getByText('A close-range powerhouse')).toBeTruthy()
      // Read-only detail modal has no Edit footer button.
      expect(screen.queryByLabelText('Edit Star Platinum')).toBeNull()
    })
  })
})
