import { act, fireEvent, renderWithProviders } from '@/test/render'

import { FocalPointPicker } from '../focal-point-picker'

// The original bug: `panResponder.panHandlers` were spread onto a Tamagui
// `YStack`, which renders a plain `div` on web and never implements RN's
// responder system, so the handlers were spread but never actually invoked
// by anything. This locks in that the gesture host is a real RN `View` -
// the same element type the (working) lobby drag
// (`use-player-drag.ts`/`player-row.tsx`) spreads its own panHandlers onto -
// and that it actually carries the responder props RN needs to claim a
// touch/click.
describe('FocalPointPicker', () => {
  it('hosts the gesture on a real RN View carrying the responder props', async () => {
    const { getByTestId } = await renderWithProviders(
      <FocalPointPicker uri="file:///pic.jpg" x={0.5} y={0.5} onChange={() => {}} />
    )

    const well = getByTestId('focal-point-well')
    // The underlying host component is RN's own `View` (not react-native-web
    // aware Tamagui YStack, which would be `type: 'div'` on web and never
    // implements the responder system at all - the original bug).
    expect(well.type).toBe('View')
    expect(typeof well.props.onStartShouldSetResponder).toBe('function')
    expect(typeof well.props.onResponderGrant).toBe('function')
    expect(typeof well.props.onResponderMove).toBe('function')
    // Claims the gesture immediately on touch-down (tap-to-set) and keeps
    // it once granted (drag-to-adjust), rather than waiting for movement.
    expect(well.props.onStartShouldSetResponder()).toBe(true)
    expect(well.props.onMoveShouldSetResponder()).toBe(true)
    // A parent ScrollView (the form modal scrolls on mobile) must not be
    // able to steal the drag mid-gesture.
    expect(well.props.onResponderTerminationRequest()).toBe(false)
  })

  // react-native-web's ImageLoader wraps the raw DOM `load` event as
  // `nativeEvent` instead of RN's `{source:{width,height}}` shape, so
  // `nativeEvent.source` is undefined on web - destructuring it threw,
  // uncaught, from inside the `image.decode()` promise chain RNW calls
  // onLoad from (surfaced as "Uncaught (in promise)" in the browser).
  it('reads image size from nativeEvent.target on a web-shaped onLoad event, without throwing', async () => {
    const { getByTestId } = await renderWithProviders(
      <FocalPointPicker uri="file:///pic.jpg" x={0.5} y={0.5} onChange={() => {}} />
    )

    const image = getByTestId('focal-point-image')
    await expect(
      act(async () => {
        image.props.onLoad({
          nativeEvent: { target: { naturalWidth: 400, naturalHeight: 200 } },
        })
      })
    ).resolves.not.toThrow()
  })

  it('reads image size from nativeEvent.source on a native-shaped onLoad event', async () => {
    const { getByTestId } = await renderWithProviders(
      <FocalPointPicker uri="file:///pic.jpg" x={0.5} y={0.5} onChange={() => {}} />
    )

    const image = getByTestId('focal-point-image')
    await expect(
      act(async () => {
        image.props.onLoad({ nativeEvent: { source: { width: 400, height: 200 } } })
      })
    ).resolves.not.toThrow()
  })

  it('resets to the center focal point', async () => {
    const onChange = jest.fn()
    const { getByRole } = await renderWithProviders(
      <FocalPointPicker uri="file:///pic.jpg" x={0.1} y={0.9} onChange={onChange} />
    )

    fireEvent.press(getByRole('button', { name: /reset/i }))

    expect(onChange).toHaveBeenCalledWith(0.5, 0.5)
  })
})
