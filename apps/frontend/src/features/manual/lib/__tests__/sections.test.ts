import { MANUAL_SECTIONS, parseSection } from '../sections'

describe('parseSection', () => {
  it('accepts every known section id', () => {
    for (const id of MANUAL_SECTIONS) expect(parseSection(id)).toBe(id)
  })

  it('takes the first value of a repeated query param', () => {
    expect(parseSection(['odds', 'how'])).toBe('odds')
  })

  it('ignores anything that is not a section instead of breaking the page', () => {
    expect(parseSection(undefined)).toBeNull()
    expect(parseSection('')).toBeNull()
    expect(parseSection('nope')).toBeNull()
    expect(parseSection(42)).toBeNull()
    expect(parseSection([])).toBeNull()
  })
})
