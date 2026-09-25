import { act, render } from '@testing-library/react-native'
import { useEffect } from 'react'

import { useHoverTrigger } from '../tooltip'

// The web-only onFocus branch of useHoverTrigger - never engaged under the
// native jest project (Platform.OS !== 'web' there, triggerProps has no
// onFocus at all, only onLongPress/onPressOut). `.web.test.tsx` routes this
// to the jsdom "logic" project, same convention use-roving-group.web.test.tsx
// already set.
//
// Regression test for the bug this fix addresses: ConfirmSheet calls
// `.focus()` on its confirm GlossButton from an async `setTimeout` (pure
// keyboard-a11y aid, no user hover/Tab involved) and that used to pop the
// button's tooltip open with nobody anywhere near it. `onFocus` now only
// schedules `show` when the focused element matches `:focus-visible` -
// stubbed here via a fake event target rather than relying on jsdom's own
// (nonexistent) focus-visible heuristics, since jsdom never implements the
// real "was this a keyboard interaction" tracking browsers do.
let captured: ReturnType<typeof useHoverTrigger> | null = null

function Probe() {
  const value = useHoverTrigger()
  useEffect(() => {
    captured = value
  })
  return null
}

function api() {
  if (!captured) throw new Error('Probe has not rendered yet')
  return captured
}

function fakeTarget(focusVisible: boolean) {
  return { matches: (selector: string) => (selector === ':focus-visible' ? focusVisible : false) }
}

describe('useHoverTrigger (web onFocus branch)', () => {
  beforeEach(async () => {
    captured = null
    await render(<Probe />)
  })

  it('does not show on a focus event whose target is not :focus-visible', async () => {
    await act(async () => {
      ;(api().triggerProps as { onFocus?: (e: unknown) => void }).onFocus?.({
        target: fakeTarget(false),
      })
    })

    expect(api().visible).toBe(false)
  })

  it('shows on a focus event whose target IS :focus-visible', async () => {
    api().triggerRef.current = { measure: (cb) => cb(0, 0, 50, 20, 100, 200) }

    await act(async () => {
      ;(api().triggerProps as { onFocus?: (e: unknown) => void }).onFocus?.({
        target: fakeTarget(true),
      })
    })

    expect(api().visible).toBe(true)
  })

  it('still shows when the target has no matches() (falls back to showing)', async () => {
    api().triggerRef.current = { measure: (cb) => cb(0, 0, 50, 20, 100, 200) }

    await act(async () => {
      ;(api().triggerProps as { onFocus?: (e: unknown) => void }).onFocus?.({ target: {} })
    })

    expect(api().visible).toBe(true)
  })
})
