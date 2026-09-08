import { focusElement } from '@/shared/lib/a11y'

// Only the web branch is exercised here - this project runs under
// react-native-web (Platform.OS === 'web' is a given in the `logic`
// project), and there's no established pattern in this codebase for
// forcing Platform.OS to a native value from jsdom (see use-in-viewport.web
// .test.tsx's sibling native hook, which likewise has no jsdom-run
// counterpart). The native branch (AccessibilityInfo.setAccessibilityFocus
// via findNodeHandle) is straightforward enough to read for correctness;
// verifying it live is a manual on-device check, same as every other
// Platform.OS !== 'web' branch in this codebase.
describe('focusElement', () => {
  it('does nothing when el is null', () => {
    expect(() => focusElement(null)).not.toThrow()
  })

  it('calls .focus() on the element when it has one', () => {
    const focus = jest.fn()
    focusElement({ focus } as never)
    expect(focus).toHaveBeenCalledTimes(1)
  })

  it('does nothing when the element has no .focus method', () => {
    expect(() => focusElement({} as never)).not.toThrow()
  })
})
