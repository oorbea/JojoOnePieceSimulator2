import tamaguiConfig from '../../../../tamagui.config'

// Regression guard for the bug this fixed: @tamagui/config/v4's defaultConfig
// ships `media` as max-width breakpoints, but every `$md`/`$lg`/`$xl` in this
// codebase was written assuming mobile-first min-width semantics — silently
// inverting every responsive rule in the app (desktop nav links rendering on
// mobile, the mobile dock rendering on desktop, etc). tamagui.config.ts now
// defines its own min-width scale; this test makes sure nobody reverts to
// (or re-introduces) a max-width tier under one of the plain `sm`/`md`/`lg`/
// `xl` keys without renaming it.
describe('tamagui media config', () => {
  const media = tamaguiConfig.media as Record<string, { minWidth?: number; maxWidth?: number }>
  const growingKeys = ['sm', 'md', 'lg', 'xl'] as const

  it('every non-"max*" key is a min-width breakpoint', () => {
    for (const [key, query] of Object.entries(media)) {
      if (key.startsWith('max')) continue
      expect(query.maxWidth).toBeUndefined()
      expect(typeof query.minWidth).toBe('number')
    }
  })

  it('growing tiers are declared in ascending order (widest-match-wins relies on this)', () => {
    const keys = Object.keys(media).filter((k) => growingKeys.includes(k as (typeof growingKeys)[number]))
    expect(keys).toEqual(growingKeys.slice(0, keys.length))
  })

  it('growing tiers have strictly increasing thresholds', () => {
    const widths = growingKeys.map((k) => media[k].minWidth!)
    for (let i = 1; i < widths.length; i++) {
      expect(widths[i]).toBeGreaterThan(widths[i - 1])
    }
  })

  it('the "mobile only" alias is a max-width query below the smallest growing tier', () => {
    expect(media.maxSm.maxWidth).toBeLessThan(media.sm.minWidth!)
  })
})
