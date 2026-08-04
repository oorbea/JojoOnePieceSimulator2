import { Platform } from 'react-native'

import { a11yProps } from './a11y'

// Named `.native.test.ts` (not `__tests__/…`) so it's picked up only by the
// native jest project's generic *.test.ts matcher, not the logic project's
// dir-scoped one — it needs the real react-native Platform module (with a
// mutable OS getter) to exercise both branches directly, rather than
// react-native-web's fixed 'web' report under jsdom.
describe('a11yProps', () => {
  const originalOS = Platform.OS

  afterEach(() => {
    Object.defineProperty(Platform, 'OS', { get: () => originalOS })
  })

  it('maps to RN accessibility props on native', () => {
    Object.defineProperty(Platform, 'OS', { get: () => 'ios' })

    expect(a11yProps('Log out', 'button', { disabled: true })).toEqual({
      accessibilityLabel: 'Log out',
      accessibilityRole: 'button',
      accessibilityState: { disabled: true },
    })
  })

  it('maps to aria-* attributes on web, never RN accessibility* props', () => {
    Object.defineProperty(Platform, 'OS', { get: () => 'web' })

    const props = a11yProps('Log out', 'button', { disabled: true })

    expect(props).toEqual({
      'aria-label': 'Log out',
      role: 'button',
      'aria-disabled': true,
    })
    // The bug this guards: RN accessibility* props forwarded straight to a
    // DOM element render as unknown attributes and React logs a warning for
    // each one (see the a11y-web-leak project note) — none should appear.
    expect(props).not.toHaveProperty('accessibilityLabel')
    expect(props).not.toHaveProperty('accessibilityRole')
    expect(props).not.toHaveProperty('accessibilityState')
  })

  it('translates the "none" role to "presentation" on web', () => {
    Object.defineProperty(Platform, 'OS', { get: () => 'web' })

    expect(a11yProps('Stands, coming soon', 'none')).toEqual({
      'aria-label': 'Stands, coming soon',
      role: 'presentation',
    })
  })

  it('omits keys entirely when no label/role/state is given', () => {
    Object.defineProperty(Platform, 'OS', { get: () => 'ios' })
    expect(a11yProps()).toEqual({})

    Object.defineProperty(Platform, 'OS', { get: () => 'web' })
    expect(a11yProps()).toEqual({})
  })
})
