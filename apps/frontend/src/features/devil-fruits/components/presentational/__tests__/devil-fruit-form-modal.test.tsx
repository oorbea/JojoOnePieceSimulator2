import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyTranslationsForm } from '@/shared/lib/power-translations'
import type { Locale } from '@/shared/contracts/enums'
import { devilFruitFormSchema, type DevilFruitFormValues } from '@/features/devil-fruits/types/devil-fruits.types'

import { DevilFruitFormModal } from '../devil-fruit-form-modal'

// A function, not a constant - see stand-form-modal.test.tsx's
// createDefaultValues for why.
function createDefaultValues(): DevilFruitFormValues {
  return {
    name: '',
    translations: createEmptyTranslationsForm(),
    rarity: 'COMMON',
    fruitType: 'PARAMECIA',
  }
}

function Harness({ onSubmit, onCancel }: { onSubmit: () => void; onCancel: () => void }) {
  const {
    control,
    formState: { errors },
  } = useForm<DevilFruitFormValues>({
    resolver: zodResolver(devilFruitFormSchema),
    defaultValues: createDefaultValues(),
  })
  const [activeLocale, setActiveLocale] = useState<Locale>('en-GB')

  return (
    <DevilFruitFormModal
      visible
      mode="create"
      control={control}
      errors={errors}
      onSubmit={onSubmit}
      onCancel={onCancel}
      isSaving={false}
      pictureUri={null}
      onPickPicture={jest.fn()}
      isPictureBusy={false}
      activeLocale={activeLocale}
      onLocaleChange={setActiveLocale}
      erroredLocales={[]}
    />
  )
}

describe('DevilFruitFormModal', () => {
  it('renders the title and every field in create mode', async () => {
    await renderWithProviders(<Harness onSubmit={jest.fn()} onCancel={jest.fn()} />)

    expect(screen.getByText('New Devil Fruit')).toBeTruthy()
    expect(screen.getByLabelText('Name')).toBeTruthy()
    expect(screen.getByLabelText('Description')).toBeTruthy()
    expect(screen.getByText('Fruit Type')).toBeTruthy()
  })

  it('fires onCancel and onSubmit from their own buttons', async () => {
    const onSubmit = jest.fn()
    const onCancel = jest.fn()
    const view = await renderWithProviders(<Harness onSubmit={onSubmit} onCancel={onCancel} />)

    // See stand-form-modal.test.tsx's equivalent test for why each press
    // gets its own awaited act().
    await act(async () => {
      fireEvent.press(view.getByLabelText('Cancel'))
    })
    expect(onCancel).toHaveBeenCalledTimes(1)

    await act(async () => {
      fireEvent.press(view.getByLabelText('Save Devil Fruit'))
    })
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })

  it('switches locale tabs to an independent, empty Description field', async () => {
    const view = await renderWithProviders(<Harness onSubmit={jest.fn()} onCancel={jest.fn()} />)

    await act(async () => {
      fireEvent.changeText(view.getByLabelText('Description'), 'English description')
    })
    await act(async () => {
      fireEvent.press(view.getByLabelText(/Español \(España\)/))
    })

    expect(view.getByLabelText('Description').props.value).toBe('')
  })
})
