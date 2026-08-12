import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyTranslationsForm } from '@/shared/lib/power-translations'
import {
  devilFruitFormSchema,
  type DevilFruitFormValues,
  type DevilFruitResponse,
} from '@/features/devil-fruits/types/devil-fruits.types'

import { DevilFruitsScreen } from '../devil-fruits-screen'

function baseFruit(overrides: Partial<DevilFruitResponse> = {}): DevilFruitResponse {
  return {
    id: 'fruit-1',
    name: 'Gomu Gomu no Mi',
    description: 'Turns the user into rubber',
    rarity: 'LEGENDARY',
    skills: ['Gum-Gum Pistol'],
    picture: '',
    pictureThumb: '',
    pictureStatus: 'NONE',
    fruitType: 'MYTHICAL_ZOAN',
    ...overrides,
  }
}

function Harness(
  props: Partial<React.ComponentProps<typeof DevilFruitsScreen>> & {
    devilFruits?: DevilFruitResponse[]
  }
) {
  const {
    control,
    formState: { errors },
  } = useForm<DevilFruitFormValues>({
    resolver: zodResolver(devilFruitFormSchema),
    defaultValues: {
      name: '',
      translations: createEmptyTranslationsForm(),
      rarity: 'COMMON',
      fruitType: 'PARAMECIA',
    },
  })

  return (
    <DevilFruitsScreen
      devilFruits={[baseFruit()]}
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
      fruitTypeFilter={null}
      fruitTypeFilterOptions={[{ value: 'MYTHICAL_ZOAN', label: 'Mythical Zoan' }]}
      onFruitTypeFilterChange={jest.fn()}
      hasActiveFilters={false}
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
      {...props}
    />
  )
}

describe('DevilFruitsScreen', () => {
  it('renders the search, rarity, and fruit-type filters', async () => {
    await renderWithProviders(<Harness />)

    expect(screen.getByLabelText('Search')).toBeTruthy()
    expect(screen.getByText('Filter by rarity')).toBeTruthy()
    expect(screen.getByText('Filter by fruit type')).toBeTruthy()
  })

  it('calls onSearchChange as the user types', async () => {
    const onSearchChange = jest.fn()
    await renderWithProviders(<Harness onSearchChange={onSearchChange} />)

    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Search'), 'gomu')
    })

    expect(onSearchChange).toHaveBeenCalledWith('gomu')
  })

  it('shows the plain empty state with no filters active', async () => {
    await renderWithProviders(<Harness devilFruits={[]} hasActiveFilters={false} />)

    expect(screen.getByText('No Devil Fruits yet. Create the first one.')).toBeTruthy()
  })

  it('shows the filtered empty state once a filter is active', async () => {
    await renderWithProviders(<Harness devilFruits={[]} hasActiveFilters />)

    expect(
      screen.getByText('No Devil Fruits match. Try a different search or filter.')
    ).toBeTruthy()
  })
})
