import { useEffect } from 'react'
import { Pressable, Text } from 'react-native'
import { act, fireEvent, renderWithProviders, screen } from '@/test/render'

import { notifyScroll } from '@/shared/lib/scroll-bus'

import { TooltipBubble, useTooltipTrigger } from '../tooltip'

// Exercises the hook directly against a bare RN `Pressable` rather than
// `GlossButton`, so this test locks in `useTooltipTrigger`'s own on/off
// behavior without depending on how Tamagui's `Button` forwards long-press
// events under the hood. `triggerRef.current` is stubbed with a fake
// `.measure()` in an effect, not during render (react-hooks/refs flags a
// ref write in the render body itself) - react-test-renderer's host
// instances don't implement the real native measure bridge call, so this
// fills in the shape `useTooltipTrigger` expects from any real RN host
// component, and it's set by the time a later `fireEvent` interaction runs.
function TestTrigger({ label }: { label?: string }) {
  const { visible, anchor, triggerRef, triggerProps } = useTooltipTrigger(label)
  useEffect(() => {
    triggerRef.current = { measure: (cb) => cb(0, 0, 50, 20, 100, 200) }
  }, [triggerRef])
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

  it('hides on a scroll-bus notification instead of staying stuck on screen', async () => {
    await renderWithProviders(<TestTrigger label="Explains the button" />)

    await act(async () => {
      fireEvent(screen.getByLabelText('trigger'), 'longPress')
    })
    expect(screen.getByText('Explains the button')).toBeTruthy()

    await act(async () => {
      notifyScroll()
    })

    expect(screen.queryByText('Explains the button')).toBeNull()
  })

  it('never renders without a label', async () => {
    await renderWithProviders(<TestTrigger />)

    await act(async () => {
      fireEvent(screen.getByLabelText('trigger'), 'longPress')
    })

    expect(screen.queryByText('Explains the button')).toBeNull()
  })
})
