import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { SkillsField } from '../skills-field'

function baseProps(overrides: Record<string, unknown> = {}) {
  return {
    label: 'Skills',
    skills: [] as string[],
    onAdd: jest.fn(),
    onRemove: jest.fn(),
    ...overrides,
  }
}

describe('SkillsField', () => {
  it('renders existing skills as chips', async () => {
    await renderWithProviders(<SkillsField {...baseProps({ skills: ['Time Stop', 'Punch'] })} />)

    expect(screen.getByText('Time Stop')).toBeTruthy()
    expect(screen.getByText('Punch')).toBeTruthy()
  })

  it('calls onAdd with the trimmed draft text', async () => {
    const onAdd = jest.fn()
    await renderWithProviders(<SkillsField {...baseProps({ onAdd })} />)

    // fireEvent.changeText's resulting setState needs an explicit act()
    // flush before the next query/press sees the updated `draft` — a bare
    // fireEvent.changeText leaves that update pending in this render setup.
    await act(async () => {
      fireEvent.changeText(screen.getByLabelText('Skills'), '  Barrage  ')
    })
    fireEvent.press(screen.getByLabelText('Add skill'))

    expect(onAdd).toHaveBeenCalledWith('Barrage')
  })

  it('does not call onAdd for an empty draft', async () => {
    const onAdd = jest.fn()
    await renderWithProviders(<SkillsField {...baseProps({ onAdd })} />)

    fireEvent.press(screen.getByLabelText('Add skill'))

    expect(onAdd).not.toHaveBeenCalled()
  })

  it('calls onRemove with the chip index', async () => {
    const onRemove = jest.fn()
    await renderWithProviders(<SkillsField {...baseProps({ skills: ['Time Stop', 'Punch'], onRemove })} />)

    fireEvent.press(screen.getByLabelText('Remove Punch'))

    expect(onRemove).toHaveBeenCalledWith(1)
  })

  it('shows the error message when provided', async () => {
    await renderWithProviders(<SkillsField {...baseProps({ error: 'At least one skill is required' })} />)

    expect(screen.getByText('At least one skill is required')).toBeTruthy()
  })
})
