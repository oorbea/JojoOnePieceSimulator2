import { fireEvent, renderWithProviders, screen } from '@/test/render'
import type { StageResponse } from '@/features/stages/types/stages.types'

import { StageCard } from '../stage-card'

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

describe('StageCard', () => {
  it('renders the name, manga and order badges', async () => {
    await renderWithProviders(
      <StageCard stage={baseStage()} onEdit={jest.fn()} onDelete={jest.fn()} />
    )

    expect(screen.getByText('Stardust Crusaders')).toBeTruthy()
    expect(screen.getByText("JoJo's Bizarre Adventure")).toBeTruthy()
    expect(screen.getByText('#3')).toBeTruthy()
  })

  it('renders the description', async () => {
    await renderWithProviders(
      <StageCard stage={baseStage()} onEdit={jest.fn()} onDelete={jest.fn()} />
    )

    expect(screen.getByText('A globe-trotting journey to save Holly Kujo.')).toBeTruthy()
  })

  it('fires onEdit and onDelete from their own buttons', async () => {
    const onEdit = jest.fn()
    const onDelete = jest.fn()
    await renderWithProviders(<StageCard stage={baseStage()} onEdit={onEdit} onDelete={onDelete} />)

    fireEvent.press(screen.getByLabelText('Edit Stardust Crusaders'))
    expect(onEdit).toHaveBeenCalledTimes(1)

    fireEvent.press(screen.getByLabelText('Delete Stardust Crusaders'))
    expect(onDelete).toHaveBeenCalledTimes(1)
  })
})
