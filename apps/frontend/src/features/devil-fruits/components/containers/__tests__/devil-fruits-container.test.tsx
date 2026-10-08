import { act, fireEvent, renderWithProviders, screen, waitFor } from '@/test/render'
import { getDevilFruits } from '@/features/devil-fruits/api/devil-fruits.api'

import { DevilFruitsContainer } from '../devil-fruits-container'

jest.mock('@/features/devil-fruits/api/devil-fruits.api', () => ({
  getDevilFruits: jest.fn().mockResolvedValue([]),
  getDevilFruitTranslations: jest.fn(),
}))
jest.mock('@/shared/hooks/use-picture-picker', () => ({
  usePicturePicker: () => ({ pickPicture: jest.fn() }),
}))

const mockedGetDevilFruits = getDevilFruits as jest.MockedFunction<typeof getDevilFruits>

describe('DevilFruitsContainer - no-picture filter wiring', () => {
  beforeEach(() => mockedGetDevilFruits.mockClear())

  it('sends hasPicture=false through the filters object only while the chip is on', async () => {
    await renderWithProviders(<DevilFruitsContainer />)
    await waitFor(() => expect(mockedGetDevilFruits).toHaveBeenCalled())
    // No filter active yet: the unfiltered list is requested without params.
    expect(mockedGetDevilFruits).toHaveBeenLastCalledWith(undefined)

    await act(async () => {
      fireEvent.press(screen.getByLabelText('No picture'))
    })
    await waitFor(() => expect(mockedGetDevilFruits).toHaveBeenLastCalledWith({ hasPicture: false }))

    await act(async () => {
      fireEvent.press(screen.getByLabelText('No picture'))
    })
    await waitFor(() => expect(mockedGetDevilFruits).toHaveBeenLastCalledWith(undefined))
  })
})
