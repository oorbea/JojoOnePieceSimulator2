import { act, fireEvent, renderWithProviders } from '@/test/render'

import { FocalPointModal } from '../focal-point-modal'

describe('FocalPointModal', () => {
  it('confirms the current draft, not the value it was opened with', async () => {
    const onConfirm = jest.fn()
    const { getByRole } = await renderWithProviders(
      <FocalPointModal visible uri="file:///pic.jpg" x={0.1} y={0.9} onConfirm={onConfirm} />
    )

    // The picker's onChange is wired to the modal's own local draft state,
    // not straight through to onConfirm - editing inside the picker (here,
    // via its "reset to centre" button) must show up in what gets
    // confirmed, not the original x/y props the modal was opened with.
    // Each press gets its own awaited act() - see stand-card.test.tsx.
    await act(async () => {
      fireEvent.press(getByRole('button', { name: /reset/i }))
    })
    await act(async () => {
      fireEvent.press(getByRole('button', { name: /confirm/i }))
    })

    expect(onConfirm).toHaveBeenCalledWith(0.5, 0.5)
  })

  it('has no Cancel button in the mandatory (post-upload) mode', async () => {
    const { queryByRole } = await renderWithProviders(
      <FocalPointModal visible uri="file:///pic.jpg" x={0.5} y={0.5} onConfirm={() => {}} />
    )

    expect(queryByRole('button', { name: /cancel/i })).toBeNull()
  })

  it('shows Cancel and calls it, not onConfirm, when reopened from the edit form', async () => {
    const onConfirm = jest.fn()
    const onCancel = jest.fn()
    const { getByRole } = await renderWithProviders(
      <FocalPointModal
        visible
        uri="file:///pic.jpg"
        x={0.5}
        y={0.5}
        onConfirm={onConfirm}
        onCancel={onCancel}
      />
    )

    fireEvent.press(getByRole('button', { name: /cancel/i }))

    expect(onCancel).toHaveBeenCalledTimes(1)
    expect(onConfirm).not.toHaveBeenCalled()
  })
})
