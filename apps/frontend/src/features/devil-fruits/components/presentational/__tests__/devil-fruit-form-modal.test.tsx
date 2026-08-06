import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'

import { fireEvent, renderWithProviders, screen } from '@/test/render'
import { devilFruitFormSchema, type DevilFruitFormValues } from '@/features/devil-fruits/types/devil-fruits.types'

import { DevilFruitFormModal } from '../devil-fruit-form-modal'

const DEFAULT_VALUES: DevilFruitFormValues = {
  name: '',
  description: '',
  rarity: 'COMMON',
  skills: [],
  fruitType: 'PARAMECIA',
}

function Harness({ onSubmit, onCancel }: { onSubmit: () => void; onCancel: () => void }) {
  const {
    control,
    formState: { errors },
  } = useForm<DevilFruitFormValues>({
    resolver: zodResolver(devilFruitFormSchema),
    defaultValues: DEFAULT_VALUES,
  })

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
    await renderWithProviders(<Harness onSubmit={onSubmit} onCancel={onCancel} />)

    fireEvent.press(screen.getByLabelText('Cancel'))
    expect(onCancel).toHaveBeenCalledTimes(1)

    fireEvent.press(screen.getByLabelText('Save Devil Fruit'))
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })
})
