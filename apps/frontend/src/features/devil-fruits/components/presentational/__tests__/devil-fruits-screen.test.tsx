import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
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

type HarnessProps = {
  readOnly?: boolean
  devilFruits?: DevilFruitResponse[]
  search?: string
  onSearchChange?: (search: string) => void
  hasActiveFilters?: boolean
}

// Real (not mocked) detail-modal state - see stands-screen.test.tsx's
// Harness for why.
function Harness({ readOnly, devilFruits, search, onSearchChange, hasActiveFilters }: HarnessProps) {
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
  const [detailFruit, setDetailFruit] = useState<DevilFruitResponse | null>(null)

  const base = {
    devilFruits: devilFruits ?? [baseFruit()],
    isLoading: false,
    isError: false,
    onRetry: jest.fn(),
    search: search ?? '',
    onSearchChange: onSearchChange ?? jest.fn(),
    rarityFilter: null,
    rarityFilterOptions: [{ value: 'LEGENDARY', label: 'Legendary' }],
    onRarityFilterChange: jest.fn(),
    fruitTypeFilter: null,
    fruitTypeFilterOptions: [{ value: 'MYTHICAL_ZOAN', label: 'Mythical Zoan' }],
    onFruitTypeFilterChange: jest.fn(),
    hasActiveFilters: hasActiveFilters ?? false,
    detailFruit,
    onOpenDetail: setDetailFruit,
    onCloseDetail: () => setDetailFruit(null),
  }

  if (readOnly) return <DevilFruitsScreen {...base} readOnly />

  return (
    <DevilFruitsScreen
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

  it('opens the detail modal with the fruit\'s description on card press', async () => {
    await renderWithProviders(<Harness />)

    expect(screen.queryByText('Turns the user into rubber')).toBeNull()

    await act(async () => {
      fireEvent.press(screen.getByLabelText('View Gomu Gomu no Mi details'))
    })

    expect(screen.getByText('Turns the user into rubber')).toBeTruthy()
  })

  describe('readOnly', () => {
    it('hides "New Devil Fruit" and every card\'s edit/delete buttons', async () => {
      await renderWithProviders(<Harness readOnly />)

      expect(screen.queryByLabelText('New Devil Fruit')).toBeNull()
      expect(screen.queryByLabelText('Edit Gomu Gomu no Mi')).toBeNull()
      expect(screen.queryByLabelText('Delete Gomu Gomu no Mi')).toBeNull()
    })

    it('still opens the detail modal on card press', async () => {
      await renderWithProviders(<Harness readOnly />)

      await act(async () => {
        fireEvent.press(screen.getByLabelText('View Gomu Gomu no Mi details'))
      })

      expect(screen.getByText('Turns the user into rubber')).toBeTruthy()
      expect(screen.queryByLabelText('Edit Gomu Gomu no Mi')).toBeNull()
    })
  })
})
