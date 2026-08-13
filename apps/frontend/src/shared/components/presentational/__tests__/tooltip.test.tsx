import { Pressable, Text } from 'react-native'
import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { TooltipBubble, useTooltipTrigger } from '../tooltip'

// Exercises the hook directly against a bare RN `Pressable` rather than
// `GlossButton`, so this test locks in `useTooltipTrigger`'s own on/off
// behavior without depending on how Tamagui's `Button` forwards long-press
// events under the hood. `triggerRef.current` is stubbed with a fake
// `.measure()` (react-test-renderer's host instances don't implement the
// real native measure bridge call), matching the shape `useTooltipTrigger`
// expects from any real RN host component.
function TestTrigger({ label }: { label?: string }) {
  const { visible, anchor, triggerRef, triggerProps } = useTooltipTrigger(label)
  triggerRef.current = { measure: (cb) => cb(0, 0, 50, 20, 100, 200) }
  return (
    <Pressable accessibilityLabel="trigger" {...triggerProps}>
      <Text>Press me</Text>
      <TooltipBubble visible={visible} label={label} anchor={anchor} />
    </Pressable>
  )
}

describe('useTooltipTrigger / TooltipBubble', () => {
  it('is hidden before any long-press', async () => {
    await renderWithProviders(<TestTrigger label="Explains the button" />)

    expect(screen.queryByText('Explains the button')).toBeNull()
  })

  it('shows the label after a long-press (native has no hover)', async () => {
    await renderWithProviders(<TestTrigger label="Explains the button" />)

    await act(async () => {
      fireEvent(screen.getByLabelText('trigger'), 'longPress')
    })

    expect(screen.getByText('Explains the button')).toBeTruthy()
  })

  it('never renders without a label', async () => {
    await renderWithProviders(<TestTrigger />)

    await act(async () => {
      fireEvent(screen.getByLabelText('trigger'), 'longPress')
    })

    expect(screen.queryByText('Explains the button')).toBeNull()
  })
})
