import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'

import { fireEvent, renderWithProviders, screen } from '@/test/render'
import { standFormSchema, type StandFormValues } from '@/features/stands/types/stands.types'

import { StandFormModal } from '../stand-form-modal'

const DEFAULT_VALUES: StandFormValues = {
  name: '',
  description: '',
  rarity: 'COMMON',
  skills: [],
  attackPower: 'NULL',
  speed: 'NULL',
  attackRange: 'NULL',
  endurance: 'NULL',
  precision: 'NULL',
  potential: 'NULL',
  evolvesFromId: null,
}

function Harness({ onSubmit, onCancel }: { onSubmit: () => void; onCancel: () => void }) {
  const {
    control,
    formState: { errors },
  } = useForm<StandFormValues>({
    resolver: zodResolver(standFormSchema),
    defaultValues: DEFAULT_VALUES,
  })

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
    await renderWithProviders(<Harness onSubmit={onSubmit} onCancel={onCancel} />)

    fireEvent.press(screen.getByLabelText('Cancel'))
    expect(onCancel).toHaveBeenCalledTimes(1)

    fireEvent.press(screen.getByLabelText('Save Stand'))
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })
})
