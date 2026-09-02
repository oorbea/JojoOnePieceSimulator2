import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyTranslationsForm } from '@/shared/lib/power-translations'
import type { Locale } from '@/shared/contracts/enums'
import { standFormSchema, type StandFormValues } from '@/features/stands/types/stands.types'

import { StandFormModal } from '../stand-form-modal'

// A function, not a constant - createEmptyTranslationsForm() must return a
// fresh object per Harness instance, or edits from one test's form would
// leak into the next test's initial render (see power-translations.ts).
function createDefaultValues(): StandFormValues {
  return {
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
  }
}

function Harness({ onSubmit, onCancel }: { onSubmit: () => void; onCancel: () => void }) {
  const {
    control,
    formState: { errors },
  } = useForm<StandFormValues>({
    resolver: zodResolver(standFormSchema),
    defaultValues: createDefaultValues(),
  })
  const [activeLocale, setActiveLocale] = useState<Locale>('en-GB')

  return (
    <StandFormModal
      visible
      mode="create"
      control={control}
      errors={errors}
      onSubmit={onSubmit}
      onCancel={onCancel}
      isSaving={false}
      evolvesFromOptions={[{ value: 'stand-1', label: 'Star Platinum' }]}
      pictureUri={null}
      onPickPicture={jest.fn()}
      isPictureBusy={false}
      activeLocale={activeLocale}
      onLocaleChange={setActiveLocale}
      erroredLocales={[]}
    />
  )
}

describe('StandFormModal', () => {
  it('renders the title and every field in create mode', async () => {
    await renderWithProviders(<Harness onSubmit={jest.fn()} onCancel={jest.fn()} />)

    expect(screen.getByText('New Stand')).toBeTruthy()
    expect(screen.getByLabelText('Name')).toBeTruthy()
    expect(screen.getByLabelText('Description')).toBeTruthy()
    expect(screen.getByText('Attack Power')).toBeTruthy()
    expect(screen.getByText('Evolves From')).toBeTruthy()
  })

  it('fires onCancel and onSubmit from their own buttons', async () => {
    const onSubmit = jest.fn()
    const onCancel = jest.fn()
    const view = await renderWithProviders(<Harness onSubmit={onSubmit} onCancel={onCancel} />)

    // Each press gets its own awaited act() - two bare fireEvent.press
    // calls back to back leave overlapping act scopes (RN Animated's
    // pressStyle spring resolves inside act asynchronously) that corrupt
    // React's act-nesting for whatever test renders next in this file.
    await act(async () => {
      fireEvent.press(view.getByLabelText('Cancel'))
    })
    expect(onCancel).toHaveBeenCalledTimes(1)

    await act(async () => {
      fireEvent.press(view.getByLabelText('Save Stand'))
    })
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })

  it('switches locale tabs to an independent, empty Description field', async () => {
    const view = await renderWithProviders(<Harness onSubmit={jest.fn()} onCancel={jest.fn()} />)

    await act(async () => {
      fireEvent.changeText(view.getByLabelText('Description'), 'English description')
    })
    await act(async () => {
      fireEvent.press(view.getByLabelText(/Català \(Catalunya\)/))
    })

    expect(view.getByLabelText('Description').props.value).toBe('')
  })
})
