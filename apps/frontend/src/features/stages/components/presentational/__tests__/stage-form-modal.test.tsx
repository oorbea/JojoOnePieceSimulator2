import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { act, fireEvent, renderWithProviders, screen } from '@/test/render'
import { createEmptyStageTranslationsForm } from '@/shared/lib/stage-translations'
import type { Locale } from '@/shared/lib/zod'
import { stageFormSchema, type StageFormValues } from '@/features/stages/types/stages.types'

import { StageFormModal } from '../stage-form-modal'

// A function, not a constant - same reasoning as stand-form-modal.test.tsx's
// createDefaultValues: createEmptyStageTranslationsForm() must return a
// fresh object per Harness instance.
function createDefaultValues(): StageFormValues {
  return {
    manga: 'JOJO',
    order: 0,
    name: '',
    translations: createEmptyStageTranslationsForm(),
  }
}

function Harness({ onSubmit, onCancel }: { onSubmit: () => void; onCancel: () => void }) {
  const {
    control,
    formState: { errors },
  } = useForm<StageFormValues>({
    resolver: zodResolver(stageFormSchema),
    defaultValues: createDefaultValues(),
  })
  const [activeLocale, setActiveLocale] = useState<Locale>('en-GB')

  return (
    <StageFormModal
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

describe('StageFormModal', () => {
  it('renders the title and every field in create mode', async () => {
    await renderWithProviders(<Harness onSubmit={jest.fn()} onCancel={jest.fn()} />)

    expect(screen.getByText('New Stage')).toBeTruthy()
    expect(screen.getByLabelText('Name')).toBeTruthy()
    expect(screen.getByLabelText('Description')).toBeTruthy()
    expect(screen.getByText('Manga')).toBeTruthy()
    expect(screen.getByText('Order')).toBeTruthy()
  })

  it('fires onCancel and onSubmit from their own buttons', async () => {
    const onSubmit = jest.fn()
    const onCancel = jest.fn()
    const view = await renderWithProviders(<Harness onSubmit={onSubmit} onCancel={onCancel} />)

    // Each press gets its own awaited act() - see stand-form-modal.test.tsx
    // for why (RN Animated's pressStyle spring resolves inside act
    // asynchronously).
    await act(async () => {
      fireEvent.press(view.getByLabelText('Cancel'))
    })
    expect(onCancel).toHaveBeenCalledTimes(1)

    await act(async () => {
      fireEvent.press(view.getByLabelText('Save Stage'))
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
