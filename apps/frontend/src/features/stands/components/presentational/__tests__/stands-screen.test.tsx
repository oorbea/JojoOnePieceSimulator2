import { zodResolver } from '@hookform/resolvers/zod'
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

function Harness(
  props: Partial<React.ComponentProps<typeof StandsScreen>> & { stands?: StandResponse[] }
) {
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

  return (
    <StandsScreen
      stands={[baseStand()]}
      isLoading={false}
      isError={false}
      onRetry={jest.fn()}
      onCreateNew={jest.fn()}
      onEdit={jest.fn()}
      onDelete={jest.fn()}
      openingEditId={null}
      search=""
      onSearchChange={jest.fn()}
      rarityFilter={null}
      rarityFilterOptions={[{ value: 'LEGENDARY', label: 'Legendary' }]}
      onRarityFilterChange={jest.fn()}
      statFilters={{
        attackPower: null,
        speed: null,
        attackRange: null,
        endurance: null,
        precision: null,
        potential: null,
      }}
      statFilterOptions={[{ value: 'A', label: 'A' }]}
      onStatFilterChange={jest.fn()}
      evolvesFromFilter={null}
      evolvesFromFilterOptions={[]}
      onEvolvesFromFilterChange={jest.fn()}
      filtersExpanded={false}
      onToggleFilters={jest.fn()}
      moreFiltersCount={0}
      onClearFilters={jest.fn()}
      hasActiveFilters={false}
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
      {...props}
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
})
