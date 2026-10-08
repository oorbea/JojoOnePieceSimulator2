import { act, fireEvent, renderWithProviders, screen, waitFor } from '@/test/render'
import { getStands } from '@/features/stands/api/stands.api'

import { StandsContainer } from '../stands-container'

jest.mock('@/features/stands/api/stands.api', () => ({
  getStands: jest.fn().mockResolvedValue([]),
  getStandTranslations: jest.fn(),
  getStandOptions: jest.fn().mockResolvedValue([]),
}))
jest.mock('@/shared/hooks/use-picture-picker', () => ({
  usePicturePicker: () => ({ pickPicture: jest.fn() }),
}))

const mockedGetStands = getStands as jest.MockedFunction<typeof getStands>

describe('StandsContainer - no-picture filter wiring', () => {
  beforeEach(() => mockedGetStands.mockClear())

  it('sends hasPicture=false through the filters object only while the chip is on', async () => {
    await renderWithProviders(<StandsContainer />)
    await waitFor(() => expect(mockedGetStands).toHaveBeenCalled())

    await act(async () => {
      fireEvent.press(screen.getByLabelText('No picture'))
    })
    await waitFor(() =>
      expect(mockedGetStands).toHaveBeenLastCalledWith({ hasPicture: false })
    )

    await act(async () => {
      fireEvent.press(screen.getByLabelText('No picture'))
    })
    await waitFor(() => expect(mockedGetStands.mock.calls.at(-1)?.[0]).toBeUndefined())
  })
})
